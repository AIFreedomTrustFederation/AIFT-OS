"""Read-only publisher preflight and owner-only hosted dispatch contracts."""
import importlib.util
import json
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location(
    "aift_secure_cloud_worker_preflight", ROOT / "scripts/aift_secure_cloud_worker.py"
)
worker = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(worker)
WORKFLOW = ROOT / ".github/workflows/aift-pull-doc-publisher.yml"
ISSUE = {
    "number": 23, "state": "open",
    "labels": [{"name": worker.LABEL}],
    "body": "Target: docs/example.md\nInstruction: Clarify the example.\n",
}
ORIGINAL = "# Example\n\nThe docs use local checks.\n"


def plan():
    edit = {"old": "local checks", "new": "local verification", "summary": "Clarified terminology"}
    proposal = worker.apply_exact_replacement(ORIGINAL, edit)
    return {
        "schema_version": worker.PLAN_SCHEMA_VERSION,
        "head_sha": "a" * 40,
        "issue_number": 23,
        "issue_body_sha256": worker.sha256(ISSUE["body"]),
        "target": "docs/example.md",
        "original_sha256": worker.sha256(ORIGINAL),
        "old": edit["old"], "new": edit["new"], "summary": edit["summary"],
        "updated_sha256": worker.sha256(proposal["content"]),
        "changed_lines": worker.checked_edit(ORIGINAL, proposal["content"]),
    }


class PullPublisherTests(unittest.TestCase):
    def test_preflight_validates_without_app_credential_or_git_writes(self):
        with TemporaryDirectory() as temp:
            doc = Path(temp) / "docs.md"
            doc.write_text(ORIGINAL, encoding="utf-8")
            artifact = Path(temp) / "plan.json"
            artifact.write_text(json.dumps(plan()), encoding="utf-8")
            with (
                patch.object(worker, "verify_checkout"),
                patch.object(worker, "verify_live_main"),
                patch.object(worker, "safe_document", return_value=doc),
                patch.object(worker, "github_get", return_value=ISSUE) as read,
                patch.object(worker.subprocess, "run", side_effect=AssertionError("No git writes")) as git,
            ):
                record, verified = worker.preflight_plan(
                    artifact, read_token="read-only", expected_sha="a" * 40,
                )
            self.assertEqual(record["issue_number"], 23)
            self.assertEqual(verified["content"], "# Example\n\nThe docs use local verification.\n")
            self.assertTrue(read.called)
            git.assert_not_called()
            self.assertEqual(doc.read_text(encoding="utf-8"), ORIGINAL)

    def test_preflight_rejects_payload_tampering_and_changed_issue(self):
        with TemporaryDirectory() as temp:
            doc = Path(temp) / "docs.md"
            doc.write_text(ORIGINAL, encoding="utf-8")
            artifact = Path(temp) / "plan.json"
            for tweak in (
                lambda r: r.update({"updated_content": "bad\n"}),
                lambda r: r.update({"head_sha": "b" * 40}),
                lambda r: r.update({"new": "extra\nsecond line"}),
                lambda r: r.update({"updated_sha256": "f" * 64}),
            ):
                candidate = plan()
                tweak(candidate)
                artifact.write_text(json.dumps(candidate), encoding="utf-8")
                with (
                    patch.object(worker, "verify_checkout"),
                patch.object(worker, "verify_live_main"),
                    patch.object(worker, "safe_document", return_value=doc),
                    patch.object(worker, "github_get", return_value=ISSUE),
                    self.assertRaises(worker.Blocked),
                ):
                    worker.preflight_plan(artifact, read_token="read-only", expected_sha="a" * 40)
            artifact.write_text(json.dumps(plan()), encoding="utf-8")
            with (
                patch.object(worker, "verify_checkout"),
                patch.object(worker, "verify_live_main"),
                patch.object(worker, "safe_document", return_value=doc),
                patch.object(worker, "github_get", return_value={**ISSUE, "labels": []}),
                self.assertRaises(worker.Blocked),
            ):
                worker.preflight_plan(artifact, read_token="read-only", expected_sha="a" * 40)

    def test_preflight_refuses_symlink_and_large_artifact(self):
        with TemporaryDirectory() as temp:
            path = Path(temp) / "big.json"
            path.write_bytes(b"x" * 8193)
            with (patch.object(worker, "verify_checkout"),
                  patch.object(worker, "verify_live_main"),
                  self.assertRaises(worker.Blocked)):
                worker.preflight_plan(path, read_token="read-only", expected_sha="a" * 40)
            target = Path(temp) / "target.json"
            target.write_text(json.dumps(plan()), encoding="utf-8")
            path.unlink()
            path.symlink_to(target)
            with (patch.object(worker, "verify_checkout"),
                  patch.object(worker, "verify_live_main"),
                  self.assertRaises(worker.Blocked)):
                worker.preflight_plan(path, read_token="read-only", expected_sha="a" * 40)

    def test_workflow_never_gives_model_access_or_executes_pr_head(self):
        source = WORKFLOW.read_text(encoding="utf-8")
        self.assertIn("github.actor == 'AIFreedomTrustFederation'", source)
        self.assertIn("github.ref == 'refs/heads/main'", source)
        self.assertIn("inputs.operator_approved == true", source)
        self.assertIn("github.event_name == 'workflow_dispatch'", source)
        self.assertIn("preflight", source)
        self.assertIn("permission-contents: write", source)
        self.assertIn("permission-pull-requests: write", source)
        self.assertIn("persist-credentials: false", source)
        self.assertIn("ref: $" "{{ github.sha }}", source)
        self.assertNotIn("runs-on: [self-hosted", source)
        self.assertNotIn("labels: [self-hosted", source)
        self.assertNotIn("AIFT_MODEL_URL", source)
        self.assertNotIn("pull_request_target:", source)
        self.assertNotIn("git push origin main", source)
        self.assertLess(source.index("scripts/aift_secure_cloud_worker.py preflight"),
                        source.index("actions/create-github-app-token@"))

    def test_legacy_actions_worker_is_not_implicitly_enabled(self):
        source = (ROOT / ".github/workflows/aift-secure-cloud-worker.yml").read_text(encoding="utf-8")
        self.assertIn("vars.AIFT_AUTODEV_AI_ENABLED == 'true'", source)
        self.assertIn("aift-isolated-ephemeral", source)


if __name__ == "__main__":
    unittest.main()
