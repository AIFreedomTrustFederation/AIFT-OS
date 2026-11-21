"""Safety and offline regression tests for the AIFT cloud AI documentation worker."""
from io import BytesIO
import importlib.util
import json
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location(
    "aift_secure_cloud_worker", ROOT / "scripts/aift_secure_cloud_worker.py"
)
worker = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(worker)
WORKFLOW = ROOT / ".github/workflows/aift-secure-cloud-worker.yml"


class MockHTTP(BytesIO):
    def __enter__(self):
        return self

    def __exit__(self, *_):
        self.close()


class SecureCloudWorkerTests(unittest.TestCase):
    def test_safe_document_rejects_non_docs_and_sensitive_paths(self):
        for path in (
            "../docs/demo.md", "/tmp/demo.md", ".github/workflows/demo.md",
            "docs/security.md", "docs/trust-gate.md", "docs/governance.md",
            "docs/secret-rotation.md", "docs/identity-guide.md", "docs/../README.md",
            "README.md", "docs/demo.py",
        ):
            with self.subTest(path=path), self.assertRaises(worker.Blocked):
                worker.safe_document(path)

    def test_edit_rejects_overlarge_non_newline_or_empty_changes(self):
        with self.assertRaises(worker.Blocked):
            worker.checked_edit("old\n", "old\n")
        with self.assertRaises(worker.Blocked):
            worker.checked_edit("old\n", "new")
        with self.assertRaises(worker.Blocked):
            worker.checked_edit("old\n", "new\n" * 130)
        with self.assertRaises(worker.Blocked):
            worker.checked_edit("old\n", "-----BEGIN PRIVATE KEY-----\n")

    def test_bounded_document_edit_is_accepted(self):
        self.assertEqual(worker.checked_edit("before\n", "after\n"), 1)

    def test_issue_requires_approved_label_and_explicit_target(self):
        body = "Target: docs/validation.md\nInstruction: Clarify the existing example.\n"
        original = {"number": 101, "state": "open", "body": body,
                    "labels": [{"name": worker.LABEL}]}
        with patch.object(worker, "safe_document", return_value=Path("docs/validation.md")):
            self.assertEqual(worker.parse_issue(original),
                             ("docs/validation.md", "Clarify the existing example."))
            with self.assertRaises(worker.Blocked):
                worker.parse_issue({**original, "labels": []})
            with self.assertRaises(worker.Blocked):
                worker.parse_issue({**original, "body": "Instruction: edit everything"})

    def test_select_issue_skips_existing_pr_and_invalid_issue(self):
        valid = {"number": 5, "state": "open",
                 "body": "Target: docs/validation.md\nInstruction: clarify text\n",
                 "labels": [{"name": worker.LABEL}]}
        invalid = {"number": 6, "state": "open", "body": "hi", "labels": [{"name": worker.LABEL}]}
        def api(path, _token):
            if "/issues?" in path:
                return [valid, invalid]
            return [{"body": "AIFT-AUTODEV-ISSUE:5"}]
        with patch.object(worker, "github_get", side_effect=api):
            self.assertIsNone(worker.select_issue("read_token"))

    def test_exact_head_mismatch_fails_closed(self):
        with patch.object(worker, "run", side_effect=lambda *cmd: "f" * 40):
            with self.assertRaises(worker.Blocked):
                worker.verify_checkout("e" * 40)
        with self.assertRaises(worker.Blocked):
            worker.verify_checkout("main")

    def test_responses_request_has_no_tools_or_storage_and_structured_output(self):
        response = {
            "status": "completed",
            "output": [{
                "type": "message", "status": "completed",
                "content": [{"type": "output_text", "text": json.dumps({
                    "content": "Improved\n", "summary": "Clarified language"
                })}]
            }]
        }
        requests = []
        def fake_urlopen(req, timeout):
            requests.append(json.loads(req.data))
            return MockHTTP(json.dumps(response).encode())
        with patch.object(worker, "urlopen", side_effect=fake_urlopen):
            result = worker.model_edit(
                api_key="test_only", model="gpt-6-luna",
                path="docs/example.md", content="Original\n",
                instruction="Clarify the text")
        self.assertEqual(result["content"], "Improved\n")
        self.assertEqual(requests[0]["tools"], [])
        self.assertIs(requests[0]["store"], False)
        self.assertEqual(requests[0]["text"]["format"]["type"], "json_schema")
        self.assertTrue(requests[0]["text"]["format"]["strict"])

    def test_plan_without_eligible_issue_does_not_call_model(self):
        with TemporaryDirectory() as tmp:
            destination = Path(tmp) / "plan.json"
            with patch.object(worker, "verify_checkout"):
                with patch.object(worker, "select_issue", return_value=None):
                    with patch.object(worker, "model_edit") as model:
                        self.assertFalse(worker.create_plan(
                            destination, read_token="fake", api_key="fake",
                            model="gpt-6-luna", expected_sha="a" * 40))
                        model.assert_not_called()
            self.assertFalse(destination.exists())

    def test_github_api_route_does_not_accept_arbitrary_hosts(self):
        with self.assertRaises(worker.Blocked):
            worker.github_get("/user", token="fake")
        with self.assertRaises(worker.Blocked):
            worker.github_get("/repos/other/repo/issues", token="fake")

    def test_workflow_separates_model_and_write_credentials(self):
        workflow = WORKFLOW.read_text(encoding="utf-8")
        self.assertIn("vars.AIFT_AUTODEV_AI_ENABLED == 'true'", workflow)
        self.assertIn("github.ref == 'refs/heads/main'", workflow)
        self.assertIn("persist-credentials: false", workflow)
        self.assertIn("permission-pull-requests: write", workflow)
        self.assertIn("permission-contents: write", workflow)
        self.assertIn("needs.plan.outputs.has_plan == 'true'", workflow)
        self.assertIn("AIFT_AUTODEV_OPENAI_API_KEY", workflow)
        self.assertIn("AIFT_AUTODEV_APP_PRIVATE_KEY", workflow)
        self.assertNotIn("pull_request_target:", workflow)
        self.assertNotIn("gh pr merge", workflow)
        self.assertNotIn("git push origin main", workflow)

    def test_publish_uses_draft_and_no_direct_merge(self):
        source = (ROOT / "scripts/aift_secure_cloud_worker.py").read_text(encoding="utf-8")
        self.assertIn('"--draft"', source)
        self.assertIn('"--base", "main"', source)
        self.assertIn("AIFT-AUTODEV-ISSUE:", source)
        self.assertNotIn("git merge", source)
        self.assertNotIn("subprocess.run([\"bash\"", source)


if __name__ == "__main__":
    unittest.main()
