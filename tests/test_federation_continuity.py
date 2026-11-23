"""Offline regression contract for the side-effect-free continuity guardian."""
import unittest
from dataclasses import replace

from federation_automation.continuity import Candidate, Evidence, assess

A = "a" * 40
B = "b" * 40


class GuardianTests(unittest.TestCase):
    def setUp(self):
        self.candidate = Candidate(
            "AIFreedomTrustFederation/AIFT-OS", A, B,
            frozenset({"documentation"}), ("docs/guide.md",),
        )
        self.evidence = Evidence(A, B, {"verify": "success", "test": "success"},
                                 True, True, True, True, True)

    def test_clean_routine_evidence_is_eligible(self):
        result = assess(self.candidate, self.evidence)
        self.assertTrue(result.eligible)
        self.assertEqual(result.risk, "routine")

    def test_missing_or_nonpassing_checks_fail_closed(self):
        for checks in ({}, {"verify": "pending"}, {"verify": "failure"}, {"verify": "success", "test": "skipped"}):
            with self.subTest(checks=checks):
                self.assertIn("required_checks_not_verified",
                              assess(self.candidate, replace(self.evidence, required_checks=checks)).reasons)

    def test_missing_each_independent_gate_blocks(self):
        fields = ("required_reviews_satisfied", "branch_protection_verified",
                  "mergeable", "recovery_verified", "evidence_trusted")
        for field in fields:
            with self.subTest(field=field):
                self.assertFalse(assess(self.candidate, replace(self.evidence, **{field: False})).eligible)

    def test_exact_head_and_base_required(self):
        for field in ("observed_head_sha", "observed_base_sha"):
            with self.subTest(field=field):
                self.assertIn("stale_commit_evidence",
                              assess(self.candidate, replace(self.evidence, **{field: "c" * 40})).reasons)

    def test_sensitive_paths_override_routine_label(self):
        for path in (".github/workflows/ci.yml", "deploy/app.yml", "internal/uxi/inference.go",
                     "federation_automation/core.py", "AGENTS.md", "docs/security-policy.md"):
            with self.subTest(path=path):
                result = assess(replace(self.candidate, changed_paths=(path,)), self.evidence)
                self.assertEqual(result.risk, "high")
                self.assertFalse(result.eligible)

    def test_sensitive_category_blocks(self):
        for category in ("security", "trust", "identity", "recovery", "deployment"):
            with self.subTest(category=category):
                self.assertFalse(assess(replace(self.candidate, categories=frozenset({category})),
                                        self.evidence).eligible)

    def test_unclassified_and_bad_paths_fail(self):
        cases = [
            replace(self.candidate, categories=frozenset()),
            replace(self.candidate, categories=frozenset({"unknown"})),
            replace(self.candidate, changed_paths=("../escape",)),
            replace(self.candidate, changed_paths=()),
            replace(self.candidate, head_sha="short"),
            replace(self.candidate, repository="OtherOrg/repo"),
        ]
        for candidate in cases:
            with self.subTest(candidate=candidate):
                self.assertFalse(assess(candidate, self.evidence).eligible)

    def test_no_external_actions_or_mutations(self):
        before = (self.candidate, self.evidence)
        assess(*before)
        self.assertEqual(before, (self.candidate, self.evidence))


if __name__ == "__main__":
    unittest.main()
