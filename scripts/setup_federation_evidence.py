#!/usr/bin/env python3
"""Configure AIFT-OS evidence credentials locally; preview unless --apply."""
import argparse
import json
import os
from pathlib import Path
import re
import stat
import subprocess
import sys

REPO = "AIFreedomTrustFederation/AIFT-OS"
WORKFLOW = ".github/workflows/federation-evidence-collector.yml"
VARIABLE = "AIFT_FEDERATION_EVIDENCE_APP_CLIENT_ID"
SECRET = "AIFT_FEDERATION_EVIDENCE_APP_PRIVATE_KEY"

class SetupError(ValueError):
    pass

def gh(*args, input=None, timeout=60):
    # Never relay CLI stderr: a credential-handling child could echo its input.
    env = dict(os.environ, GH_HOST="github.com", GH_PROMPT_DISABLED="1")
    try:
        result = subprocess.run(["gh", *args], input=input, capture_output=True,
                                timeout=timeout, env=env)
    except (OSError, subprocess.TimeoutExpired) as exc:
        raise SetupError("GitHub CLI unavailable or timed out") from exc
    if result.returncode:
        raise SetupError("GitHub CLI operation failed; check authentication and administrator permissions")
    return result.stdout

def api(path):
    try:
        return json.loads(gh("api", "--hostname", "github.com", path))
    except (ValueError, UnicodeError) as exc:
        raise SetupError("Invalid GitHub API response") from exc

def read_key(path):
    flags = os.O_RDONLY | getattr(os, "O_NOFOLLOW", 0) | getattr(os, "O_NONBLOCK", 0)
    fd = os.open(path, flags)
    try:
        info = os.fstat(fd)
        if not stat.S_ISREG(info.st_mode):
            raise SetupError("Private key must be a regular file")
        if os.name == "posix" and info.st_mode & 0o077:
            raise SetupError("Private key permissions must exclude group/other access (chmod 600)")
        with os.fdopen(fd, "rb", closefd=False) as handle:
            key = handle.read(65537)
    finally:
        os.close(fd)
    if len(key) > 65536 or not (
        key.startswith(b"-----BEGIN RSA PRIVATE KEY-----") or
        key.startswith(b"-----BEGIN PRIVATE KEY-----")
    ):
        raise SetupError("Expected a PEM private key no larger than 64 KiB")
    return key

def validate_run(run_id):
    branch = api(f"repos/{REPO}/branches/main")
    run = api(f"repos/{REPO}/actions/runs/{run_id}")
    if (run.get("id") != run_id or run.get("event") != "push" or
        run.get("head_branch") != "main" or run.get("head_sha") != branch["commit"]["sha"] or
        run.get("path") != WORKFLOW or run.get("status") != "completed" or
        run.get("conclusion") != "failure" or
        run.get("repository", {}).get("full_name") != REPO or
        run.get("head_repository", {}).get("full_name") != REPO):
        raise SetupError("Rerun must be the failed collector push run for current main")
    return run

def configure(args):
    if args.client_id and not re.fullmatch(r"[A-Za-z0-9_]{1,128}", args.client_id):
        raise SetupError("Invalid App client ID")
    if args.apply and (not args.client_id or not args.private_key):
        raise SetupError("--apply requires --client-id and --private-key")
    if args.rerun and (not args.apply or not 1 <= args.rerun < 10**20):
        raise SetupError("--rerun requires --apply and a positive run ID")
    info = api(f"repos/{REPO}")
    if info.get("full_name") != REPO or info.get("permissions", {}).get("admin") is not True:
        raise SetupError("An administrator-authenticated GitHub CLI session is required")
    print(f"Target: {REPO} on github.com")
    print(f"Set Actions variable: {VARIABLE}; set Actions secret: {SECRET}")
    print("App must already be installed on .github and AIFT-Runtime with Contents/Actions read.")
    if not args.apply:
        print("Preview only; no key read, settings changed, or workflows rerun.")
        return
    if args.rerun:
        validate_run(args.rerun)
    # Validate/read before any remote mutation. Keys go through stdin, never argv.
    key = read_key(args.private_key)
    gh("secret", "set", SECRET, "--repo", REPO, input=key)
    del key
    gh("variable", "set", VARIABLE, "--repo", REPO, "--body", args.client_id)
    print("Settings written. Secret value cannot be read back; live validation is still required.")
    if args.rerun:
        validate_run(args.rerun)  # Fail closed if main advanced during setup.
        gh("run", "rerun", str(args.rerun), "--failed", "--repo", REPO)
        print(f"Rerun requested: https://github.com/{REPO}/actions/runs/{args.rerun}")
        print(f"Watch: gh run watch {args.rerun} --repo {REPO} --exit-status")
    print("If a later operation failed, earlier writes may persist; rerun safely after repair.")

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--client-id")
    parser.add_argument("--private-key", type=Path)
    parser.add_argument("--apply", action="store_true", help="Explicitly authorize credential writes")
    parser.add_argument("--rerun", type=int, metavar="FAILED_MAIN_RUN_ID")
    args = parser.parse_args()
    try:
        configure(args)
    except (SetupError, OSError, KeyError) as exc:
        # Do not print paths, key bytes, child output, or raw API responses.
        print("Setup failed. No success is claimed; earlier successful writes may persist.", file=sys.stderr)
        if isinstance(exc, SetupError):
            print(str(exc), file=sys.stderr)
        return 1
    return 0

if __name__ == "__main__":
    sys.exit(main())
