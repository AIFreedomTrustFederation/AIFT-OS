#!/usr/bin/env python3
"""Read-only, bounded live smoke test for the AIFT-OS Aetherion docs planner.

No issue selection, repository edits, runner registration, App token, or PR publishing.
The only network request is delegated to model_edit's literal-loopback guard.
"""
from __future__ import annotations

import argparse
from datetime import datetime, timezone
import json
import os
import re
import statistics
import subprocess
import sys
import time
from typing import Any, Callable

try:
    from scripts.aift_secure_cloud_worker import model_edit
except ModuleNotFoundError:  # Direct execution: python scripts/aift_aetherion_readiness.py
    from aift_secure_cloud_worker import model_edit

SOURCE = "# AIFT Example\n\nAIFT performs local checks.\n"
EXPECTED = "# AIFT Example\n\nAIFT performs local verification.\n"
INSTRUCTION = "Replace 'local checks' with 'local verification'. Preserve all other content."
SHA_RE = re.compile(r"[0-9a-f]{40}\Z")


def current_sha() -> str:
    return subprocess.check_output(
        ["git", "rev-parse", "HEAD"], text=True, stderr=subprocess.DEVNULL,
    ).strip()


def evaluate(*, repetitions: int, model: str, endpoint: str,
             expected_sha: str | None = None,
             infer: Callable[..., dict[str, str]] = model_edit,
             clock: Callable[[], float] = time.monotonic,
             resolve_sha: Callable[[], str] = current_sha) -> tuple[dict[str, Any], int]:
    """Validate one synthetic edit repeatedly; never log responses or source contents."""
    if not 1 <= repetitions <= 10:
        raise ValueError("repetitions must be between 1 and 10")
    report: dict[str, Any] = {
        "schema_version": 1,
        "checked_at_utc": datetime.now(timezone.utc).isoformat(),
        "status": "blocked",
        "source_sha": None,
        "checks": {"exact_head": "not_requested", "local_worker_inference": "not_checked"},
        "model": model,
        "attempted": 0,
        "passed": 0,
        "failed": 0,
        "durations_ms": [],
        "median_duration_ms": None,
        "max_duration_ms": None,
        "runner_registration": "not_checked",
        "publisher": "not_checked",
        "production_deployment": "not_checked",
    }
    if expected_sha is not None:
        if not SHA_RE.fullmatch(expected_sha):
            report["checks"]["exact_head"] = "invalid_expected_sha"
            return report, 1
        try:
            observed = resolve_sha()
        except (OSError, subprocess.CalledProcessError):
            report["checks"]["exact_head"] = "unavailable"
            return report, 1
        report["source_sha"] = observed if SHA_RE.fullmatch(observed) else None
        if observed != expected_sha:
            report["checks"]["exact_head"] = "mismatch"
            return report, 1
        report["checks"]["exact_head"] = "passed"
    for _ in range(repetitions):
        start = clock()
        passed = False
        try:
            result = infer(
                model=model, endpoint=endpoint, path="docs/example.md",
                content=SOURCE, instruction=INSTRUCTION,
            )
            passed = isinstance(result, dict) and result.get("content") == EXPECTED
        except Exception:  # Never print exception messages; they may include model output.
            passed = False
        duration_ms = max(0, round((clock() - start) * 1000))
        report["durations_ms"].append(duration_ms)
        report["attempted"] += 1
        report["passed" if passed else "failed"] += 1
    report["median_duration_ms"] = round(statistics.median(report["durations_ms"]))
    report["max_duration_ms"] = max(report["durations_ms"])
    if report["passed"] == repetitions:
        report["status"] = "local_worker_inference_verified"
        report["checks"]["local_worker_inference"] = "passed"
        return report, 0
    report["checks"]["local_worker_inference"] = "failed"
    return report, 1


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repetitions", type=int, default=5)
    parser.add_argument("--model", default=os.environ.get("AIFT_MODEL_NAME") or "aetherion:latest")
    parser.add_argument("--model-url", default=os.environ.get("AIFT_MODEL_URL") or "http://127.0.0.1:11434/v1")
    parser.add_argument("--expected-sha", default=None)
    args = parser.parse_args(argv)
    if not 1 <= args.repetitions <= 10:
        parser.error("--repetitions must be between 1 and 10")
    report, exit_code = evaluate(
        repetitions=args.repetitions, model=args.model, endpoint=args.model_url,
        expected_sha=args.expected_sha,
    )
    print(json.dumps(report, sort_keys=True))
    return exit_code


if __name__ == "__main__":
    sys.exit(main())
