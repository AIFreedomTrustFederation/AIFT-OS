"""Offline regressions for credentialless pull worker and durable outbox."""
import base64
import importlib.util
import json
import os
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location("aift_pull_docs_worker", ROOT / "scripts/aift_pull_docs_worker.py")
poller = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(poller)
worker = poller.worker

ISSUE = {"number": 17, "state": "open",
         "labels": [{"name": worker.LABEL}],
         "body": "Target: docs/example.md\nInstruction: Replace 'checks' with 'verification'.\n"}
ORIGINAL = "# Example\n\nAIFT performs local checks.\n"
EDIT = {"old": "checks", "new": "verification", "summary": "Clarified terminology"}


def artifact():
    updated = worker.apply_exact_replacement(ORIGINAL, EDIT)
    return {
        "schema_version": 2, "head_sha": "a" * 40, "issue_number": 17,
        "issue_body_sha256": worker.sha256(ISSUE["body"]),
        "target": "docs/example.md", "original_sha256": worker.sha256(ORIGINAL),
        "old": EDIT["old"], "new": EDIT["new"], "summary": EDIT["summary"],
        "updated_sha256": worker.sha256(updated["content"]),
        "changed_lines": 1,
    }


class PullWorkerTests(unittest.TestCase):
    def test_anonymous_client_refuses_all_unapproved_routes(self):
        for route in (
            "/user", "https://example.com/",
            "/repos/AIFreedomTrustFederation/VPS/branches/main",
            "/repos/AIFreedomTrustFederation/AIFT-OS/issues/1",
        ):
            with self.subTest(route=route), self.assertRaises(worker.Blocked):
                poller.public_get(route)

    def test_anonymous_request_has_no_auth_or_redirect(self):
        request_seen = []
        class Response:
            def __enter__(self):
                return self
            def __exit__(self, *_):
                return False
            def read(self, _limit):
                return b'{"commit":{"sha":"' + b"a" * 40 + b'"}}'
        class Opener:
            def open(self, request, timeout):
                request_seen.append(request)
                return Response()
        with patch.object(poller, "build_opener", return_value=Opener()):
            self.assertEqual(poller.exact_public_main(), "a" * 40)
        self.assertEqual(len(request_seen), 1)
        request = request_seen[0]
        self.assertEqual(request.get_method(), "GET")
        self.assertEqual(request.host, "api.github.com")
        self.assertNotIn("Authorization", request.headers)
        self.assertNotIn("Cookie", request.headers)

    def test_private_outbox_is_outside_repository_and_restart_safe(self):
        with TemporaryDirectory() as temp:
            root = poller.prepare_state(Path(temp) / "state")
            self.assertEqual(root.stat().st_mode & 0o777, 0o700)
            self.assertEqual((root / "outbox").stat().st_mode & 0o777, 0o700)
            self.assertEqual(root, poller.prepare_state(root))

    def test_symlink_or_shared_state_refused(self):
        with TemporaryDirectory() as temp:
            p = Path(temp) / "state"
            p.mkdir(mode=0o755)
            with self.assertRaises(worker.Blocked):
                poller.prepare_state(p)
            p.chmod(0o700)
            link = Path(temp) / "link"
            link.symlink_to(p)
            with self.assertRaises(worker.Blocked):
                poller.prepare_state(link)

    def test_one_poll_queues_at_most_one_real_artifact_and_export_is_repeatable(self):
        with TemporaryDirectory() as temp:
            root = Path(temp) / "state"
            doc = Path(temp) / "example.md"
            doc.write_text(ORIGINAL, encoding="utf-8")
            selected = (ISSUE, "docs/example.md", "Replace 'checks' with 'verification'.")
            def choose(_read_token, *, fetcher, exclude_issues):
                self.assertIs(fetcher, poller.public_get)
                return selected if 17 not in exclude_issues else None
            with (patch.object(poller, "exact_public_main", return_value="a" * 40),
                  patch.object(worker, "run", return_value="main"),
                  patch.object(worker, "verify_checkout"),
                  patch.object(worker, "select_issue", side_effect=choose),
                  patch.object(worker, "safe_document", return_value=doc),
                  patch.object(worker, "model_edit", return_value=worker.apply_exact_replacement(ORIGINAL, EDIT))):
                result = poller.poll_once(root, model="aetherion:latest",
                                          endpoint="http://127.0.0.1:11434/v1")
                self.assertEqual(result["status"], "queued")
                self.assertEqual(result["pending"], 1)
                self.assertEqual(poller.poll_once(root, model="aetherion:latest",
                                                  endpoint="http://127.0.0.1:11434/v1")["status"], "idle")
            serialized = poller.export_plan(root, 17)
            self.assertEqual(json.loads(base64.b64decode(serialized)), artifact())
            self.assertEqual(poller.queue_status(root)["pending"], ["issue-17.json"])
            poller.mark_delivered(root, 17)
            self.assertEqual(poller.queue_status(root)["pending"], [])
            self.assertEqual(poller.queue_status(root)["archived"], 1)
            with self.assertRaises(worker.Blocked):
                poller.export_plan(root, 17)

    def test_duplicate_issue_is_skipped_at_selection_layer(self):
        issues = [
            {"number": 17, **{k: v for k, v in ISSUE.items() if k != "number"}},
            {"number": 18, **{k: v for k, v in ISSUE.items() if k != "number"}},
        ]
        def fetch(path, _token):
            return issues if "/issues?" in path else []
        with patch.object(worker, "safe_document", return_value=Path("docs/example.md")):
            selected = worker.select_issue("", fetcher=fetch, exclude_issues={17})
        self.assertEqual(selected[0]["number"], 18)

    def test_stale_main_refuses_model_or_issue_read(self):
        with TemporaryDirectory() as temp:
            with (patch.object(poller, "exact_public_main", return_value="a" * 40),
                  patch.object(worker, "run", return_value="main"),
                  patch.object(worker, "verify_checkout", side_effect=worker.Blocked("stale")),
                  patch.object(worker, "select_issue") as choose):
                with self.assertRaises(worker.Blocked):
                    poller.poll_once(Path(temp) / "state", model="aetherion:latest",
                                     endpoint="http://127.0.0.1:11434/v1")
                choose.assert_not_called()

    def test_queue_full_skips_network_and_model(self):
        with TemporaryDirectory() as temp:
            root = poller.prepare_state(Path(temp) / "state")
            for i in range(poller.MAX_PENDING):
                path = poller._plan_path(root, i + 1)
                path.write_text(json.dumps(artifact()), encoding="utf-8")
                path.chmod(0o600)
            with patch.object(poller, "exact_public_main") as api:
                self.assertEqual(poller.poll_once(root, model="aetherion:latest",
                                                  endpoint="http://127.0.0.1:11434/v1")["status"], "queue_full")
                api.assert_not_called()

    def test_outbox_refuses_tampering_and_symlinks(self):
        with TemporaryDirectory() as temp:
            root = poller.prepare_state(Path(temp) / "state")
            path = poller._plan_path(root, 17)
            path.write_text(json.dumps({**artifact(), "updated_content": "bad"}), encoding="utf-8")
            path.chmod(0o600)
            with self.assertRaises(worker.Blocked):
                poller.export_plan(root, 17)
            path.unlink()
            target = Path(temp) / "target"
            target.write_text(json.dumps(artifact()), encoding="utf-8")
            path.symlink_to(target)
            with self.assertRaises(worker.Blocked):
                poller.export_plan(root, 17)

    def test_source_has_no_secret_import_or_push(self):
        source = (ROOT / "scripts/aift_pull_docs_worker.py").read_text(encoding="utf-8")
        self.assertIn("NoRedirects", source)
        self.assertNotIn("gh auth", source)
        self.assertNotIn("gh pr", source)
        self.assertNotIn("git push", source)
        self.assertNotIn("Authorization", source.split('def public_get')[1].split('def exact_public_main')[0].split('Request(')[1].split('method=')[0])


if __name__ == "__main__":
    unittest.main()
