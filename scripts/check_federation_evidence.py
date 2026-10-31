#!/usr/bin/env python3
"""Validate supplied AIFT v1 evidence against explicitly expected repository commits.

This is an offline structural consumer, not an attestation or GitHub artifact fetcher.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import math
import re
import sys
from pathlib import Path

REPO = re.compile(r"AIFreedomTrustFederation/[A-Za-z0-9_.-]+\Z")
SHA = re.compile(r"[0-9a-f]{40}\Z")
GATE_STATUSES = {"passed", "failed", "not_checked", "not_applicable"}
MAX_BYTES = 1024 * 1024
MAX_INTEGER_DIGITS = 4300


class EvidenceError(ValueError):
    """Invalid or mismatched evidence."""


def no_duplicate_keys(pairs: list[tuple[str, object]]) -> dict:
    result = {}
    for key, value in pairs:
        if key in result:
            raise EvidenceError(f"duplicate JSON key: {key}")
        result[key] = value
    return result


def reject_constant(value: str) -> None:
    raise EvidenceError(f"invalid numeric literal: {value}")


def finite_json_float(value: str) -> float:
    """Reject JSON exponent overflow instead of silently accepting infinity."""
    number = float(value)
    if not math.isfinite(number):
        raise EvidenceError("non-finite JSON number")
    return number


def bounded_json_int(value: str) -> int:
    """Bound JSON integer parsing independent of Python's interpreter limit."""
    if len(value.lstrip("-")) > MAX_INTEGER_DIGITS:
        raise EvidenceError("JSON integer exceeds 4300 digits")
    return int(value)


def read_record(path: Path) -> dict:
    # A stat-only limit is vulnerable to a file growing between stat and read.
    # Bound bytes actually consumed, including for changing or special files.
    with path.open("rb") as stream:
        raw = stream.read(MAX_BYTES + 1)
    if len(raw) > MAX_BYTES:
        raise EvidenceError("evidence exceeds 1 MiB")
    try:
        record = json.loads(raw.decode("utf-8"), object_pairs_hook=no_duplicate_keys,
                            parse_constant=reject_constant, parse_float=finite_json_float,
                            parse_int=bounded_json_int)
    except (UnicodeError, ValueError, RecursionError) as exc:
        raise EvidenceError(f"invalid JSON: {exc}") from exc
    if not isinstance(record, dict):
        raise EvidenceError("record must be a JSON object")
    return record


def validate(record: dict, expected: dict[str, str]) -> dict:
    if record.get("schema") != "aift.verification.v1":
        raise EvidenceError("unsupported verification schema")
    repo, commit = record.get("repository"), record.get("commit_sha")
    if not isinstance(repo, str) or not REPO.fullmatch(repo):
        raise EvidenceError("invalid repository identity")
    if repo not in expected:
        raise EvidenceError(f"unexpected repository: {repo}")
    if not isinstance(commit, str) or not SHA.fullmatch(commit) or commit != expected[repo]:
        raise EvidenceError(f"commit mismatch for {repo}")
    try:
        generated = dt.datetime.fromisoformat(record["generated_at"])
    except (KeyError, TypeError, ValueError) as exc:
        raise EvidenceError("invalid generated_at") from exc
    if generated.tzinfo is None or generated.utcoffset() is None:
        raise EvidenceError("generated_at must be timezone-aware")
    source = record.get("source")
    if not isinstance(source, dict) or source.get("type") not in ("local", "github-actions"):
        raise EvidenceError("invalid source type")
    run_url = source.get("run_url")
    if source["type"] == "github-actions" and run_url is None:
        raise EvidenceError("GitHub Actions source requires a run URL")
    if run_url is not None and (not isinstance(run_url, str) or
                                not re.fullmatch(rf"https://github\.com/{re.escape(repo)}/actions/runs/[1-9][0-9]*", run_url)):
        raise EvidenceError("source run URL does not match repository")
    checks = record.get("checks")
    if not isinstance(checks, list) or not checks:
        raise EvidenceError("checks must be a nonempty list")
    seen = set()
    for check in checks:
        if not isinstance(check, dict):
            raise EvidenceError("check must be an object")
        check_id, command, status, code = (check.get(k) for k in ("id", "command", "status", "exit_code"))
        if not isinstance(check_id, str) or not check_id.strip() or check_id in seen:
            raise EvidenceError("duplicate or invalid check id")
        seen.add(check_id)
        if not isinstance(command, str) or not command.strip():
            raise EvidenceError(f"invalid command for {check_id}")
        if status not in ("passed", "failed") or type(code) is not int:
            raise EvidenceError(f"invalid check status or exit code for {check_id}")
        if (status == "passed") != (code == 0):
            raise EvidenceError(f"contradictory status and exit code for {check_id}")
    derived = "passed" if all(c["status"] == "passed" for c in checks) else "failed"
    if record.get("overall_status") != derived:
        raise EvidenceError("overall status contradicts check results")
    gaps = []
    failed_gates = []
    for gate in ("security", "dependencies", "deployment"):
        value = record.get(gate)
        if (not isinstance(value, dict) or not isinstance(value.get("status"), str)
                or value["status"] not in GATE_STATUSES):
            raise EvidenceError(f"invalid or missing {gate} gate")
        if value["status"] in ("not_checked", "not_applicable"):
            gaps.append(f"{gate}:{value['status']}")
        elif value["status"] == "failed":
            failed_gates.append(gate)
    limits = record.get("limitations")
    if not isinstance(limits, list) or not all(isinstance(x, str) and x.strip() for x in limits):
        raise EvidenceError("limitations must be a list of nonblank strings")
    if derived == "failed" or failed_gates:
        status = "failed"
    elif gaps:
        status = "reported_checks_passed_with_gaps"
    else:
        status = "reported_checks_passed"
    return {"repository": repo, "commit_sha": commit, "status": status,
            "unverified_gates": gaps, "failed_gates": failed_gates,
            "check_count": len(checks)}


def inspect(paths: list[Path], expected: dict[str, str]) -> tuple[dict, int]:
    results = []
    observed = set()
    errors = []
    for path in paths:
        try:
            record = read_record(path)
            result = validate(record, expected)
            if result["repository"] in observed:
                raise EvidenceError(f"duplicate evidence for {result['repository']}")
            observed.add(result["repository"])
            results.append(result)
        except (OSError, EvidenceError) as exc:
            errors.append({"file": str(path), "error": str(exc)})
    missing = sorted(set(expected) - observed)
    for repo in missing:
        errors.append({"repository": repo, "error": "expected repository evidence missing"})
    status = "failed" if errors or any(r["status"] == "failed" for r in results) else (
        "partial" if any(r["unverified_gates"] for r in results) else "reported_pass")
    return {"schema": "aift.federation-assessment.v1", "status": status,
            "provenance": "unsigned_supplied_json_not_authenticated",
            "repositories": results, "errors": errors}, 1 if status == "failed" else 0


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--expect", action="append", required=True, metavar="OWNER/REPO=SHA")
    parser.add_argument("--input", action="append", required=True, type=Path, metavar="RECORD.json")
    args = parser.parse_args(argv)
    expected = {}
    for item in args.expect:
        repo, sep, sha = item.partition("=")
        if not sep or not REPO.fullmatch(repo) or not SHA.fullmatch(sha) or repo in expected:
            parser.error(f"invalid or duplicate --expect: {item}")
        expected[repo] = sha
    report, code = inspect(args.input, expected)
    print(json.dumps(report, sort_keys=True, indent=2))
    return code


if __name__ == "__main__":
    sys.exit(main())
