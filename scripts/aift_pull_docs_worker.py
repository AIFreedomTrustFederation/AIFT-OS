#!/usr/bin/env python3
"""Opt-in credentialless Aetherion issue poller with durable local-only outbox.

This process NEVER authenticates to GitHub, installs a runner, executes model
tool calls, invokes gh/git push, or publishes PRs. Only a separate operator
can hand a validated plan to the GitHub-hosted publisher workflow.
"""
from __future__ import annotations

import argparse
import base64
from contextlib import contextmanager
import fcntl
import json
import os
from pathlib import Path
import re
import stat
import sys
import tempfile
import time
from urllib.error import HTTPError, URLError
from urllib.request import HTTPRedirectHandler, Request, build_opener

try:
    from scripts import aift_secure_cloud_worker as worker
except ModuleNotFoundError:  # Direct execution: python3 scripts/aift_pull_docs_worker.py
    import aift_secure_cloud_worker as worker

MAX_PENDING = 8
MAX_PLAN_BYTES = 8192
MIN_INTERVAL = 600
DEFAULT_STATE = Path.home() / ".local" / "state" / "aift-aetherion-pull"


class NoRedirects(HTTPRedirectHandler):
    def redirect_request(self, *_args, **_kwargs):
        return None


def public_get(path: str, _unused_token: str = ""):
    """Anonymous, bounded read-only GitHub API access, no auth header or redirects."""
    base = f"/repos/{worker.OWNER}/{worker.REPO}"
    allowed = (
        f"{base}/branches/main",
        f"{base}/issues?state=open&labels={worker.LABEL}&per_page={worker.MAX_ISSUES}&sort=created&direction=asc",
        f"{base}/pulls?state=open&per_page={worker.MAX_OPEN_PRS}",
    )
    if path not in allowed:
        raise worker.Blocked("Anonymous API path not allowed")
    request = Request(
        "https://api.github.com" + path,
        headers={
            "Accept": "application/vnd.github+json",
            "User-Agent": "aift-anonymous-docs-poller",
            "X-GitHub-Api-Version": "2022-11-28",
        },
        method="GET",
    )
    try:
        with build_opener(NoRedirects).open(request, timeout=15) as response:
            body = response.read(262145)
            if len(body) > 262144:
                raise worker.Blocked("Public GitHub response exceeds limit")
        return json.loads(body)
    except HTTPError as error:
        raise worker.Blocked(f"Public GitHub read unavailable (HTTP {error.code})") from error
    except (URLError, TimeoutError, ValueError, OSError) as error:
        raise worker.Blocked("Public GitHub read unavailable") from error


def exact_public_main() -> str:
    result = public_get(f"/repos/{worker.OWNER}/{worker.REPO}/branches/main")
    try:
        sha = result["commit"]["sha"]
    except (KeyError, TypeError) as error:
        raise worker.Blocked("GitHub main revision missing") from error
    if not isinstance(sha, str) or not worker.SHA_RE.fullmatch(sha):
        raise worker.Blocked("GitHub main revision invalid")
    return sha


def _private_dir(path: Path) -> Path:
    if path.is_symlink():
        raise worker.Blocked("State directory cannot be a symlink")
    if path.exists() and not path.is_dir():
        raise worker.Blocked("State path is not a directory")
    path.mkdir(parents=True, exist_ok=True, mode=0o700)
    if path.is_symlink() or stat.S_IMODE(path.stat().st_mode) & 0o077:
        raise worker.Blocked("State directory must be private (chmod 700)")
    return path


def prepare_state(root: Path) -> Path:
    """State must be outside repository, owner-only, and never a symlink."""
    root = root.expanduser().absolute()
    repo = Path.cwd().resolve()
    if root == repo or repo in root.parents:
        raise worker.Blocked("Outbox must be outside the Git checkout")
    _private_dir(root)
    _private_dir(root / "outbox")
    _private_dir(root / "archive")
    return root


@contextmanager
def exclusive_lock(root: Path):
    fd = os.open(root / ".poll.lock", os.O_WRONLY | os.O_CREAT | os.O_NOFOLLOW, 0o600)
    try:
        if stat.S_IMODE(os.fstat(fd).st_mode) & 0o077:
            raise worker.Blocked("Poll lock permissions unsafe")
        try:
            fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError as error:
            raise worker.Blocked("Another local worker already owns the lock") from error
        yield
    finally:
        os.close(fd)


def _plan_path(root: Path, issue: int, *, archived: bool = False) -> Path:
    if type(issue) is not int or issue <= 0:
        raise worker.Blocked("Issue number invalid")
    return root / ("archive" if archived else "outbox") / f"issue-{issue}.json"


def _plans(folder: Path) -> list[Path]:
    plans = sorted(folder.glob("issue-*.json"))
    if any(path.is_symlink() or not path.is_file() for path in plans):
        raise worker.Blocked("Unsafe outbox entry")
    return plans


def _read_plan(path: Path) -> bytes:
    if path.is_symlink() or not path.is_file() or not 0 < path.stat().st_size <= MAX_PLAN_BYTES:
        raise worker.Blocked("Missing or oversized plan")
    if stat.S_IMODE(path.stat().st_mode) & 0o077:
        raise worker.Blocked("Plan permissions unsafe")
    data = path.read_bytes()
    try:
        record = json.loads(data)
    except (UnicodeError, ValueError) as error:
        raise worker.Blocked("Invalid plan JSON") from error
    if not isinstance(record, dict) or set(record) != worker.PLAN_KEYS:
        raise worker.Blocked("Invalid v2 plan fields")
    if record["schema_version"] != worker.PLAN_SCHEMA_VERSION:
        raise worker.Blocked("Unsupported plan version")
    return data


