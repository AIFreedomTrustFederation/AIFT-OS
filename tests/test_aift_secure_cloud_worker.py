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

    def test_exact_span_preserves_unrelated_lines_and_trailing_newline(self):
        original = "# AIFT Example\n\nAIFT performs local checks.\n"
        result = worker.apply_exact_replacement(original, {
            "old": "local checks",
            "new": "local verification",
            "summary": "Replaced terminology",
        })
        self.assertEqual(
            result["content"], "# AIFT Example\n\nAIFT performs local verification.\n")
        self.assertEqual(worker.checked_edit(original, result["content"]), 1)

    def test_exact_span_rejects_ambiguous_invalid_or_multiline_edits(self):
        original = "# Heading\n\nlocal checks\nlocal checks\n"
        examples = [
            {"old": "local checks", "new": "local verification", "summary": "x"},
            {"old": "missing", "new": "replacement", "summary": "x"},
            {"old": "local checks", "new": "local checks", "summary": "x"},
            {"old": "local checks", "new": "different\nunapproved line", "summary": "x"},
            {"old": "local checks\n", "new": "changed", "summary": "x"},
            {"old": "", "new": "something", "summary": "x"},
            {"old": "local checks", "new": "local verification"},
        ]
        for item in examples:
            with self.subTest(item=item), self.assertRaises(worker.Blocked):
                worker.apply_exact_replacement(original, item)

    def test_exact_span_fails_if_original_lacks_required_final_newline(self):
        with self.assertRaises(worker.Blocked):
            worker.apply_exact_replacement("Original", {
                "old": "Original", "new": "Changed", "summary": "x"})

    def test_aetherion_loopback_structured_completion_and_no_credentials(self):
        response = {"choices": [{"message": {"content": json.dumps({
            "old": "Original", "new": "Improved", "summary": "Clarified language"
        })}}]}
        requests = []
        class LocalConnection:
            def __init__(self, host, port, timeout):
                requests.append({"host": host, "port": port, "timeout": timeout})
            def request(self, method, path, body, headers):
                requests[-1].update({"method": method, "path": path,
                                     "body": json.loads(body), "headers": headers})
            def getresponse(self):
                class Result:
                    status = 200
                    def read(self, limit):
                        return json.dumps(response).encode()
                return Result()
            def close(self):
                pass
        with patch.object(worker.http.client, "HTTPConnection", LocalConnection):
            result = worker.model_edit(
                model="aetherion.gguf", endpoint="http://127.0.0.1:8080/v1",
                path="docs/example.md", content="Original\n",
                instruction="Clarify the text")
        self.assertEqual(result["content"], "Improved\n")
        self.assertEqual(requests[0]["host"], "127.0.0.1")
        self.assertEqual(requests[0]["port"], 8080)
        self.assertEqual(requests[0]["path"], "/v1/chat/completions")
        self.assertIs(requests[0]["body"]["stream"], False)
        self.assertEqual(requests[0]["body"]["model"], "aetherion.gguf")
        self.assertNotIn("Authorization", requests[0]["headers"])
        self.assertNotIn("tools", requests[0]["body"])
        self.assertEqual(requests[0]["body"]["temperature"], 0)
        self.assertEqual(requests[0]["body"]["response_format"], {
            "type": "json_schema",
            "json_schema": {"name": "aift_document_edit", "strict": True,
                            "schema": worker.SCHEMA},
        })
        self.assertEqual(set(worker.SCHEMA["required"]), {"old", "new", "summary"})
        self.assertIs(worker.SCHEMA["additionalProperties"], False)

    def test_local_model_refuses_remote_dns_https_and_redirect_urls(self):
        for url in (
            "https://api.openai.com/v1", "http://localhost:8080/v1",
            "http://192.168.1.7:8080/v1", "http://127.0.0.1:8080/v1?url=example",
            "http://127.0.0.1:bad/v1", "http://127.0.0.1:8080/secret",
            "http://127.0.0.1:8080/v1#fragment", "http://user@127.0.0.1:8080/v1",
        ):
            with self.subTest(url=url), self.assertRaises(worker.Blocked):
                worker._loopback_model_endpoint(url)

    def test_unavailable_local_model_fails_without_cloud_fallback(self):
        with patch.object(worker.http.client, "HTTPConnection") as connection:
            connection.return_value.request.side_effect = OSError("offline")
            with self.assertRaisesRegex(worker.Blocked, "no cloud fallback"):
                worker.model_edit(model="local", path="docs/example.md",
                                  content="Original\n", instruction="Clarify the text")
            self.assertEqual(connection.call_args.args[0], "127.0.0.1")

    def test_local_model_rejects_bad_or_unstructured_output(self):
        class LocalConnection:
            def __init__(self, *_args, **_kwargs):
                pass
            def request(self, *_args, **_kwargs):
                pass
            def getresponse(self):
                class Result:
                    status = 200
                    def read(self, limit):
                        return b'{"choices":[{"message":{"content":"Not JSON"}}]}'
                return Result()
            def close(self):
                pass
        with patch.object(worker.http.client, "HTTPConnection", LocalConnection):
            with self.assertRaises(worker.Blocked):
                worker.model_edit(model="local", path="docs/example.md",
                                  content="Original\n", instruction="Clarify the text")

    def test_truncated_local_completion_fails_closed(self):
        class LocalConnection:
            def __init__(self, *_args, **_kwargs):
                pass
            def request(self, *_args, **_kwargs):
                pass
            def getresponse(self):
                class Result:
                    status = 200
                    def read(self, limit):
                        return json.dumps({"choices": [{
                            "finish_reason": "length",
                            "message": {"content": '{"content":"partial","summary":"x"}'}
                        }]}).encode()
                return Result()
            def close(self):
                pass
        with patch.object(worker.http.client, "HTTPConnection", LocalConnection):
            with self.assertRaisesRegex(worker.Blocked, "truncated"):
                worker.model_edit(model="local", path="docs/example.md",
                                  content="Original\n", instruction="Clarify the text")

    def test_plan_without_eligible_issue_does_not_call_model(self):
        with TemporaryDirectory() as tmp:
            destination = Path(tmp) / "plan.json"
            with patch.object(worker, "verify_checkout"):
                with patch.object(worker, "select_issue", return_value=None):
                    with patch.object(worker, "model_edit") as model:
                        self.assertFalse(worker.create_plan(
                            destination, read_token="fake",
                            model="local", expected_sha="a" * 40))
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
        self.assertIn("runs-on: [self-hosted, linux, aift-local-model]", workflow)
        self.assertIn("AIFT_MODEL_URL", workflow)
        self.assertNotIn("AIFT_AUTODEV_OPENAI_API_KEY", workflow)
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
