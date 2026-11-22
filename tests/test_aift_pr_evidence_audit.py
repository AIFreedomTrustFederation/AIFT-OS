"""Offline proof that federation PR audits are read-only, bounded and fail closed."""
import importlib.util
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location("aift_pr_evidence_audit", ROOT / "scripts/aift_pr_evidence_audit.py")
audit = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(audit)
REPO = "AIFreedomTrustFederation/AIFT-OS"
SHA = "a" * 40


def fixture(*, action_required=False, head=SHA, protected=False):
    prefix = "/repos/" + REPO
    detail = {
        "number": 90, "head": {"sha": head},
        "base": {"sha": "b" * 40}, "changed_files": 1,
        "draft": False, "mergeable": True,
        "html_url": "https://github.com/" + REPO + "/pull/90",
    }
    return {
        prefix + "/branches/main": {"commit": {"sha": "b" * 40}, "protected": protected},
        prefix + "/pulls?state=open&sort=updated&direction=desc&per_page=11": [{"number": 90}],
        prefix + "/pulls/90": detail,
        prefix + "/pulls/90/files?per_page=100": [{"filename": "docs/TESTING.md"}],
        prefix + "/commits/" + head + "/check-runs?per_page=100": {
            "total_count": 1,
            "check_runs": [{
                "name": "test", "head_sha": head,
                "status": "completed", "conclusion": "success",
            }],
        },
        prefix + "/actions/runs?head_sha=" + head + "&per_page=100": {
            "total_count": 1,
            "workflow_runs": [{
                "name": "CI", "head_sha": head,
                "conclusion": "action_required" if action_required else "success",
            }],
        },
        prefix + "/pulls/90/reviews?per_page=100": [],
    }


class AuditTests(unittest.TestCase):
    def test_routine_docs_are_not_automatically_merge_authorized(self):
        f = fixture()
        result = audit.collect((REPO,), lambda repo, path: f[path])
        entry = result["repositories"][0]
        self.assertFalse(entry["main_protected_observed"])
        pr = entry["pull_requests"][0]
        self.assertEqual(pr["risk_triage"], "routine_docs_candidate")
        self.assertEqual(pr["expected_check_state"], "observed_success")
        self.assertEqual(pr["gate"], "human_or_policy_review")
        self.assertFalse(pr["auto_merge_eligible"])
        self.assertEqual(pr["merge_readiness"], "not_authorized_or_verified")
        self.assertIn("#90", audit.markdown(result))

    def test_workflow_action_required_has_priority_over_green_check(self):
        f = fixture(action_required=True)
        entry = audit.collect((REPO,), lambda repo, path: f[path])["repositories"][0]
        pr = entry["pull_requests"][0]
        self.assertEqual(pr["gate"], "workflow_approval_required")
        self.assertEqual(pr["workflow_approval_required"], ["CI"])
        self.assertFalse(pr["auto_merge_eligible"])

    def test_missing_required_ci_never_turns_green(self):
        f = fixture()
        path = "/repos/" + REPO + "/commits/" + SHA + "/check-runs?per_page=100"
        f[path] = {"total_count": 0, "check_runs": []}
        pr = audit.collect((REPO,), lambda repo, path: f[path])["repositories"][0]["pull_requests"][0]
        self.assertEqual(pr["gate"], "ci_not_verified")

    def test_sensitive_paths_are_elevated(self):
        for path in (".github/workflows/ci.yml", "scripts/secret-helper.py",
                     "docs/TRUST.md", "docs/deployment.md", "docs/identity-contract.md"):
            with self.subTest(path=path):
                self.assertEqual(audit.classify([path], changed_files=1), "restricted_review")
        self.assertEqual(audit.classify(["src/cinematic.js"], changed_files=1), "engineering_review")
        self.assertEqual(audit.classify(["tests/test_docs.py"], changed_files=1),
                         "routine_tests_candidate")
        self.assertEqual(audit.classify(["docs/TESTING.md"], changed_files=2),
                         "unknown_requires_review")

    def test_invalid_or_unavailable_data_not_reported_successful(self):
        f = fixture()
        missing = "/repos/" + REPO + "/commits/" + SHA + "/check-runs?per_page=100"
        f.pop(missing)
        pr = audit.collect((REPO,), lambda repo, path: f[path])["repositories"][0]["pull_requests"][0]
        self.assertEqual(pr["gate"], "evidence_unavailable")
        self.assertFalse(pr["auto_merge_eligible"])

    def test_advanced_pr_head_fails_closed(self):
        f = fixture()
        calls = [0]
        def get(repo, path):
            if path.endswith("/pulls/90"):
                calls[0] += 1
                if calls[0] > 1:
                    return {**f[path], "head": {"sha": "c" * 40}}
            return f[path]
        pr = audit.collect((REPO,), get)["repositories"][0]["pull_requests"][0]
        self.assertEqual(pr["gate"], "evidence_unavailable")
        self.assertIn("advanced", pr["error"])

    def test_bounded_repos_and_no_token_to_other_repositories(self):
        with self.assertRaises(ValueError):
            audit.collect((REPO,), lambda repo, path: {}, max_per_repo=100)
        unknown = audit.collect(("OtherOwner/private",), lambda repo, path: {})
        self.assertEqual(unknown["repositories"][0]["error"],
                         "Repository outside explicit read-only allowlist")
        captured = []
        class Response:
            def __enter__(self): return self
            def __exit__(self, *args): return False
            def read(self, n): return b'{}'
        def opener(req, timeout):
            captured.append(req)
            return Response()
        class Opener:
            def open(self, req, timeout):
                return opener(req, timeout)
        with patch.object(audit, "build_opener", return_value=Opener()) as factory:
            audit.read_api("/repos/" + REPO + "/branches/main", repository=REPO, token="secret")
            external = audit.REPOSITORIES[1]
            audit.read_api("/repos/" + external + "/branches/main",
                           repository=external, token="secret")
        self.assertTrue(all(call.args == (audit.NoRedirects,) for call in factory.call_args_list))
        self.assertIsNone(audit.NoRedirects().redirect_request(None, None, 302, "",
                                                                {}, "https://untrusted.invalid/"))
        self.assertIn("Authorization", captured[0].headers)
        self.assertNotIn("Authorization", captured[1].headers)

    def test_no_writes_in_workflow_or_script(self):
        script = (ROOT / "scripts/aift_pr_evidence_audit.py").read_text()
        workflow = (ROOT / ".github/workflows/aift-pr-evidence-audit.yml").read_text()
        self.assertNotIn("pull_request_target:", workflow)
        self.assertNotIn("write-all", workflow)
        self.assertNotIn("permission-contents: write", workflow)
        self.assertNotIn("git push", script)
        self.assertNotIn("gh pr merge", script)
        self.assertIn("schedule:", workflow)
        self.assertIn("workflow_dispatch:", workflow)


if __name__ == "__main__":
    unittest.main()
