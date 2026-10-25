#!/usr/bin/env python3
"""Collect verified default-branch GitHub Actions evidence via authenticated gh CLI.

This verifies GitHub run/artifact metadata and then delegates structural validation
of unsigned aift.verification.v1 records to check_federation_evidence.py.
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path

import check_federation_evidence as consumer

REPO = re.compile(r"AIFreedomTrustFederation/[A-Za-z0-9_.-]+\Z")
SHA = re.compile(r"[0-9a-f]{40}\Z")
ARTIFACTS = {
    "AIFreedomTrustFederation/.github": ("aift-catalog-verification", "aift-catalog-verification.json"),
    "AIFreedomTrustFederation/AIFT-Runtime": ("aift-runtime-verification", "aift-runtime-verification.json"),
}
# Restrict provenance to the designated producer workflows, not any successful push run.
WORKFLOWS = {
    "AIFreedomTrustFederation/.github": (362136764, ".github/workflows/catalog-integrity.yml"),
    "AIFreedomTrustFederation/AIFT-Runtime": (361044013, ".github/workflows/qa.yml"),
}


class CollectionError(ValueError):
    """GitHub evidence could not be safely corroborated."""


def gh(*args: str) -> str:
    try:
        result = subprocess.run(["gh", *args], capture_output=True, text=True, check=False, timeout=60)
    except (OSError, subprocess.TimeoutExpired) as exc:
        raise CollectionError(f"gh invocation failed: {exc}") from exc
    if result.returncode:
        raise CollectionError(f"gh command failed (exit {result.returncode}): {' '.join(args[:3])}")
    return result.stdout


def api(path: str) -> dict:
    try:
        value = json.loads(gh("api", path), object_pairs_hook=consumer.no_duplicate_keys,
                           parse_constant=consumer.reject_constant)
    except (ValueError, consumer.EvidenceError) as exc:
        raise CollectionError(f"invalid GitHub API response for {path}: {exc}") from exc
    if not isinstance(value, dict):
        raise CollectionError(f"invalid GitHub API object for {path}")
    return value


def parse_pairs(values: list[str], kind: str) -> dict[str, str]:
    result = {}
    for item in values:
        repo, sep, value = item.partition("=")
        if (not sep or not REPO.fullmatch(repo) or repo in result or
                not (SHA.fullmatch(value) if kind == "commit" else value.isascii() and value.isdecimal() and int(value) > 0)):
            raise CollectionError(f"invalid or duplicate {kind}: {item}")
        result[repo] = value
    return result


def collect(runs: dict[str, str], expected: dict[str, str], output: Path) -> dict:
    if not runs or set(runs) != set(expected):
        raise CollectionError("run and expected-commit repository sets must match")
    if any(repo not in ARTIFACTS for repo in runs):
        raise CollectionError("unsupported repository artifact contract")
    output = output.resolve()
    if output.exists():
        raise CollectionError("output directory already exists; refusing overwrite")
    # Validate all GitHub metadata and record contents before publishing results.
    with tempfile.TemporaryDirectory(prefix="aift-evidence-") as temp:
        root = Path(temp)
        records = []
        provenance = []
        for index, (repo, run_id) in enumerate(sorted(runs.items())):
            sha = expected[repo]
            info = api(f"repos/{repo}")
            branch = info.get("default_branch")
            if not isinstance(branch, str) or not branch or "/" in branch:
                raise CollectionError(f"invalid default branch for {repo}")
            branch_info = api(f"repos/{repo}/branches/{branch}")
            if not isinstance(branch_info.get("commit"), dict) or branch_info["commit"].get("sha") != sha:
                raise CollectionError(f"expected commit is not current {repo}/{branch}")
            run = api(f"repos/{repo}/actions/runs/{run_id}")
            workflow_id, workflow_path = WORKFLOWS[repo]
            if (type(run.get("workflow_id")) is not int or run["workflow_id"] != workflow_id or
                    run.get("path") != workflow_path):
                raise CollectionError(f"run workflow identity mismatch: {repo}")
            if (not isinstance(run.get("repository"), dict) or run["repository"].get("full_name") != repo or
                    run.get("head_sha") != sha or run.get("head_branch") != branch or
                    run.get("event") != "push" or run.get("status") != "completed" or
                    run.get("conclusion") != "success"):
                raise CollectionError(f"run metadata does not establish successful default-branch CI: {repo}")
            listing = api(f"repos/{repo}/actions/runs/{run_id}/artifacts?per_page=100")
            artifact_name, filename = ARTIFACTS[repo]
            artifacts = listing.get("artifacts")
            if not isinstance(artifacts, list) or listing.get("total_count") != len(artifacts):
                raise CollectionError(f"artifact list incomplete for {repo}")
            matches = [a for a in artifacts if isinstance(a, dict) and a.get("name") == artifact_name]
            if len(matches) != 1 or matches[0].get("expired") is not False:
                raise CollectionError(f"missing, duplicate, or expired verification artifact: {repo}")
            artifact = matches[0]
            if type(artifact.get("id")) is not int or artifact["id"] <= 0:
                raise CollectionError(f"invalid artifact identity: {repo}")
            directory = root / f"repo-{index}"
            directory.mkdir()
            gh("run", "download", run_id, "--repo", repo, "--name", artifact_name, "--dir", str(directory))
            files = list(directory.rglob("*"))
            if len(files) != 1 or files[0].name != filename or not files[0].is_file() or files[0].is_symlink():
                raise CollectionError(f"artifact must contain exactly {filename}: {repo}")
            record = consumer.read_record(files[0])
            consumer.validate(record, expected)
            if (record.get("source", {}).get("type") != "github-actions" or
                    record["source"].get("run_url") != f"https://github.com/{repo}/actions/runs/{run_id}"):
                raise CollectionError(f"artifact source does not match authenticated run: {repo}")
            records.append(files[0])
            provenance.append({"repository": repo, "run_id": int(run_id), "artifact_id": artifact["id"],
                               "commit_sha": sha, "default_branch": branch})
        assessment, code = consumer.inspect(records, expected)
        if code:
            raise CollectionError("retrieved evidence failed structural verification")
        # A branch can advance while artifacts are downloaded. Recheck every
        # producer immediately before publishing a current-commit assessment.
        for entry in provenance:
            repo, branch, sha = entry["repository"], entry["default_branch"], entry["commit_sha"]
            latest_repo = api(f"repos/{repo}")
            if latest_repo.get("default_branch") != branch:
                raise CollectionError(f"default branch changed during collection: {repo}")
            latest_branch = api(f"repos/{repo}/branches/{branch}")
            if (not isinstance(latest_branch.get("commit"), dict) or
                    latest_branch["commit"].get("sha") != sha):
                raise CollectionError(f"default branch advanced during collection: {repo}")
        assessment["provenance"] = "github_api_run_and_artifact_checked_unsigned"
        assessment["github_runs"] = provenance
        output.mkdir(parents=True, exist_ok=False)
        for source, entry in zip(records, provenance):
            target = output / ARTIFACTS[entry["repository"]][1]
            target.write_bytes(source.read_bytes())
        (output / "federation-assessment.json").write_text(
            json.dumps(assessment, indent=2, sort_keys=True) + "\n", encoding="utf-8")
        return assessment


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--run", action="append", required=True, metavar="OWNER/REPO=RUN_ID")
    parser.add_argument("--expect", action="append", required=True, metavar="OWNER/REPO=SHA")
    parser.add_argument("--output-dir", required=True, type=Path)
    args = parser.parse_args(argv)
    try:
        assessment = collect(parse_pairs(args.run, "run"), parse_pairs(args.expect, "commit"), args.output_dir)
        print(json.dumps({"status": assessment["status"], "provenance": assessment["provenance"],
                          "output_dir": str(args.output_dir)}, sort_keys=True))
        return 0
    except (CollectionError, consumer.EvidenceError, OSError) as exc:
        print(f"Federation evidence collection failed: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
