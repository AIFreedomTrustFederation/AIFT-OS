#!/usr/bin/env python3
"""Collect verified default-branch GitHub Actions evidence via authenticated gh CLI.

This verifies GitHub run/artifact metadata and then delegates structural validation
of unsigned aift.verification.v1 records to check_federation_evidence.py.
"""

from __future__ import annotations

import argparse
from datetime import datetime, timedelta, timezone
import hashlib
import io
import json
import math
import os
import selectors
import zipfile
import zlib
import re
import subprocess
import sys
import tempfile
import time
from pathlib import Path

import check_federation_evidence as consumer

REPO = re.compile(r"AIFreedomTrustFederation/[A-Za-z0-9_.-]+\Z")
SHA = re.compile(r"[0-9a-f]{40}\Z")
MAX_ARCHIVE_BYTES = 4 * 1024 * 1024
MAX_METADATA_BYTES = 1024 * 1024
ARTIFACTS = {
    "AIFreedomTrustFederation/.github": ("aift-catalog-verification", "aift-catalog-verification.json"),
    "AIFreedomTrustFederation/AIFT-Runtime": ("aift-runtime-verification", "aift-runtime-verification.json"),
}
# Restrict provenance to the designated producer workflows, not any successful push run.
WORKFLOWS = {
    "AIFreedomTrustFederation/.github": (362136764, ".github/workflows/catalog-integrity.yml"),
    "AIFreedomTrustFederation/AIFT-Runtime": (361044013, ".github/workflows/qa.yml"),
}

# Verified organization repositories that do not yet publish collector-compatible
# evidence. This is a known, non-exhaustive integration backlog, not a roster.
KNOWN_UNINTEGRATED = (
    "AIFreedomTrustFederation/AIFT-Genesis",
    "AIFreedomTrustFederation/AIFT-Forge",
    "AIFreedomTrustFederation/TheMindofAll",
    "AIFreedomTrustFederation/VPS",
    "AIFreedomTrustFederation/AI-Freedom-Trust",
    "AIFreedomTrustFederation/www.aifreedomtrust.com",
)


class CollectionError(ValueError):
    """GitHub evidence could not be safely corroborated."""


def gh(*args: str) -> str:
    # API metadata is also untrusted input. Do not buffer unlimited stdout
    # before parsing JSON, even when the artifact archive itself is bounded.
    if len(args) != 2 or args[0] != "api":
        raise CollectionError("unsupported gh invocation")
    raw = bounded_archive_download("GitHub API metadata", args[1], MAX_METADATA_BYTES)
    try:
        return raw.decode("utf-8")
    except UnicodeDecodeError as exc:
        raise CollectionError("GitHub API metadata is not UTF-8") from exc


def finite_json_float(value: str) -> float:
    """Reject overflowing JSON exponents instead of accepting Python infinity."""
    number = float(value)
    if not math.isfinite(number):
        raise CollectionError("non-finite GitHub API number")
    return number


def api(path: str) -> dict:
    try:
        value = json.loads(gh("api", path), object_pairs_hook=consumer.no_duplicate_keys,
                           parse_constant=consumer.reject_constant,
                           parse_float=finite_json_float)
    except (ValueError, RecursionError, consumer.EvidenceError) as exc:
        raise CollectionError(f"invalid GitHub API response for {path}: {exc}") from exc
    if not isinstance(value, dict):
        raise CollectionError(f"invalid GitHub API object for {path}")
    return value



def bounded_archive_download(repo: str, endpoint: str, max_bytes: int = MAX_ARCHIVE_BYTES) -> bytes:
    """Read at most max_bytes + 1 bytes, with a wall-clock deadline.

    Do not use subprocess.run(capture_output=True) for untrusted binary content:
    that buffers the complete response before its size can be checked.
    """
    try:
        process = subprocess.Popen(["gh", "api", endpoint], stdout=subprocess.PIPE,
                                   stderr=subprocess.DEVNULL, bufsize=0)
    except OSError as exc:
        raise CollectionError(f"artifact download failed for {repo}: {exc}") from exc
    deadline = time.monotonic() + 60
    try:
        data = bytearray()
        with selectors.DefaultSelector() as selector:
            selector.register(process.stdout, selectors.EVENT_READ)
            while True:
                remaining = deadline - time.monotonic()
                if remaining <= 0 or not selector.select(remaining):
                    raise CollectionError(f"artifact download timed out for {repo}")
                chunk = os.read(process.stdout.fileno(), min(65536, max_bytes + 1 - len(data)))
                if not chunk:
                    break
                data.extend(chunk)
                if len(data) > max_bytes:
                    raise CollectionError(f"artifact archive missing or too large: {repo}")
        try:
            status = process.wait(timeout=max(0.001, deadline - time.monotonic()))
        except subprocess.TimeoutExpired as exc:
            raise CollectionError(f"artifact download timed out for {repo}") from exc
        if status:
            raise CollectionError(f"artifact ID download failed for {repo} (exit {status})")
        return bytes(data)
    finally:
        if process.poll() is None:
            process.kill()
            process.wait()
        process.stdout.close()


