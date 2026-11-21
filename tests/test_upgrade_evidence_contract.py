"""Consumer contract using synthetic Runtime/Forge records, never live services."""
import copy
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
FIXTURES = ROOT / "tests/fixtures/upgrade-evidence"
SCRIPT = ROOT / "scripts/check_federation_evidence.py"


class UpgradeEvidenceContract(unittest.TestCase):
    def setUp(self):
        self.records = [json.loads((FIXTURES / (name + ".json")).read_text())
                        for name in ("AIFT-Runtime", "AIFT-Forge")]

    def consume(self, records):
        with tempfile.TemporaryDirectory() as tmp:
            cmd = [sys.executable, "-B", str(SCRIPT)]
            for original in self.records:
                cmd += ["--expect", original["repository"] + "=" + original["commit_sha"]]
            for i, record in enumerate(records):
                path = Path(tmp) / (str(i) + ".json")
                path.write_text(json.dumps(record))
                cmd += ["--input", str(path)]
            result = subprocess.run(cmd, capture_output=True, text=True, timeout=15)
        return result.returncode, json.loads(result.stdout)

    def test_passed_source_gates_remain_partial_and_unsigned(self):
        code, report = self.consume(self.records)
        self.assertEqual(code, 0)
        self.assertEqual(report["status"], "partial")
        self.assertEqual(report["provenance"], "unsigned_supplied_json_not_authenticated")
        for original, summary in zip(sorted(self.records, key=lambda r: r["repository"]), report["repositories"]):
            self.assertEqual(summary["reported_checks"], original["checks"])
            self.assertEqual(summary["reported_source"]["type"], "local")
            self.assertEqual(len(summary["unverified_gates"]), 3)
            self.assertEqual(summary["reported_limitations"], original["limitations"])

    def test_failed_qa_remains_failed_for_either_producer(self):
        for index in range(2):
            with self.subTest(index=index):
                records = copy.deepcopy(self.records)
                records[index]["checks"][0].update(status="failed", exit_code=7)
                records[index]["overall_status"] = "failed"
                code, report = self.consume(records)
                self.assertEqual(code, 1)
                self.assertEqual(report["status"], "failed")

    def test_stale_revision_and_missing_producer_are_rejected(self):
        records = copy.deepcopy(self.records)
        records[0]["commit_sha"] = "c" * 40
        for supplied in (records, self.records[:1]):
            with self.subTest(supplied=len(supplied)):
                code, report = self.consume(supplied)
                self.assertEqual(code, 1)
                self.assertTrue(report["errors"])

    def test_duplicate_producer_cannot_supply_missing_evidence(self):
        code, report = self.consume([self.records[0], self.records[0]])
        self.assertEqual(code, 1)
        self.assertTrue(report["errors"])
