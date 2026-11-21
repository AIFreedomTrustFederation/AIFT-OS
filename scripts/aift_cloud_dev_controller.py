#!/usr/bin/env python3
"""Auditable federation status polling and narrowly scoped Go-test formatting.

No AI decisions, merges, deployments, credential writing or untrusted PR checkouts.
"""
from __future__ import annotations

import argparse
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import re
import subprocess
import sys
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.parse import quote
from urllib.request import Request, urlopen

OWNER = "AIFreedomTrustFederation"
DEFAULT_REPOS = "AIFT-OS,AIFT-Runtime,AIFT-Forge,VPS,www.aifreedomtrust.com"
REPO_PATTERN = re.compile(r"^[A-Za-z0-9_.-]+$")
MAX_REPOS = 12
MAX_PRS_PER_REPO = 8
MAX_CHECKS_PER_PR = 100
MAX_PR_HEADS_CHECKED = 4
API_BASE = "https://api.github.com"


class ApiError(RuntimeError):
    """Explicitly unverified API response; never treated as passing."""


def repo_names(config: str) -> list[str]:
    names = [x.strip() for x in config.split(",") if x.strip()]
    if not names or len(names) > MAX_REPOS or len(set(names)) != len(names):
        raise ValueError(f"Require 1..{MAX_REPOS} unique, explicit repository names")
    if any(not REPO_PATTERN.fullmatch(name) or name in (".", "..") for name in names):
        raise ValueError("Invalid repository allowlist entry")
    return names


class GitHubAPI:
    def __init__(self, token: str):
        self.token = token

    def get(self, path: str) -> Any:
        if not path.startswith("/repos/") or "//" in path or ".." in path:
            raise ValueError("Disallowed API path")
        headers = {
            "Accept": "application/vnd.github+json",
            "User-Agent": "aift-cloud-development-controller",
            "X-GitHub-Api-Version": "2022-11-28",
        }
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        request = Request(API_BASE + path, headers=headers)
        try:
            with urlopen(request, timeout=15) as response:
                return json.load(response)
        except HTTPError as error:
            raise ApiError(f"HTTP {error.code} for {path}") from error
        except (URLError, TimeoutError, ValueError) as error:
            raise ApiError(f"Could not verify {path}: {type(error).__name__}") from error


def reported_checks(runs: list[dict[str, Any]]) -> str:
    """Summarise only observed check runs; never call this merge readiness."""
    if not runs:
        return "no_checks_observed"
    conclusions = {run.get("conclusion") for run in runs}
    if conclusions & {"failure", "timed_out", "cancelled", "action_required"}:
        return "reported_checks_failed"
    if any(run.get("status") != "completed" for run in runs):
        return "checks_in_progress"
    if conclusions <= {"success", "neutral", "skipped"}:
        return "reported_checks_completed"
    return "checks_require_review"


def audit(api: GitHubAPI, repos: list[str], *, now: str) -> dict[str, Any]:
    observed: list[dict[str, Any]] = []
    for repo in repos:
        locator = f"/repos/{quote(OWNER)}/{quote(repo)}"
        entry: dict[str, Any] = {
            "repository": f"{OWNER}/{repo}",
            "url": f"https://github.com/{OWNER}/{repo}",
            "status": "unverified",
            "open_prs_sampled": 0,
            "pr_sample_limit": MAX_PRS_PER_REPO,
            "more_prs_possible": False,
            "pull_requests": [],
        }
        observed.append(entry)
        try:
            metadata = api.get(locator)
            if not isinstance(metadata, dict):
                raise ApiError("Malformed repository response")
            if metadata.get("full_name", "").casefold() != f"{OWNER}/{repo}".casefold():
                raise ApiError("Unexpected repository identity")
            if metadata.get("archived", False):
                entry["status"] = "archived"
                continue
            pulls = api.get(locator + f"/pulls?state=open&per_page={MAX_PRS_PER_REPO}")
            if not isinstance(pulls, list):
                raise ApiError("Malformed pull request response")
            entry["status"] = "observed"
            entry["open_prs_sampled"] = len(pulls)
            entry["more_prs_possible"] = len(pulls) == MAX_PRS_PER_REPO
            for i, pr in enumerate(pulls):
                number = pr.get("number")
                sha = (pr.get("head") or {}).get("sha")
                if not isinstance(number, int) or not isinstance(sha, str) or not re.fullmatch(r"[0-9a-f]{40}", sha):
                    entry["status"] = "partial"
                    continue
                detail: dict[str, Any] = {
                    "number": number,
                    "url": f"https://github.com/{OWNER}/{repo}/pull/{number}",
                    "head_sha": sha,
                    "draft": bool(pr.get("draft")),
                    "checks": "not_inspected_due_to_bound",
                    "checks_sampled": 0,
                    "required_checks_verified": False,
                    "merge_readiness_verified": False,
                }
                entry["pull_requests"].append(detail)
                if i >= MAX_PR_HEADS_CHECKED:
                    continue
                try:
                    check_data = api.get(locator + f"/commits/{sha}/check-runs?per_page={MAX_CHECKS_PER_PR}")
                    checks = check_data.get("check_runs")
                    if not isinstance(checks, list):
                        raise ApiError("Malformed checks response")
                    detail["checks_sampled"] = len(checks)
                    detail["checks"] = reported_checks(checks)
                    detail["more_checks_possible"] = check_data.get("total_count", 0) > len(checks)
                    if detail["more_checks_possible"]:
                        detail["checks"] = "incomplete_check_sample"
                except ApiError as error:
                    detail["checks"] = "unavailable"
                    detail["error"] = str(error)
        except ApiError as error:
            entry["status"] = "partial" if entry["open_prs_sampled"] else "unverified"
            entry["error"] = str(error)
    return {
        "schema_version": 1,
        "observed_at_utc": now,
        "scope": "explicit_allowlist_only",
        "notes": [
            "Checks are exact-head observations, not proof of branch protection, reviews, or mergeability.",
            "Open PR and check-run samples are bounded; missing data is not success.",
            "Cross-repository access can be unavailable with a repository-scoped GITHUB_TOKEN.",
        ],
        "repositories": observed,
    }


