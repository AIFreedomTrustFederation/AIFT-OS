"""Regression contract for opt-in live evidence dispatch.

No third-party YAML parser needed; assert the guarded workflow source contract.
"""
from pathlib import Path
import re
import unittest

WORKFLOW = Path(__file__).parents[1] / ".github/workflows/federation-evidence-collector.yml"

class DispatchContract(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.source = WORKFLOW.read_text(encoding="utf-8")

    def test_manual_dispatch_declared(self):
        self.assertRegex(self.source, r"(?m)^  workflow_dispatch:\s*$")

    def test_live_job_only_for_trusted_main_dispatch(self):
        match = re.search(r"(?m)^  live-producer-evidence:\s*\n(?P<block>(?:    .*\n)*)", self.source)
        self.assertIsNotNone(match)
        guard = match.group("block")
        self.assertIn("github.event_name == 'workflow_dispatch'", guard)
        self.assertIn("github.ref == 'refs/heads/main'", guard)
        self.assertIn("github.event_name == 'push'", guard)
        self.assertIn("github.event_name == 'schedule'", guard)
        self.assertNotIn("github.event_name == 'pull_request'", guard)

    def test_live_job_still_requires_offline_qa(self):
        self.assertRegex(self.source, r"(?m)^    needs: collector$")
        self.assertIn("Require producer evidence App credentials", self.source)
        self.assertIn("Verify trusted collector checkout before App token creation", self.source)

if __name__ == "__main__":
    unittest.main()
