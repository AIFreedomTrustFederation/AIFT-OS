"""Offline tests: no GitHub credentials and no network calls."""
import base64
import hashlib
import hmac
import json
import os
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest
from unittest.mock import patch

from federation_automation.core import (
    App, Blocked, Queue, READ, WRITE, doc_ok, receive, repo_ok, signature_ok, work
)


class WorkerSafetyTests(unittest.TestCase):
    def setUp(self):
        temp = TemporaryDirectory()
        self.addCleanup(temp.cleanup)
        self.path = str(Path(temp.name) / "jobs.db")
        self.now = [1000.]
        self.q = Queue(self.path, clock=lambda: self.now[0])
        self.repo = "AIFreedomTrustFederation/AIFT-OS"

    def job(self, kind="audit", **extra):
        self.q.enqueue("job", kind, {"repo": self.repo, **extra})
        return dict(self.q.db.execute("SELECT * FROM jobs WHERE id='job'").fetchone())

    def test_webhook_signature_hmac_sha256(self):
        secret, body = b"secret" * 8, b'{"test":true}'
        header = "sha256=" + hmac.new(secret, body, hashlib.sha256).hexdigest()
        self.assertTrue(signature_ok(secret, body, header))
        self.assertFalse(signature_ok(secret, body + b"x", header))
        self.assertFalse(signature_ok(secret, body, "sha1=" + header[7:]))
        self.assertFalse(signature_ok(b"", body, header))

    def test_fail_closed_pause_and_resume(self):
        self.assertTrue(self.q.paused())
        self.job()
        self.assertIsNone(self.q.claim())
        self.q.set("paused", "0")
        self.assertEqual(self.q.claim()["id"], "job")
        self.q.set("paused", "1")
        self.assertIsNone(self.q.claim())

    def test_queue_duplicate_and_delivery_idempotence(self):
        self.assertTrue(self.q.enqueue("one", "audit", {"repo": self.repo}))
        self.assertFalse(self.q.enqueue("one", "audit", {"repo": self.repo}))
        delivery_id = "11223344-5566-7788-9900-aabbccddeeff"
        self.assertTrue(self.q.delivery(delivery_id))
        self.assertFalse(self.q.delivery(delivery_id))

    def test_worker_restart_lease_recovery(self):
        self.q.set("paused", "0")
        self.job()
        self.assertEqual(self.q.claim()["id"], "job")
        self.assertIsNone(self.q.claim())
        self.now[0] += 241
        recovered = Queue(self.path, clock=lambda: self.now[0]).claim()
        self.assertEqual(recovered["id"], "job")
        self.assertEqual(recovered["attempts"], 1)

    def test_retry_backoff_and_dead_letter(self):
        self.q.set("paused", "0")
        self.job()
        for attempt in range(4):
            job = self.q.claim()
            self.assertEqual(self.q.finish(job, error=Blocked("secret must not leak")),
                             "dead" if attempt == 3 else "retry")
            self.now[0] += 4000
        self.assertIsNone(self.q.claim())
        stored = self.q.db.execute("SELECT result FROM jobs WHERE id='job'").fetchone()[0]
        self.assertNotIn("secret must not leak", stored)
        self.assertEqual(json.loads(stored)["error_type"], "Blocked")

    def test_reject_unsafe_repo_and_path(self):
        repo_ok(self.repo)
        doc_ok("docs/developer-guide.md")
        for value in ("somebody/repo", self.repo + "/../bad"):
            with self.assertRaises(Blocked):
                repo_ok(value)
        for value in ("README.md", "docs/../README.md", "docs/governance.md",
                      "docs/identity/a.md", ".github/ci.md", "docs/security.md"):
            with self.assertRaises(Blocked):
                doc_ok(value)

    def test_signed_event_dedup_by_delivery_and_head(self):
        sha = "a" * 40
        payload = {"repository": {"full_name": self.repo, "default_branch": "main"},
                   "sender": {"type": "User"}, "ref": "refs/heads/main", "after": sha}
        d1 = "11223344-5566-7788-9900-aabbccddeeff"
        d2 = "aabbccdd-5566-7788-9900-aabbccddeeff"
        self.assertEqual(receive(self.q, "push", payload, d1), "queued")
        self.assertEqual(receive(self.q, "push", payload, d1), "duplicate_delivery")
        self.assertEqual(receive(self.q, "push", payload, d2), "duplicate_head")

    def test_bot_and_nondefault_events_do_not_queue(self):
        payload = {"repository": {"full_name": self.repo, "default_branch": "main"},
                   "sender": {"type": "Bot"}, "ref": "refs/heads/main", "after": "a" * 40}
        self.assertEqual(receive(self.q, "push", payload,
                         "11223344-5566-7788-9900-aabbccddeeff"), "bot_ignored")
        payload["sender"] = {"type": "User"}
        payload["ref"] = "refs/heads/feature"
        self.assertEqual(receive(self.q, "push", payload,
                         "aabbccdd-5566-7788-9900-aabbccddeeff"), "non_default_push")

    def test_audit_exact_head_never_claims_merge_readiness(self):
        requested = []
        class FakeAPI:
            def request(self, method, path, body=None):
                requested.append((method, path))
                if path.endswith("/AIFT-OS"):
                    return {"full_name": "AIFreedomTrustFederation/AIFT-OS"}
                if "/pulls?" in path:
                    return [{"number": 9, "draft": True, "head": {"sha": "a" * 40},
                             "base": {"sha": "b" * 40}}]
                return {"check_runs": [{"head_sha": "a" * 40, "status": "completed",
                                        "conclusion": "success", "name": "test"}]}
        class FakeAuth:
            def token(self, repo, permissions):
                assert repo == "AIFreedomTrustFederation/AIFT-OS"
                assert permissions == READ
                return FakeAPI()
        report = work(self.job(), FakeAuth())
        self.assertIsNone(report["open_pr_sample"][0]["merge_ready"])
        self.assertEqual(report["merge_ready"], "unverified")
        self.assertEqual(len(requested), 3)

    def test_proposals_off_until_explicit_opt_in(self):
        with patch.dict(os.environ, {"AIFT_ENABLE_DOC_PROPOSALS": "0"}):
            with self.assertRaises(Blocked):
                work(self.job("doc_format", path="docs/readme.md"),
                     object(), {self.repo})

    def test_scope_is_exact_and_rejects_admin_grant(self):
        with self.assertRaises(Blocked):
            App("/nonexistent").token(self.repo, {"administration": "write"})
        self.assertNotEqual(READ, WRITE)

    def test_draft_proposal_creates_only_isolated_branch(self):
        original = "# Title  \nBody\n"
        writes = []
        class FakeAPI:
            def request(self, method, path, body=None):
                writes.append((method, path, body))
                if path.endswith("/AIFT-OS"):
                    return {"default_branch": "main"}
                if "/branches/" in path:
                    return {"commit": {"sha": "c" * 40}}
                if "/contents/" in path and method == "GET":
                    return {"encoding": "base64", "size": len(original),
                            "content": base64.b64encode(original.encode()).decode(),
                            "sha": "d" * 40}
                if method == "POST" and path.endswith("/pulls"):
                    return {"draft": True, "html_url": "https://github.com/AIFreedomTrustFederation/AIFT-OS/pull/999",
                            "number": 999}
                return {}
        class FakeAuth:
            def token(self, repo, permissions):
                assert permissions in (READ, WRITE)
                return FakeAPI()
        with patch.dict(os.environ, {"AIFT_ENABLE_DOC_PROPOSALS": "1"}):
            report = work(self.job("doc_format", path="docs/format-guide.md"),
                          FakeAuth(), {self.repo})
        self.assertEqual(report["outcome"], "draft_pr")
        self.assertEqual([entry[0] for entry in writes].count("PUT"), 1)
        self.assertTrue([entry for entry in writes if entry[0] == "POST" and
                         entry[1].endswith("/pulls")][0][2]["draft"])
        self.assertFalse(any(entry[1].endswith("/merges") for entry in writes))

    def test_operator_recent_jobs_and_schedule_evidence(self):
        self.job()
        self.q.set("last_scheduled", "1000")
        self.assertEqual(self.q.status()["last_scheduled"], 1000.)
        recent = self.q.recent()
        self.assertEqual(recent[0]["repo"], self.repo)
        self.assertEqual(recent[0]["state"], "queued")
        self.assertIsNone(recent[0]["result"])

    def test_metadata_only_installation_discovery_refuses_extra_scope(self):
        responses = []
        class FakeAPI:
            def __init__(self, token):
                self.token = token
            def request(self, method, path, body=None):
                responses.append(body)
                return {"token": "ephemeral", "permissions": {"metadata": "read"}}
        auth = App("synthetic-not-read", factory=FakeAPI)
        with patch.object(auth, "identify", return_value=({}, {"id": 123})), patch.object(
                auth, "jwt", return_value="synthetic-jwt"):
            self.assertIsInstance(auth.discovery_api(), FakeAPI)
        self.assertEqual(responses[0]["permissions"], {"metadata": "read"})

    def test_readonly_console_is_local_and_offers_no_public_mutations(self):
        html = (Path(__file__).parents[1] / "federation_automation" /
                "console.html").read_text(encoding="utf-8")
        self.assertIn("PAUSED", html)
        self.assertIn("/api/jobs", html)
        self.assertIn("No inferred merge readiness", html)
        self.assertNotIn('fetch("/control/', html)


if __name__ == "__main__":
    unittest.main()
