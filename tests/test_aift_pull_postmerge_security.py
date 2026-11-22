"""Post-merge security regressions: live main, queue rejection, path isolation."""
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import subprocess
from tempfile import TemporaryDirectory
from types import SimpleNamespace
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location(
    "aift_pull_docs_security", ROOT / "scripts/aift_pull_docs_worker.py"
)
poller = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(poller)
worker = poller.worker


class PullSecurityTests(unittest.TestCase):
    def test_remote_main_matches_exact_dispatched_head(self):
        def remote(*args, **kwargs):
            self.assertEqual(args[0], ["git", "ls-remote", "--heads", "origin", "refs/heads/main"])
            self.assertEqual(kwargs["timeout"], 20)
            return SimpleNamespace(returncode=0,
                stdout="a" * 40 + "\trefs/heads/main\n")
        with patch.object(worker.subprocess, "run", side_effect=remote):
            worker.verify_live_main("a" * 40)

    def test_remote_main_move_unreachable_or_invalid_is_blocked(self):
        for response in (
            SimpleNamespace(returncode=0, stdout="b" * 40 + "\trefs/heads/main\n"),
            SimpleNamespace(returncode=128, stdout=""),
            SimpleNamespace(returncode=0, stdout="malformed"),
            SimpleNamespace(returncode=0, stdout="a" * 40 + "\trefs/heads/other\n"),
        ):
            with self.subTest(response=response.stdout), patch.object(
                worker.subprocess, "run", return_value=response
            ), self.assertRaises(worker.Blocked):
                worker.verify_live_main("a" * 40)
        with patch.object(worker.subprocess, "run",
                          side_effect=subprocess.TimeoutExpired(["git"], 20)):
            with self.assertRaises(worker.Blocked):
                worker.verify_live_main("a" * 40)
        with self.assertRaises(worker.Blocked):
            worker.verify_live_main("invalid")

    def test_preflight_blocks_remote_advance_before_loading_artifact(self):
        with TemporaryDirectory() as tmp:
            missing = Path(tmp) / "missing-plan.json"
            with patch.object(worker, "verify_checkout"), patch.object(
                worker, "verify_live_main", side_effect=worker.Blocked("stale")
            ), self.assertRaises(worker.Blocked):
                worker.preflight_plan(missing, read_token="fake", expected_sha="a" * 40)

    def test_state_path_normalization_blocks_traversal_and_symlinked_parent(self):
        with TemporaryDirectory() as tmp:
            root = Path(tmp)
            repo = root / "repo"
            repo.mkdir()
            link = root / "link"
            link.symlink_to(repo, target_is_directory=True)
            with patch.object(poller.Path, "cwd", return_value=repo):
                bad_paths = (
                    root / "scratch" / ".." / "repo" / ".git" / "outbox",
                    link / ".git" / "outbox",
                    repo / "internal",
                )
                for path in bad_paths:
                    with self.subTest(path=str(path)), self.assertRaises(worker.Blocked):
                        poller.prepare_state(path)
                    self.assertFalse((repo / ".git" / "outbox").exists())
                outside = root / "outside"
                resolved = poller.prepare_state(outside)
                self.assertEqual(resolved, outside)
                with self.assertRaises(worker.Blocked):
                    poller.prepare_state(link)

    def test_discard_quarantines_unpublished_plan_without_delivery(self):
        with TemporaryDirectory() as tmp:
            root = poller.prepare_state(Path(tmp) / "state")
            record = {field: "" for field in worker.PLAN_KEYS}
            record["schema_version"] = worker.PLAN_SCHEMA_VERSION
            record["issue_number"] = 23
            raw = (json.dumps(record) + "\n").encode("utf-8")
            digest = hashlib.sha256(raw).hexdigest()
            source = root / "outbox" / "issue-23.json"
            source.write_bytes(raw)
            source.chmod(0o600)
            with self.assertRaises(worker.Blocked):
                poller.discard_plan(root, 23, expected_digest="f" * 64)
            self.assertTrue(source.exists())
            name = poller.discard_plan(root, 23, expected_digest=digest)
            self.assertTrue((root / "discarded" / name).exists())
            self.assertFalse(source.exists())
            self.assertEqual(poller.queue_status(root)["archived"], 0)
            self.assertEqual(poller.queue_status(root)["pending"], [])
            with self.assertRaises(worker.Blocked):
                poller.discard_plan(root, 23, expected_digest=digest)

    def test_handoff_has_independent_publication_gate(self):
        workflow = (ROOT / ".github/workflows/aift-pull-doc-publisher.yml").read_text()
        module = (ROOT / "scripts/aift_secure_cloud_worker.py").read_text()
        self.assertIn("verify_live_main(expected_sha)", module)
        self.assertIn("verify_live_main(expected_sha)", module.split("def publish_plan(")[1])
        self.assertIn("verify_live_main(expected_sha)", module.split("def preflight_plan(")[1])
        self.assertLess(workflow.index("aift_pull_plan_handoff.py"),
                        workflow.index("create-github-app-token@"))


if __name__ == "__main__":
    unittest.main()
