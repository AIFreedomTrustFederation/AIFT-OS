"""Offline tests for local Android App key bootstrap; synthetic PEMs only."""
import importlib.util
import os
from pathlib import Path
import sys
import tempfile
import unittest
from unittest.mock import patch

SCRIPT = Path(__file__).parents[1] / "scripts" / "bootstrap_local_app_key.py"
spec = importlib.util.spec_from_file_location("aift_bootstrap_test_module", SCRIPT)
bootstrap = importlib.util.module_from_spec(spec)
spec.loader.exec_module(bootstrap)


class BootstrapTests(unittest.TestCase):
    def setUp(self):
        temp = tempfile.TemporaryDirectory()
        self.addCleanup(temp.cleanup)
        self.directory = Path(temp.name)
        (self.directory / "existing-app.pem").write_bytes(
            b"-----BEGIN RSA PRIVATE KEY-----\nSYNTHETIC-NOT-REAL\n")

    def run_main(self, *args):
        with patch.object(sys, "argv", ["bootstrap_local_app_key.py", "--downloads-dir",
                                        str(self.directory), *args]):
            return bootstrap.main()

    def test_preview_does_not_read_private_key_or_use_app(self):
        with patch.object(bootstrap, "App") as app, patch.object(
                bootstrap.subprocess, "call", return_value=0) as cli:
            self.assertEqual(self.run_main(), 0)
            app.assert_not_called()
            argv = cli.call_args.args[0]
            self.assertNotIn("--apply", argv)
            self.assertNotIn("--private-key", argv)

    def test_ambiguous_pems_fail_closed(self):
        (self.directory / "other.pem").write_bytes(
            b"-----BEGIN PRIVATE KEY-----\nSYNTHETIC-NOT-REAL\n")
        with patch.object(bootstrap.subprocess, "call") as cli:
            self.assertEqual(self.run_main("--apply"), 1)
            cli.assert_not_called()

    def test_apply_discovers_client_id_validates_producer_scopes_and_cleans_temp(self):
        captured = {}
        class Api:
            def request(self, method, path):
                if path.startswith("/repos/") and "/actions/" not in path:
                    return {"full_name": path[len("/repos/"):]}
                if path.endswith("/actions/workflows?per_page=1"):
                    return {"total_count": 1}
                raise AssertionError("Unexpected route")
        class App:
            def __init__(self, keypath):
                captured["private_path"] = Path(keypath)
                captured["private_mode"] = captured["private_path"].stat().st_mode & 0o777
            def identify(self):
                return {"id": 5861093, "client_id": "Iv1.a1b2c3"}, {"id": 55}
            def token(self, repo, permissions):
                self_outer.assertEqual(permissions, bootstrap.EVIDENCE)
                captured.setdefault("repos", []).append(repo)
                return Api()
        self_outer = self
        with patch.object(bootstrap, "App", App), patch.object(
                bootstrap.subprocess, "call", return_value=0) as cli:
            self.assertEqual(self.run_main("--apply"), 0)
            argv = cli.call_args.args[0]
            self.assertEqual(argv[argv.index("--client-id") + 1], "Iv1.a1b2c3")
            self.assertEqual(captured["private_mode"], 0o600)
            self.assertEqual(len(captured["repos"]), 2)
            self.assertFalse(captured["private_path"].exists())
            self.assertNotIn("SYNTHETIC-NOT-REAL", " ".join(argv))

    def test_mismatched_client_id_refuses_writes(self):
        class App:
            def __init__(self, *_):
                pass
            def identify(self):
                return {"id": 5861093, "client_id": "Iv1.actual"}, {"id": 55}
        with patch.object(bootstrap, "App", App), patch.object(
                bootstrap.subprocess, "call") as cli:
            self.assertEqual(self.run_main("--apply", "--client-id", "Iv1.wrong"), 1)
            cli.assert_not_called()

    def test_missing_producer_access_refuses_writes(self):
        class Api:
            def request(self, method, path):
                return {"full_name": "not-the-requested-repository"}
        class App:
            def __init__(self, *_):
                pass
            def identify(self):
                return {"id": 5861093, "client_id": "Iv1.actual"}, {"id": 55}
            def token(self, *_):
                return Api()
        with patch.object(bootstrap, "App", App), patch.object(
                bootstrap.subprocess, "call") as cli:
            self.assertEqual(self.run_main("--apply"), 1)
            cli.assert_not_called()


if __name__ == "__main__":
    unittest.main()
