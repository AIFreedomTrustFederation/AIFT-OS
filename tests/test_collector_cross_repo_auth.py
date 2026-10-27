"""Offline regressions for the AIFT-OS collector's cross-repo authorization boundary.

Passing these static tests is not proof that an App is installed or that a
trusted default-branch run can retrieve actual producer artifacts.
"""
from pathlib import Path
import re
import os
import subprocess
import tempfile
import textwrap
import unittest

ROOT = Path(__file__).resolve().parents[1]
WORKFLOW = ROOT / ".github/workflows/federation-evidence-collector.yml"
APP_ACTION = "actions/create-github-app-token@bcd2ba49218906704ab6c1aa796996da409d3eb1"


def live_job(text):
    return text.split("\n  live-producer-evidence:\n", 1)[1]


class CrossRepositoryAuthorizationContract(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.source = WORKFLOW.read_text(encoding="utf-8")
        cls.live = live_job(cls.source)

    def test_live_job_never_runs_on_pull_requests(self):
        self.assertIn("  push:\n    branches: [main]\n", self.source)
        self.assertIn("  pull_request:\n", self.source)
        self.assertIn("    if: github.event_name == 'push' || github.event_name == 'schedule'\n", self.live)
        self.assertNotIn("github.event.pull_request", self.live)

    def test_pinned_app_action_mints_for_exactly_two_producers(self):
        self.assertIn(f"uses: {APP_ACTION}", self.live)
        self.assertIn("id: producer-token", self.live)
        self.assertIn("owner: AIFreedomTrustFederation", self.live)
        self.assertRegex(self.live, r"(?m)^          repositories: \|\n            \.github\n            AIFT-Runtime\n")
        self.assertNotIn("skip-token-revoke: true", self.live)

    def test_app_credentials_are_explicit_not_repository_token(self):
        self.assertIn("client-id: ${{ vars.AIFT_FEDERATION_EVIDENCE_APP_CLIENT_ID }}", self.live)
        self.assertIn("private-key: ${{ secrets.AIFT_FEDERATION_EVIDENCE_APP_PRIVATE_KEY }}", self.live)
        self.assertEqual(self.live.count("GH_TOKEN: ${{ steps.producer-token.outputs.token }}"), 2)
        self.assertNotIn("GH_TOKEN: ${{ github.token }}", self.live)
        self.assertNotIn("GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}", self.live)

    def test_app_token_is_downscoped_to_read_permissions(self):
        self.assertIn("permission-actions: read", self.live)
        self.assertIn("permission-contents: read", self.live)
        self.assertNotRegex(self.live, r"(?m)^\s*permission-[a-z-]+: write\s*$")

    def test_missing_token_and_unavailable_producer_api_fail_closed(self):
        self.assertIn('if [[ -z "$GH_TOKEN" ]]; then', self.live)
        self.assertIn('actual=$(gh api "repos/${repo}" --jq \'.full_name\')', self.live)
        self.assertIn('if [[ "$actual" != "$repo" ]]; then', self.live)
        self.assertIn('gh api "repos/${repo}/actions/workflows?per_page=1"', self.live)
        self.assertIn("set -euo pipefail", self.live)
        self.assertIn("for repo in AIFreedomTrustFederation/.github AIFreedomTrustFederation/AIFT-Runtime; do", self.live)

    def _execute_preflight(self, token, deny=None):
        # Execute the exact shell script from the workflow with a synthetic gh.
        block = self.live.split(
            "      - name: Fail closed if producer Actions APIs are not accessible\n", 1
        )[1].split("      - name: Verify live producer artifacts", 1)[0]
        script = textwrap.dedent(block.split("        run: |\n", 1)[1])
        with tempfile.TemporaryDirectory() as temp:
            fake = Path(temp) / "gh"
            fake.write_text(r"""#!/usr/bin/env bash
path="$2"
if [[ -n "${FAKE_DENY_REPO:-}" && "$path" == *"$FAKE_DENY_REPO"* ]]; then
  exit 22
fi
if [[ "$path" == */actions/workflows\?per_page=1 ]]; then
  echo 1
elif [[ "$path" == repos/* ]]; then
  echo "${path#repos/}"
else
  exit 23
fi
""", encoding="utf-8")
            fake.chmod(0o755)
            env = dict(os.environ, PATH=f"{temp}:{os.environ['PATH']}", GH_TOKEN=token)
            if deny:
                env['FAKE_DENY_REPO'] = deny
            return subprocess.run(["bash", "-c", script], env=env,
                                  text=True, capture_output=True, timeout=5)

    def test_actual_preflight_rejects_empty_token(self):
        result = self._execute_preflight("")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("No GitHub App installation token", result.stderr)

    def test_actual_preflight_rejects_denied_producer(self):
        result = self._execute_preflight("synthetic-not-a-secret", "AIFT-Runtime")
        self.assertNotEqual(result.returncode, 0)

    def test_actual_preflight_accepts_both_accessible_producers(self):
        result = self._execute_preflight("synthetic-not-a-secret")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_token_is_only_in_trusted_job(self):
        offline_job = self.source.split("\n  collector:\n", 1)[1].split("\n  live-producer-evidence:\n", 1)[0]
        self.assertNotIn("secrets.AIFT_FEDERATION_EVIDENCE_APP_PRIVATE_KEY", offline_job)
        self.assertNotIn("steps.producer-token.outputs.token", offline_job)
        self.assertIn("Test cross-repository token scope contract", offline_job)


if __name__ == "__main__":
    unittest.main()
