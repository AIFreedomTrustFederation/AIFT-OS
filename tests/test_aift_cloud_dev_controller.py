"""Offline contract tests for the guarded AIFT cloud development controller."""
from pathlib import Path
import importlib.util
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location(
    "aift_cloud_dev_controller", ROOT / "scripts/aift_cloud_dev_controller.py"
)
controller = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(controller)
WORKFLOW = ROOT / ".github/workflows/aift-cloud-dev-controller.yml"


class FakeAPI:
    def __init__(self, responses):
        self.responses = responses
        self.calls = []

    def get(self, path):
        self.calls.append(path)
        result = self.responses[path]
        if isinstance(result, Exception):
            raise result
        return result


class CloudControllerContract(unittest.TestCase):
    def test_allowlist_rejects_invalid_and_duplicate_repos(self):
        for invalid in ("", "AIFT-OS,AIFT-OS", "../private", "bad repo", ",".join(["a"] * 13)):
            with self.subTest(invalid=invalid), self.assertRaises(ValueError):
                controller.repo_names(invalid)
        self.assertEqual(controller.repo_names("VPS,AIFT-OS"), ["VPS", "AIFT-OS"])

    def test_report_is_exact_head_and_never_claims_merge_ready(self):
        sha = "a" * 40
        base = "/repos/AIFreedomTrustFederation/VPS"
        api = FakeAPI({
            base: {"full_name": "AIFreedomTrustFederation/VPS", "archived": False},
            base + "/pulls?state=open&per_page=8": [
                {"number": 5, "head": {"sha": sha}, "draft": False}
            ],
            base + f"/commits/{sha}/check-runs?per_page=100": {
                "check_runs": [{"status": "completed", "conclusion": "success"}],
                "total_count": 1,
            },
        })
        report = controller.audit(api, ["VPS"], now="2026-11-15T12:00:00+00:00")
        pr = report["repositories"][0]["pull_requests"][0]
        self.assertEqual(pr["head_sha"], sha)
        self.assertEqual(pr["checks"], "reported_checks_completed")
        self.assertFalse(pr["merge_readiness_verified"])
        self.assertFalse(pr["required_checks_verified"])
        self.assertIn(f"/commits/{sha}/check-runs", api.calls[2])

    def test_unavailable_evidence_fails_closed(self):
        sha = "b" * 40
        base = "/repos/AIFreedomTrustFederation/AIFT-OS"
        api = FakeAPI({
            base: {"full_name": "AIFreedomTrustFederation/AIFT-OS"},
            base + "/pulls?state=open&per_page=8": [{"number": 1, "head": {"sha": sha}}],
            base + f"/commits/{sha}/check-runs?per_page=100": controller.ApiError("HTTP 403"),
        })
        pr = controller.audit(api, ["AIFT-OS"], now="now")["repositories"][0]["pull_requests"][0]
        self.assertEqual(pr["checks"], "unavailable")
        self.assertFalse(pr["merge_readiness_verified"])

    def test_failed_and_incomplete_checks_are_not_green(self):
        self.assertEqual(controller.reported_checks([]), "no_checks_observed")
        self.assertEqual(controller.reported_checks([{"status": "queued"}]), "checks_in_progress")
        self.assertEqual(controller.reported_checks([{"status": "completed", "conclusion": "failure"}]), "reported_checks_failed")

    def test_no_dirty_worktree_formatting(self):
        with patch.object(controller.subprocess, "check_output", return_value=b" M internal/test.go\n"):
            with self.assertRaisesRegex(RuntimeError, "dirty checkout"):
                controller.format_tests()

    def test_workflow_uses_app_only_on_main_and_never_merges(self):
        workflow = WORKFLOW.read_text(encoding="utf-8")
        self.assertIn("github.ref == 'refs/heads/main'", workflow)
        self.assertIn("AIFT_AUTODEV_AUTO_PROPOSE == 'true'", workflow)
        self.assertIn("inputs.propose_routine_fixes == true", workflow)
        self.assertIn("permission-pull-requests: write", workflow)
        self.assertIn("--draft", workflow)
        self.assertIn("persist-credentials: false", workflow)
        self.assertNotIn("gh pr merge", workflow)
        self.assertNotIn("git push origin main", workflow)
        self.assertNotIn("pull_request_target:", workflow)
        self.assertIn("permission-contents: write", workflow)

    def test_output_report_is_non_authoritative(self):
        sample = {"observed_at_utc": "now", "repositories": [{
            "repository": "AIFreedomTrustFederation/VPS", "status": "unverified",
            "open_prs_sampled": 0, "pull_requests": []
        }]}
        md = controller.markdown_report(sample)
        self.assertIn("does not establish merge readiness", md)
        self.assertIn("unverified", md)


if __name__ == "__main__":
    unittest.main()
