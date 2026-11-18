#!/usr/bin/env python3
"""Offline source inventory; never execute manifest commands or certify readiness."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess

FILES = ("README.md", "AGENTS.md", "aift.repo.json", "package.json",
         "pnpm-lock.yaml", "package-lock.json", "go.mod", "go.sum",
         "pyproject.toml", "requirements.txt", "docs/status.md", "STATUS.md")


def inspect(root):
    root = Path(root).resolve()
    result = {"repository": root.name, "verification": "not_run", "files": {},
              "limitations": ["Source presence is not capability or CI evidence.",
                              "No network access or repository commands executed."]}
    if not root.is_dir():
        result["error"] = "missing_directory"
        return result
    for name in FILES:
        path = root / name
        if path.is_symlink() or not path.resolve().is_relative_to(root):
            result["files"][name] = {"state": "refused_symlink"}
        elif not path.is_file():
            result["files"][name] = {"state": "missing"}
        else:
            data = path.read_bytes()
            entry = {"state": "present", "sha256": hashlib.sha256(data).hexdigest()}
            if name in ("aift.repo.json", "package.json"):
                try:
                    value = json.loads(data)
                    if not isinstance(value, dict):
                        raise ValueError("object required")
                    # Only field names: manifest commands remain untrusted source data.
                    entry["keys"] = sorted(value)
                except (ValueError, UnicodeError):
                    entry["state"] = "invalid_json_object"
            result["files"][name] = entry
    try:
        actual = subprocess.check_output(
            ["git", "-C", str(root), "rev-parse", "--show-toplevel"],
            stderr=subprocess.DEVNULL, text=True, timeout=10).strip()
        if Path(actual).resolve() != root:
            raise ValueError("not repository root")
        result["head"] = subprocess.check_output(
            ["git", "-C", str(root), "rev-parse", "HEAD"],
            stderr=subprocess.DEVNULL, text=True, timeout=10).strip()
        dirty = subprocess.check_output(
            ["git", "--no-optional-locks", "-C", str(root), "status", "--porcelain",
             "--untracked-files=normal"], stderr=subprocess.DEVNULL, text=True,
            timeout=10)
        result["worktree"] = "dirty" if dirty else "clean"
    except (OSError, ValueError, subprocess.SubprocessError):
        result["git_evidence"] = "unavailable"
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("repositories", nargs="+")
    args = parser.parse_args()
    try:
        records = [inspect(path) for path in args.repositories]
    except OSError as error:
        parser.exit(2, "inventory read failed: " + str(error) + "\n")
    print(json.dumps({"schema_version": 1, "scope": "offline_source_inventory",
                      "repositories": records}, indent=2, sort_keys=True))
    return 1 if any("error" in r or "git_evidence" in r or any(
        f["state"] in ("invalid_json_object", "refused_symlink")
        for f in r["files"].values()) for r in records) else 0


if __name__ == "__main__":
    raise SystemExit(main())