def poll_once(root: Path, *, model: str, endpoint: str) -> dict:
    """One bounded poll, at most one model edit, no GitHub credentials or writes."""
    worker.validate_model_config(endpoint, model)
    root = prepare_state(root)
    with exclusive_lock(root):
        if len(_plans(root / "outbox")) >= MAX_PENDING:
            return {"status": "queue_full", "max_pending": MAX_PENDING}
        if worker.run("git", "branch", "--show-current") != "main":
            raise worker.Blocked("Local checkout must be on main")
        sha = exact_public_main()
        worker.verify_checkout(sha)
        selection = worker.select_issue("", fetcher=public_get)
        if selection is None:
            return {"status": "idle", "head_sha": sha}
        issue, _target, _instruction = selection
        issue_number = issue["number"]
        final = _plan_path(root, issue_number)
        archived = _plan_path(root, issue_number, archived=True)
        if final.exists() or archived.exists():
            return {"status": "already_seen", "issue": issue_number}
        # A crash before os.replace may leave a hidden temporary file but cannot
        # produce a visible partial queue item or duplicate an existing issue.
        fd, temporary = tempfile.mkstemp(prefix=".pending-", suffix=".json", dir=root / "outbox")
        os.close(fd)
        temp_path = Path(temporary)
        try:
            worker.create_plan(
                temp_path, read_token="", model=model, endpoint=endpoint,
                expected_sha=sha, selection=selection,
            )
            data = _read_plan(temp_path)
            record = json.loads(data)
            if record.get("head_sha") != sha or record.get("issue_number") != issue_number:
                raise worker.Blocked("Local plan provenance mismatch")
            # Refuse to replace an existing queue item; lock prevents two pollers.
            if final.exists() or archived.exists():
                raise worker.Blocked("Duplicate local plan")
            os.replace(temp_path, final)
            return {"status": "queued", "issue": issue_number, "head_sha": sha, "pending": len(_plans(root / "outbox"))}
        finally:
            temp_path.unlink(missing_ok=True)


def queue_status(root: Path) -> dict:
    root = prepare_state(root)
    with exclusive_lock(root):
        return {
            "status": "read_only",
            "pending": [p.name for p in _plans(root / "outbox")],
            "archived": len(_plans(root / "archive")),
            "max_pending": MAX_PENDING,
            "publisher": "not_checked",
        }


def export_plan(root: Path, issue: int) -> str:
    root = prepare_state(root)
    with exclusive_lock(root):
        data = _read_plan(_plan_path(root, issue))
        record = json.loads(data)
        if record.get("issue_number") != issue:
            raise worker.Blocked("Plan issue mismatch")
        return base64.b64encode(data).decode("ascii")


def mark_delivered(root: Path, issue: int) -> None:
    """Operator-only local acknowledgement; not proof of publication."""
    root = prepare_state(root)
    with exclusive_lock(root):
        source = _plan_path(root, issue)
        _read_plan(source)
        destination = _plan_path(root, issue, archived=True)
        if destination.exists():
            raise worker.Blocked("Issue already archived")
        os.rename(source, destination)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--state-dir", type=Path, default=DEFAULT_STATE)
    commands = parser.add_subparsers(dest="command", required=True)
    for name in ("once", "loop"):
        p = commands.add_parser(name)
        p.add_argument("--model", required=True)
        p.add_argument("--model-url", required=True)
        if name == "loop":
            p.add_argument("--interval", type=int, default=MIN_INTERVAL)
    commands.add_parser("status")
    export = commands.add_parser("export")
    export.add_argument("--issue", type=int, required=True)
    acknowledged = commands.add_parser("mark-delivered")
    acknowledged.add_argument("--issue", type=int, required=True)
    args = parser.parse_args(argv)
    try:
        if args.command == "status":
            print(json.dumps(queue_status(args.state_dir), sort_keys=True))
        elif args.command == "export":
            print(export_plan(args.state_dir, args.issue))
        elif args.command == "mark-delivered":
            mark_delivered(args.state_dir, args.issue)
            print(json.dumps({"status": "locally_archived", "issue": args.issue}))
        elif args.command == "once":
            print(json.dumps(poll_once(args.state_dir, model=args.model, endpoint=args.model_url), sort_keys=True))
        else:
            if not MIN_INTERVAL <= args.interval <= 3600:
                parser.error("--interval must be between 600 and 3600 seconds")
            failures = 0
            while True:
                try:
                    report = poll_once(args.state_dir, model=args.model, endpoint=args.model_url)
                    print(json.dumps(report, sort_keys=True), flush=True)
                    failures = 0
                except (worker.Blocked, OSError, ValueError) as error:
                    failures += 1
                    print(json.dumps({"status": "blocked", "reason": str(error)[:140]}), flush=True)
                time.sleep(min(3600, args.interval * (2 ** min(failures, 4))))
    except (worker.Blocked, OSError, ValueError) as error:
        print(json.dumps({"status": "blocked", "reason": str(error)[:140]}), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
