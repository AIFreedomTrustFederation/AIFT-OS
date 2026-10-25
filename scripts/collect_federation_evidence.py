#!/usr/bin/env python3
"""Collect verified default-branch GitHub Actions evidence via authenticated gh CLI.

This verifies GitHub run/artifact metadata and then delegates structural validation
of unsigned aift.verification.v1 records to check_federation_evidence.py.
"""

from __future__ import annotations

import argparse
import hashlib
import io
import json
import zipfile
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



def download_record(repo: str, artifact_id: int, filename: str, digest: str, directory: Path) -> Path:
    """Fetch the *listed artifact ID*, not a potentially different same-name archive."""
    endpoint = f"repos/{repo}/actions/artifacts/{artifact_id}/zip"
    try:
        result = subprocess.run(["gh", "api", endpoint], capture_output=True, check=False, timeout=60)
    except (OSError, subprocess.TimeoutExpired) as exc:
        raise CollectionError(f"artifact download failed for {repo}: {exc}") from exc
    if result.returncode:
        raise CollectionError(f"artifact ID download failed for {repo} (exit {result.returncode})")
    # GitHub publishes a SHA-256 digest for Actions artifacts. Bind the
    # downloaded ZIP bytes to the exact artifact metadata we inspected.
    if not isinstance(digest, str) or not re.fullmatch(r"sha256:[0-9a-f]{64}", digest):
        raise CollectionError(f"missing or invalid artifact SHA-256 digest: {repo}")
    if hashlib.sha256(result.stdout).hexdigest() != digest.removeprefix("sha256:"):
        raise CollectionError(f"artifact SHA-256 digest mismatch: {repo}")
    # These verification JSON records should be small. Bound archive and
    # expanded member sizes; never extract arbitrary paths or links.
    if not result.stdout or len(result.stdout) > 4 * 1024 * 1024:
        raise CollectionError(f"artifact archive missing or too large: {repo}")
    try:
        with zipfile.ZipFile(io.BytesIO(result.stdout)) as archive:
            members = archive.infolist()
            if len(members) != 1:
                raise CollectionError(f"artifact must contain exactly one file: {repo}")
            member = members[0]
            kind = (member.external_attr >> 16) & 0o170000
            if (member.filename != filename or member.is_dir() or
                    kind not in (0, 0o100000) or member.file_size > 1024 * 1024):
                raise CollectionError(f"artifact contains an unsafe or unexpected file: {repo}")
            payload = archive.read(member)
    except (zipfile.BadZipFile, zipfile.LargeZipFile, RuntimeError, EOFError, OSError,
            NotImplementedError) as exc:
        raise CollectionError(f"invalid artifact ZIP for {repo}: {exc}") from exc
    target = directory / filename
    target.write_bytes(payload)
    return target

def parse_pairs(values: list[str], kind: str) -> dict[str, str]:
    result = {}
    for item in values:
        repo, sep, value = item.partition("=")
        if (not sep or not REPO.fullmatch(repo) or repo in result or
                not (SHA.fullmatch(value) if kind == "commit" else value.isascii() and value.isdecimal() and int(value) > 0)):
            raise CollectionError(f"invalid or duplicate {kind}: {item}")
        result[repo] = value
    return result


def publish_records(output: Path, records: list[Path], provenance: list[dict], assessment: dict) -> None:
    """Publish only a complete assessment; discard staged files on failure."""
    output.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix=".aift-evidence-stage-", dir=output.parent) as temp:
        stage = Path(temp) / "complete"
        stage.mkdir()
        for source, entry in zip(records, provenance):
            target = stage / ARTIFACTS[entry["repository"]][1]
            target.write_bytes(source.read_bytes())
        (stage / "federation-assessment.json").write_text(
            json.dumps(assessment, indent=2, sort_keys=True) + "\n", encoding="utf-8")
        # The final rename makes the complete directory visible in one step.
        # Never intentionally replace a previously published assessment.
        if output.exists():
            raise CollectionError("output directory already exists; refusing overwrite")
        stage.rename(output)


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
            downloaded = download_record(repo, artifact["id"], filename, artifact.get("digest"), directory)
            record = consumer.read_record(downloaded)
            consumer.validate(record, expected)
            if (record.get("source", {}).get("type") != "github-actions" or
                    record["source"].get("run_url") != f"https://github.com/{repo}/actions/runs/{run_id}"):
                raise CollectionError(f"artifact source does not match authenticated run: {repo}")
            records.append(downloaded)
            provenance.append({"repository": repo, "run_id": int(run_id), "artifact_id": artifact["id"],
                               "artifact_digest": artifact["digest"], "commit_sha": sha,
                               "default_branch": branch})
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
        # A caller may intentionally collect a subset of supported producers.
        # Preserve that CLI behavior, but never label subset evidence as a
        # complete reported pass for the supported Federation baseline.
        missing_producers = sorted(set(ARTIFACTS) - set(runs))
        assessment["producer_coverage"] = {
            "collected": sorted(runs),
            "not_collected": missing_producers,
        }
        if missing_producers and assessment["status"] == "reported_pass":
            assessment["status"] = "partial"
        assessment["provenance"] = "github_api_run_and_artifact_checked_unsigned"
        assessment["github_runs"] = provenance
        publish_records(output, records, provenance, assessment)
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
