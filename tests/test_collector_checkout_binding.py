"""Real Git regressions for Federation collector checkout integrity."""
from pathlib import Path
import os
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/verify_collector_checkout.sh"
WORKFLOW = ROOT / ".github/workflows/federation-evidence-collector.yml"


class CollectorCheckoutBinding(unittest.TestCase):
    def test_workflow_runs_binding_in_both_jobs(self):
        source = WORKFLOW.read_text(encoding="utf-8")
        offline, live = source.split("\n  live-producer-evidence:\n", 1)
        invocation = "bash scripts/verify_collector_checkout.sh"
        self.assertIn(invocation, offline)
        self.assertIn(invocation, live)
        self.assertLess(live.index(invocation), live.index("Mint scoped producer evidence token"))

    def _run(self, mutation=None, sha_override=None, omit_sha=False, bad_workspace=False):
        with tempfile.TemporaryDirectory() as directory:
            repo = Path(directory) / "repo"
            repo.mkdir()
            def git(*args):
                return subprocess.check_output(["git", "-C", str(repo), *args], stderr=subprocess.DEVNULL, text=True).strip()
            git("init", "-q")
            git("config", "user.email", "audit@example.invalid")
            git("config", "user.name", "Audit")
            (repo / "tracked.txt").write_text("baseline\n")
            git("add", "tracked.txt")
            git("commit", "-qm", "baseline")
            original = git("rev-parse", "HEAD")
            if mutation == "advance":
                git("commit", "--allow-empty", "-qm", "advance")
            elif mutation == "tracked":
                (repo / "tracked.txt").write_text("modified\n")
            elif mutation == "untracked":
                (repo / "unexpected.txt").write_text("new\n")
            env = os.environ.copy()
            env["GITHUB_WORKSPACE"] = str(repo / "nonexistent" if bad_workspace else repo)
            if omit_sha:
                env.pop("GITHUB_SHA", None)
            else:
                env["GITHUB_SHA"] = sha_override if sha_override is not None else original
            return subprocess.run(["bash", str(SCRIPT)], env=env, capture_output=True, text=True, timeout=5)

    def test_clean_exact_commit_accepted(self):
        self.assertEqual(self._run().returncode, 0)

    def test_advanced_clean_head_rejected(self):
        self.assertNotEqual(self._run("advance").returncode, 0)

    def test_modified_tracked_file_rejected(self):
        self.assertNotEqual(self._run("tracked").returncode, 0)

    def test_untracked_file_rejected(self):
        self.assertNotEqual(self._run("untracked").returncode, 0)

    def test_missing_sha_rejected(self):
        self.assertNotEqual(self._run(omit_sha=True).returncode, 0)

    def test_malformed_sha_rejected(self):
        self.assertNotEqual(self._run(sha_override="not-a-commit").returncode, 0)

    def test_missing_workspace_rejected(self):
        self.assertNotEqual(self._run(bad_workspace=True).returncode, 0)


if __name__ == "__main__":
    unittest.main()