def markdown_report(report: dict[str, Any]) -> str:
    lines = [
        "## AIFT cloud development controller — observation only",
        "",
        f"Observed: {report['observed_at_utc']}",
        "",
        "| Repository | Observation | Open PRs sampled | Exact-head checks inspected |",
        "| --- | --- | ---: | ---: |",
    ]
    for repo in report["repositories"]:
        checked = sum(pr["checks"] != "not_inspected_due_to_bound" for pr in repo["pull_requests"])
        lines.append(f"| {repo['repository']} | {repo['status']} | {repo['open_prs_sampled']} | {checked} |")
    lines += ["", "This report does not establish merge readiness or authorize a change.",
              "Unavailable evidence is recorded as unavailable; the audit is never a green merge gate.", ""]
    return "\n".join(lines)


def tracked_go_test_files() -> list[str]:
    output = subprocess.check_output(["git", "ls-files", "-z", "--", "*.go"])
    paths = [os.fsdecode(p) for p in output.split(b"\0") if p]
    return sorted(path for path in paths if path.startswith("internal/")
                  and path.endswith("_test.go") and Path(path).is_file()
                  and not Path(path).is_symlink())


def format_tests() -> int:
    if subprocess.check_output(["git", "status", "--porcelain"]):
        raise RuntimeError("Refusing to modify a dirty checkout")
    paths = tracked_go_test_files()
    if not paths:
        print("No tracked internal Go test files; no changes")
        return 0
    subprocess.run(["gofmt", "-w", *paths], check=True)
    changed_raw = subprocess.check_output(["git", "diff", "--name-only", "-z"])
    changed = [os.fsdecode(p) for p in changed_raw.split(b"\0") if p]
    unexpected = sorted(set(changed) - set(paths))
    if unexpected:
        raise RuntimeError(f"Formatting affected forbidden paths: {unexpected}")
    print(f"Go test formatting: {len(changed)} changed, {len(paths)} checked")
    return len(changed)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    audit_cmd = sub.add_parser("audit")
    audit_cmd.add_argument("--output", required=True, type=Path)
    audit_cmd.add_argument("--summary", type=Path)
    sub.add_parser("format-tests")
    args = parser.parse_args()
    if args.command == "format-tests":
        format_tests()
        return 0
    names = repo_names(os.environ.get("AIFT_AUTODEV_REPOS", DEFAULT_REPOS))
    report = audit(GitHubAPI(os.environ.get("GITHUB_TOKEN", "")), names,
                   now=datetime.now(timezone.utc).isoformat())
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    summary = markdown_report(report)
    if args.summary:
        with args.summary.open("a", encoding="utf-8") as file:
            file.write(summary)
    print(summary)
    if not any(x["status"] in ("observed", "archived", "partial") for x in report["repositories"]):
        print("No repository metadata was verified", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (ValueError, RuntimeError, subprocess.CalledProcessError) as exc:
        print(f"AIFT cloud controller blocked: {exc}", file=sys.stderr)
        sys.exit(1)
