#!/usr/bin/env python3
"""Black-box regression tests for the offline Federation evidence consumer."""

import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

SCRIPT = Path(__file__).resolve().parents[1] / "scripts/check_federation_evidence.py"
REPOS = ("AIFreedomTrustFederation/.github", "AIFreedomTrustFederation/AIFT-Runtime")
SHAS = ("a" * 40, "b" * 40)


def record(repo, sha):
    return {"schema": "aift.verification.v1", "repository": repo, "commit_sha": sha,
            "generated_at": "2026-10-24T18:00:00+00:00", "source": {"type": "github-actions", "run_url": None},
            "checks": [{"id": "source", "command": "python -m unittest", "status": "passed", "exit_code": 0}],
            "overall_status": "passed", "security": {"status": "not_checked"},
            "dependencies": {"status": "not_checked"}, "deployment": {"status": "not_applicable"},
            "limitations": ["Does not test production."]}


class ConsumerTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)

    def run_consumer(self, records, raw=None, expect=None):
        expect = dict(zip(REPOS, SHAS)) if expect is None else expect
        cmd = [sys.executable, "-B", str(SCRIPT)]
        for repo, sha in expect.items():
            cmd.extend(["--expect", f"{repo}={sha}"])
        for i, value in enumerate(records):
            path = self.root / f"record-{i}.json"
            path.write_text(raw if raw is not None else json.dumps(value), encoding="utf-8")
            cmd.extend(["--input", str(path)])
        return subprocess.run(cmd, capture_output=True, text=True, check=False)

    def test_accepts_two_records_but_marks_unverified_gates(self):
        result = self.run_consumer([record(*pair) for pair in zip(REPOS, SHAS)])
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertEqual(report["status"], "partial")
        self.assertEqual(len(report["repositories"]), 2)
        self.assertEqual(report["provenance"], "unsigned_supplied_json_not_authenticated")

    def test_missing_repository_fails_closed(self):
        result = self.run_consumer([record(REPOS[0], SHAS[0])])
        self.assertEqual(result.returncode, 1)
        self.assertIn("expected repository evidence missing", result.stdout)

    def test_stale_or_wrong_commit_fails_closed(self):
        result = self.run_consumer([record(REPOS[0], "c" * 40), record(REPOS[1], SHAS[1])])
        self.assertEqual(result.returncode, 1)
        self.assertIn("commit mismatch", result.stdout)

    def test_duplicate_repository_fails_closed(self):
        result = self.run_consumer([record(REPOS[0], SHAS[0]), record(REPOS[0], SHAS[0]),
                                    record(REPOS[1], SHAS[1])])
        self.assertEqual(result.returncode, 1)
        self.assertIn("duplicate evidence", result.stdout)

    def test_rejects_false_success(self):
        broken = record(REPOS[0], SHAS[0])
        broken["checks"][0]["exit_code"] = 1
        result = self.run_consumer([broken, record(REPOS[1], SHAS[1])])
        self.assertEqual(result.returncode, 1)
        self.assertIn("contradictory status", result.stdout)

    def test_rejects_failed_checks_even_when_consistent(self):
        broken = record(REPOS[0], SHAS[0])
        broken["checks"][0].update(status="failed", exit_code=1)
        broken["overall_status"] = "failed"
        result = self.run_consumer([broken, record(REPOS[1], SHAS[1])])
        self.assertEqual(result.returncode, 1)
        self.assertEqual(json.loads(result.stdout)["status"], "failed")

    def test_rejects_malformed_and_duplicate_json_keys(self):
        result = self.run_consumer([record(REPOS[0], SHAS[0])], raw='{"schema":"aift.verification.v1","schema":"aift.verification.v1"}')
        self.assertEqual(result.returncode, 1)
        self.assertIn("duplicate JSON key", result.stdout)

    def test_rejects_incorrect_source_url(self):
        broken = record(REPOS[0], SHAS[0])
        broken["source"]["run_url"] = "https://github.com/another/repo/actions/runs/42"
        result = self.run_consumer([broken, record(REPOS[1], SHAS[1])])
        self.assertEqual(result.returncode, 1)
        self.assertIn("source run URL", result.stdout)

    def test_rejects_missing_gate(self):
        broken = record(REPOS[0], SHAS[0])
        del broken["security"]
        result = self.run_consumer([broken, record(REPOS[1], SHAS[1])])
        self.assertEqual(result.returncode, 1)
        self.assertIn("security gate", result.stdout)

    def test_failed_gate_does_not_pass(self):
        broken = record(REPOS[0], SHAS[0])
        broken["security"]["status"] = "failed"
        result = self.run_consumer([broken, record(REPOS[1], SHAS[1])])
        self.assertEqual(result.returncode, 1)
        self.assertEqual(json.loads(result.stdout)["status"], "failed")

    def test_rejects_non_string_gate_status_without_crashing(self):
        broken = record(REPOS[0], SHAS[0])
        broken["security"]["status"] = ["passed"]
        result = self.run_consumer([broken, record(REPOS[1], SHAS[1])])
        self.assertEqual(result.returncode, 1)
        self.assertIn("invalid or missing security gate", result.stdout)
        self.assertNotIn("Traceback", result.stderr)

    def test_rejects_boolean_exit_code(self):
        broken = record(REPOS[0], SHAS[0])
        broken["checks"][0]["exit_code"] = False
        result = self.run_consumer([broken, record(REPOS[1], SHAS[1])])
        self.assertEqual(result.returncode, 1)
        self.assertIn("invalid check status", result.stdout)


if __name__ == "__main__":
    unittest.main()