def download_record(repo: str, artifact_id: int, filename: str, digest: str, directory: Path,
                    expected_size: int) -> Path:
    """Fetch the *listed artifact ID*, not a potentially different same-name archive."""
    endpoint = f"repos/{repo}/actions/artifacts/{artifact_id}/zip"
    archive_bytes = bounded_archive_download(repo, endpoint)
    # Bind the downloaded archive length to the authenticated artifact metadata.
    # A digest match alone would not detect an inconsistent advertised size.
    if len(archive_bytes) != expected_size:
        raise CollectionError(f"artifact archive size does not match metadata: {repo}")
    # GitHub publishes a SHA-256 digest for Actions artifacts. Bind the
    # downloaded ZIP bytes to the exact artifact metadata we inspected.
    if not isinstance(digest, str) or not re.fullmatch(r"sha256:[0-9a-f]{64}", digest):
        raise CollectionError(f"missing or invalid artifact SHA-256 digest: {repo}")
    if hashlib.sha256(archive_bytes).hexdigest() != digest.removeprefix("sha256:"):
        raise CollectionError(f"artifact SHA-256 digest mismatch: {repo}")
    # These verification JSON records should be small. Bound archive and
    # expanded member sizes; never extract arbitrary paths or links.
    if not archive_bytes or len(archive_bytes) > MAX_ARCHIVE_BYTES:
        raise CollectionError(f"artifact archive missing or too large: {repo}")
    try:
        with zipfile.ZipFile(io.BytesIO(archive_bytes)) as archive:
            members = archive.infolist()
            if len(members) != 1:
                raise CollectionError(f"artifact must contain exactly one file: {repo}")
            member = members[0]
            kind = (member.external_attr >> 16) & 0o170000
            if (member.filename != filename or member.is_dir() or
                    kind not in (0, 0o100000) or member.file_size > 1024 * 1024):
                raise CollectionError(f"artifact contains an unsafe or unexpected file: {repo}")
            payload = archive.read(member)
    except (zipfile.BadZipFile, zipfile.LargeZipFile, zlib.error, RuntimeError, EOFError,
            OSError, NotImplementedError) as exc:
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
    # A mismatched mapping would be silently truncated by zip(), allowing an
    # assessment to claim repositories whose evidence files were not published.
    if not records or len(records) != len(provenance):
        raise CollectionError("incomplete or duplicate publication records")
    repositories = [entry.get("repository") for entry in provenance]
    if (len(set(repositories)) != len(repositories) or
            any(repo not in ARTIFACTS for repo in repositories)):
        raise CollectionError("incomplete or duplicate publication records")
    # Publication order must preserve producer identity; equal-length, unique
    # provenance alone would allow two source files to be silently swapped.
    for source, entry in zip(records, provenance):
        if source.name != ARTIFACTS[entry["repository"]][1]:
            raise CollectionError("publication source does not match repository")
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
        validated_records = {}
        provenance = []
        for index, (repo, run_id) in enumerate(sorted(runs.items())):
            sha = expected[repo]
            info = api(f"repos/{repo}")
            repo_id = info.get("id")
            if type(repo_id) is not int or repo_id <= 0:
                raise CollectionError(f"invalid repository identity: {repo}")
            branch = info.get("default_branch")
            if not isinstance(branch, str) or not branch or "/" in branch:
                raise CollectionError(f"invalid default branch for {repo}")
            branch_info = api(f"repos/{repo}/branches/{branch}")
            if not isinstance(branch_info.get("commit"), dict) or branch_info["commit"].get("sha") != sha:
                raise CollectionError(f"expected commit is not current {repo}/{branch}")
            run = api(f"repos/{repo}/actions/runs/{run_id}")
            # The API endpoint is selected by run ID, but independently bind
            # the returned workflow-run object to that exact identifier.
            if type(run.get("id")) is not int or run["id"] != int(run_id):
                raise CollectionError(f"run identity mismatch: {repo}")
            attempt = run.get("run_attempt")
            if type(attempt) is not int or attempt < 1:
                raise CollectionError(f"invalid workflow run attempt: {repo}")
            workflow_id, workflow_path = WORKFLOWS[repo]
            if (type(run.get("workflow_id")) is not int or run["workflow_id"] != workflow_id or
                    run.get("path") != workflow_path):
                raise CollectionError(f"run workflow identity mismatch: {repo}")
            # Bind the workflow's source and execution repositories to the
            # authenticated repository ID, not just its reusable full name.
            run_repo, head_repo = run.get("repository"), run.get("head_repository")
            if (not isinstance(run_repo, dict) or run_repo.get("full_name") != repo or
                    type(run_repo.get("id")) is not int or run_repo["id"] != repo_id or
                    not isinstance(head_repo, dict) or head_repo.get("full_name") != repo or
                    type(head_repo.get("id")) is not int or head_repo["id"] != repo_id or
                    run.get("head_sha") != sha or run.get("head_branch") != branch or
                    run.get("event") != "push" or run.get("status") != "completed" or
                    run.get("conclusion") not in ("success", "failure")):
                raise CollectionError(f"run metadata does not establish completed default-branch CI: {repo}")
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
            # The run-scoped artifact listing is not the only provenance
            # assertion: corroborate the artifact's embedded workflow_run
            # identity with the authenticated repository, run and branch.
            # Both current producers publish these GitHub metadata fields.
            artifact_run = artifact.get("workflow_run")
            if (not isinstance(artifact_run, dict) or
                    type(artifact_run.get("id")) is not int or
                    artifact_run["id"] != int(run_id) or
                    artifact_run.get("head_sha") != sha or
                    artifact_run.get("head_branch") != branch or
                    type(artifact_run.get("repository_id")) is not int or
                    artifact_run["repository_id"] != repo_id or
                    type(artifact_run.get("head_repository_id")) is not int or
                    artifact_run["head_repository_id"] != repo_id):
                raise CollectionError(f"artifact workflow run identity mismatch: {repo}")
            # Reject oversized or malformed artifacts before launching a binary download.
            # GitHub artifact metadata provides the compressed archive size.
            size = artifact.get("size_in_bytes")
            if type(size) is not int or not 0 < size <= MAX_ARCHIVE_BYTES:
                raise CollectionError(f"missing, invalid, or oversized artifact size: {repo}")
            directory = root / f"repo-{index}"
            directory.mkdir()
            downloaded = download_record(repo, artifact["id"], filename, artifact.get("digest"), directory,
                                         expected_size=size)
            record = consumer.read_record(downloaded)
            # Bind the payload to the repository whose authenticated artifact
            # supplied it, not merely to any expected Federation producer.
            if record.get("repository") != repo or record.get("commit_sha") != sha:
                raise CollectionError(f"artifact record does not match authenticated producer: {repo}")
            result = consumer.validate(record, expected)
            # A producer-reported timestamp is not an attestation, but it must
            # not claim generation later than the authenticated collection.
            # Tolerate modest clock skew between the producer and CI runner.
            generated = datetime.fromisoformat(record["generated_at"])
            if generated > datetime.now(timezone.utc) + timedelta(minutes=5):
                raise CollectionError(f"verification evidence timestamp is in the future: {repo}")
            # Self-reported generation time must also be plausible for the
            # authenticated workflow execution, not merely earlier than now.
            try:
                started = datetime.fromisoformat(run["run_started_at"])
                finished = datetime.fromisoformat(run["updated_at"])
                if started.tzinfo is None or finished.tzinfo is None or finished < started:
                    raise ValueError("invalid workflow execution interval")
            except (KeyError, TypeError, ValueError) as exc:
                raise CollectionError(f"invalid authenticated workflow timestamps: {repo}") from exc
            if not started - timedelta(minutes=5) <= generated <= finished + timedelta(minutes=5):
                raise CollectionError(f"verification evidence timestamp outside authenticated workflow window: {repo}")
            # Run IDs and commit SHAs survive workflow reruns. Corroborate the
            # artifact creation time against the latest authenticated attempt,
            # rather than relying solely on its run ID and self-reported time.
            try:
                created = datetime.fromisoformat(artifact["created_at"])
                if created.tzinfo is None or created.utcoffset() is None:
                    raise ValueError("artifact timestamp must be timezone-aware")
            except (KeyError, TypeError, ValueError) as exc:
                raise CollectionError(f"invalid authenticated artifact timestamp: {repo}") from exc
            # These timestamps all come from GitHub's API, unlike the
            # producer's generated_at. A five-minute tolerance here could
            # admit an artifact from a previous run attempt near a rerun.
            if not started <= created <= finished:
                raise CollectionError(f"artifact created outside authenticated workflow window: {repo}")
            # A failed producer workflow may publish valid negative evidence
            # through an if: always() artifact upload. Accept it only when the
            # authenticated record actually reports a failed check or gate.
            # Never turn an unrelated failed workflow into a verified pass.
            if run["conclusion"] == "failure" and result["status"] != "failed":
                raise CollectionError(f"failed CI run lacks failed verification evidence: {repo}")
            if (record.get("source", {}).get("type") != "github-actions" or
                    record["source"].get("run_url") != f"https://github.com/{repo}/actions/runs/{run_id}"):
                raise CollectionError(f"artifact source does not match authenticated run: {repo}")
            records.append(downloaded)
            validated_records[repo] = record
            provenance.append({"repository": repo, "repository_id": repo_id,
                               "run_id": int(run_id), "run_attempt": attempt,
                               "artifact_id": artifact["id"],
                               "artifact_digest": artifact["digest"], "commit_sha": sha,
                               "default_branch": branch, "workflow_id": workflow_id,
                               "workflow_path": workflow_path, "workflow_conclusion": run["conclusion"],
                               # Preserve authenticated API timing and byte-size metadata
                               # for downstream provenance audits (not a signed attestation).
                               "workflow_started_at": run["run_started_at"],
                               "workflow_updated_at": run["updated_at"],
                               "artifact_created_at": artifact["created_at"],
                               "artifact_size_in_bytes": size})
        assessment, _ = consumer.inspect(records, expected)
        # A verified record reporting a failed check or security gate is valid
        # negative evidence, not malformed input. Publish its failed status
        # while continuing to reject missing, duplicate, or invalid records.
        if assessment["errors"]:
            raise CollectionError("retrieved evidence failed structural verification")
        # Preserve producer-reported commands, outcomes, gate statuses and
        # limitations alongside independently checked GitHub provenance. These
        # fields remain self-reported, not independently verified test results.
        for summary in assessment["repositories"]:
            record = validated_records[summary["repository"]]
            summary["reported_checks"] = [
                {key: check[key] for key in ("id", "command", "status", "exit_code")}
                for check in record["checks"]
            ]
            summary["reported_gates"] = {
                gate: record[gate]["status"]
                for gate in ("security", "dependencies", "deployment")
            }
            summary["reported_limitations"] = record["limitations"]
        # A branch can advance while artifacts are downloaded. Recheck every
        # producer immediately before publishing a current-commit assessment.
        for entry in provenance:
            repo, branch, sha = entry["repository"], entry["default_branch"], entry["commit_sha"]
            latest_repo = api(f"repos/{repo}")
            if latest_repo.get("default_branch") != branch:
                raise CollectionError(f"default branch changed during collection: {repo}")
            if (type(latest_repo.get("id")) is not int or
                    latest_repo["id"] != entry["repository_id"]):
                raise CollectionError(f"repository identity changed during collection: {repo}")
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
        # A pass over supported producers is not a Federation-wide acceptance
        # result. Make that boundary machine-readable even for reported_pass.
        assessment["federation_scope"] = {
            "verification_level": "supported_producers_only",
            "supported_producers": sorted(ARTIFACTS),
            "known_unintegrated_repositories": sorted(KNOWN_UNINTEGRATED),
            "known_unintegrated_repositories_exhaustive": False,
            "federation_wide_acceptance": False,
        }
        assessment["provenance"] = "github_api_run_and_artifact_checked_unsigned"
        # Identify the exact verifier source bytes used for this assessment.
        # These are reproducibility fingerprints, NOT signed attestations.
        assessment["verifier_implementation"] = {
            "attestation": "local_file_sha256_unsigned",
            "collector": {
                "path": "scripts/collect_federation_evidence.py",
                "sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
            },
            "structural_consumer": {
                "path": "scripts/check_federation_evidence.py",
                "sha256": hashlib.sha256(Path(consumer.__file__).read_bytes()).hexdigest(),
            },
        }
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
        # Publish negative evidence, but retain a failing CLI exit status so
        # automation cannot mistake verified failed gates for a green run.
        return 1 if assessment["status"] == "failed" else 0
    except (CollectionError, consumer.EvidenceError, OSError) as exc:
        print(f"Federation evidence collection failed: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
