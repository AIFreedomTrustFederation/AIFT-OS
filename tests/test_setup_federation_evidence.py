import argparse
import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location("setup", Path(__file__).parents[1] / "scripts/setup_federation_evidence.py")
setup = importlib.util.module_from_spec(spec)
spec.loader.exec_module(setup)

class SetupTests(unittest.TestCase):
    def args(self, **kwargs):
        return argparse.Namespace(client_id=None, private_key=None, apply=False, rerun=None, **kwargs)

    def test_preview_never_reads_or_writes_key(self):
        with patch.object(setup, "api", return_value={"full_name": setup.REPO, "permissions": {"admin": True}}), patch.object(setup, "gh") as cli, patch.object(setup, "read_key") as key:
            setup.configure(self.args())
            cli.assert_not_called()
            key.assert_not_called()

    def test_rejects_non_admin_before_mutation(self):
        with patch.object(setup, "api", return_value={"full_name": setup.REPO, "permissions": {"admin": False}}), patch.object(setup, "gh") as cli:
            with self.assertRaises(setup.SetupError):
                setup.configure(self.args())
            cli.assert_not_called()

    def test_key_sent_only_through_stdin(self):
        args = self.args()
        args.apply, args.client_id, args.private_key = True, "Iv1_example", Path("unused")
        marker = b"synthetic-key"
        with patch.object(setup, "api", return_value={"full_name": setup.REPO, "permissions": {"admin": True}}), patch.object(setup, "read_key", return_value=marker), patch.object(setup, "gh") as cli:
            setup.configure(args)
            self.assertEqual(cli.call_args_list[0].kwargs["input"], marker)
            self.assertNotIn(marker, cli.call_args_list[0].args)
            self.assertEqual(cli.call_count, 2)

    def test_rejects_pr_or_stale_run(self):
        for event, sha in [("pull_request", "current"), ("push", "old")]:
            run = {"id": 1, "event": event, "head_sha": sha, "head_branch": "main",
                   "path": setup.WORKFLOW, "status": "completed", "conclusion": "failure",
                   "repository": {"full_name": setup.REPO}, "head_repository": {"full_name": setup.REPO}}
            with patch.object(setup, "api", side_effect=[{"commit": {"sha": "current"}}, run]):
                with self.assertRaises(setup.SetupError):
                    setup.validate_run(1)

    def test_rejects_insecure_key_permissions(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "fixture.pem"
            path.write_bytes(b"-----BEGIN PRIVATE KEY-----\nfixture")
            path.chmod(0o644)
            with self.assertRaises(setup.SetupError):
                setup.read_key(path)

    def test_failure_does_not_continue_writes(self):
        args = self.args()
        args.apply, args.client_id, args.private_key = True, "example", Path("unused")
        with patch.object(setup, "api", return_value={"full_name": setup.REPO, "permissions": {"admin": True}}), patch.object(setup, "read_key", return_value=b"fixture"), patch.object(setup, "gh", side_effect=setup.SetupError("failed")) as cli:
            with self.assertRaises(setup.SetupError):
                setup.configure(args)
            self.assertEqual(cli.call_count, 1)

if __name__ == "__main__":
    unittest.main()
