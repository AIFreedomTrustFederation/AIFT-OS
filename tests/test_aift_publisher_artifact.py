"""Publisher artifact provenance: all checks offline; never pushes branches."""
import copy
import importlib.util
import json
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location(
    "aift_secure_cloud_worker_artifact", ROOT / "scripts/aift_secure_cloud_worker.py"
)
worker = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(worker)

ORIGINAL = "# Example\n\nAIFT performs local checks.\n"
ISSUE = {
    "number": 17,
    "state": "open",
    "labels": [{"name": worker.LABEL}],
    "body": "Target: docs/example.md\nInstruction: replace the approved phrase.\n",
}
REPLACEMENT = {
    "old": "local checks",
    "new": "local verification",
    "summary": "Clarified terminology",
}


def make_artifact():
    updated = worker.apply_exact_replacement(ORIGINAL, REPLACEMENT)
    return {
        "schema_version": worker.PLAN_SCHEMA_VERSION,
        "head_sha": "a" * 40,
        "issue_number": 17,
        "issue_body_sha256": worker.sha256(ISSUE["body"]),
        "target": "docs/example.md",
        "original_sha256": worker.sha256(ORIGINAL),
        "old": REPLACEMENT["old"],
        "new": REPLACEMENT["new"],
        "summary": REPLACEMENT["summary"],
        "updated_sha256": worker.sha256(updated["content"]),
        "changed_lines": worker.checked_edit(ORIGINAL, updated["content"]),
    }


class PublisherArtifactTests(unittest.TestCase):
    def validate(self, record, issue=ISSUE):
        with patch.object(worker, "safe_document", return_value=Path("docs/example.md")):
            return worker.validate_plan_artifact(
                record, original=ORIGINAL, expected_sha="a" * 40, current_issue=issue
            )

    def test_publisher_reconstructs_single_span_not_full_document(self):
        record = make_artifact()
        self.assertEqual(set(record), worker.PLAN_KEYS)
        self.assertNotIn("updated_content", record)
        result = self.validate(record)
        self.assertEqual(result["content"], "# Example\n\nAIFT performs local verification.\n")
        self.assertEqual(result["old"], REPLACEMENT["old"])

    def test_artifact_tampering_fails_closed(self):
        invalid = []
        legacy = make_artifact()
        legacy["schema_version"] = 1
        legacy["updated_content"] = "malicious\n"
        invalid.append(legacy)
        payload = make_artifact()
        payload["updated_content"] = "untrusted complete document\n"
        invalid.append(payload)
        for key, value in (
            ("old", "not-in-source"),
            ("new", "local verification\nunapproved"),
            ("new", "unapproved substitution"),
            ("updated_sha256", "f" * 64),
            ("original_sha256", "e" * 64),
            ("issue_body_sha256", "d" * 64),
            ("changed_lines", 120),
            ("changed_lines", True),
            ("head_sha", "b" * 40),
            ("target", "docs/other.md"),
            ("issue_number", 18),
            ("issue_number", True),
            ("summary", "Looks fine\nCloses #999"),
        ):
            corrupted = make_artifact()
            corrupted[key] = value
            invalid.append(corrupted)
        for i, bad in enumerate(invalid):
            with self.subTest(case=i), self.assertRaises(worker.Blocked):
                self.validate(bad)
        with self.assertRaises(worker.Blocked):
            self.validate(make_artifact(), {**ISSUE, "labels": []})
        with self.assertRaises(worker.Blocked):
            self.validate(make_artifact(), {**ISSUE, "body": ISSUE["body"] + "changed"})

    def test_rejects_broken_plan_before_any_git_write(self):
        with TemporaryDirectory() as directory:
            source = Path(directory) / "example.md"
            source.write_text(ORIGINAL, encoding="utf-8")
            artifact = Path(directory) / "plan.json"
            record = make_artifact()
            record["old"] = "missing from original"
            artifact.write_text(json.dumps(record), encoding="utf-8")
            with (
                patch.object(worker, "verify_checkout"),
                patch.object(worker, "safe_document", return_value=source),
                patch.object(worker, "github_get", return_value=ISSUE),
                patch.object(worker.subprocess, "run", side_effect=AssertionError("git ran")) as git,
                self.assertRaises(worker.Blocked),
            ):
                worker.publish_plan(
                    artifact, read_token="fake-read-token",
                    expected_sha="a" * 40, gh_token="fake-write-token",
                )
            git.assert_not_called()
            self.assertEqual(source.read_text(encoding="utf-8"), ORIGINAL)

    def test_planner_only_emits_validated_span_and_hash(self):
        with TemporaryDirectory() as directory:
            source = Path(directory) / "example.md"
            source.write_text(ORIGINAL, encoding="utf-8")
            artifact = Path(directory) / "plan.json"
            model_result = worker.apply_exact_replacement(ORIGINAL, REPLACEMENT)
            with (
                patch.object(worker, "verify_checkout"),
                patch.object(worker, "safe_document", return_value=source),
                patch.object(worker, "select_issue", return_value=(ISSUE, "docs/example.md", "edit")),
                patch.object(worker, "model_edit", return_value=model_result),
            ):
                self.assertTrue(worker.create_plan(
                    artifact, read_token="read", model="aetherion:latest",
                    endpoint="http://127.0.0.1:11434/v1", expected_sha="a" * 40,
                ))
            stored = json.loads(artifact.read_text(encoding="utf-8"))
            self.assertEqual(self.validate(stored)["content"], model_result["content"])
            self.assertNotIn("updated_content", stored)
            self.assertEqual(stored["schema_version"], 2)

    def test_planner_refuses_model_that_misreports_applied_edit(self):
        with TemporaryDirectory() as directory:
            source = Path(directory) / "example.md"
            source.write_text(ORIGINAL, encoding="utf-8")
            artifact = Path(directory) / "plan.json"
            dishonest = {**REPLACEMENT, "content": "Additional text.\n"}
            with (
                patch.object(worker, "verify_checkout"),
                patch.object(worker, "safe_document", return_value=source),
                patch.object(worker, "select_issue", return_value=(ISSUE, "docs/example.md", "edit")),
                patch.object(worker, "model_edit", return_value=dishonest),
                self.assertRaises(worker.Blocked),
            ):
                worker.create_plan(artifact, read_token="read",
                                   model="aetherion:latest",
                                   endpoint="http://127.0.0.1:11434/v1",
                                   expected_sha="a" * 40)
            self.assertFalse(artifact.exists())

    def test_model_configuration_rejects_missing_and_remote(self):
        for model, endpoint in (
            ("", "http://127.0.0.1:11434/v1"),
            ("aetherion:latest", ""),
            ("aetherion:latest", "https://example.com/v1"),
            ("bad model", "http://127.0.0.1:11434/v1"),
        ):
            with self.subTest(model=model, endpoint=endpoint):
                with self.assertRaises(worker.Blocked):
                    worker.validate_model_config(endpoint, model)

    def test_plan_refuses_missing_coordinates_before_issue_lookup(self):
        with TemporaryDirectory() as directory:
            for model, endpoint in (
                ("", "http://127.0.0.1:11434/v1"),
                ("aetherion:latest", ""),
            ):
                with (
                    patch.object(worker, "verify_checkout"),
                    patch.object(worker, "select_issue", side_effect=AssertionError("issue queried")) as select,
                    self.assertRaises(worker.Blocked),
                ):
                    worker.create_plan(
                        Path(directory) / "plan.json",
                        read_token="fake", model=model, endpoint=endpoint,
                        expected_sha="a" * 40,
                    )
                select.assert_not_called()



if __name__ == "__main__":
    unittest.main()
