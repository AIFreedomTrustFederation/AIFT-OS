"""Fail-closed, side-effect-free continuity decision for engineering promotions.

This module does not authenticate, merge, deploy, or infer GitHub branch protection.
Evidence is supplied by an independently verified collector, never by the candidate.
"""
from dataclasses import dataclass
from typing import Mapping, Sequence

SENSITIVE = frozenset({
    "security", "identity", "trust", "governance", "deployment",
    "infrastructure", "destructive", "safety", "credentials", "recovery",
})
ROUTINE = frozenset({"documentation", "formatting", "tests"})
SHA_HEX = frozenset("0123456789abcdef")


@dataclass(frozen=True)
class Candidate:
    repository: str
    head_sha: str
    base_sha: str
    categories: frozenset[str]
    changed_paths: tuple[str, ...]


@dataclass(frozen=True)
class Evidence:
    observed_head_sha: str
    observed_base_sha: str
    required_checks: Mapping[str, str]
    required_reviews_satisfied: bool
    branch_protection_verified: bool
    mergeable: bool
    recovery_verified: bool
    evidence_trusted: bool


@dataclass(frozen=True)
class Decision:
    eligible: bool
    risk: str
    reasons: tuple[str, ...]


def _sha(value: str) -> bool:
    return isinstance(value, str) and len(value) == 40 and all(c in SHA_HEX for c in value)


def assess(candidate: Candidate, evidence: Evidence) -> Decision:
    """Determine eligibility for a *merge*, not authorization to deploy.

    A true result only means the supplied evidence satisfies these local rules.
    The executor MUST independently re-read exact-head checks, branch rules,
    approvals, mergeability and base SHA immediately before any GitHub mutation.
    """
    reasons: list[str] = []
    if not candidate.repository.startswith("AIFreedomTrustFederation/") or candidate.repository.count("/") != 1:
        reasons.append("repository_out_of_scope")
    if not _sha(candidate.head_sha) or not _sha(candidate.base_sha):
        reasons.append("invalid_commit_identity")
    if not candidate.changed_paths or any(
        not isinstance(p, str) or not p or p.startswith("/") or
        any(part in ("", ".", "..") for part in p.split("/"))
        for p in candidate.changed_paths
    ):
        reasons.append("invalid_change_paths")
    categories = candidate.categories
    if not categories or not categories <= (SENSITIVE | ROUTINE):
        reasons.append("unclassified_change")
    # Category labels alone cannot make a sensitive path routine.
    sensitive_path = any(
        p.startswith((".github/", "deploy/", "scripts/", "internal/", "federation_automation/")) or
        p in ("AGENTS.md", "FEDERATION-PRIME-DIRECTIVE.md") or
        any(term in p.lower() for term in
            ("security", "identity", "trust", "governance", "secret", "credential", "policy", "auth", "release"))
        for p in candidate.changed_paths
    )
    risk = "high" if categories & SENSITIVE or sensitive_path else "routine"
    if risk == "high":
        reasons.append("human_review_required_for_sensitive_change")
    if candidate.head_sha != evidence.observed_head_sha or candidate.base_sha != evidence.observed_base_sha:
        reasons.append("stale_commit_evidence")
    if not evidence.evidence_trusted:
        reasons.append("untrusted_evidence")
    if not evidence.required_checks or any(v != "success" for v in evidence.required_checks.values()):
        reasons.append("required_checks_not_verified")
    if not evidence.required_reviews_satisfied:
        reasons.append("required_reviews_not_verified")
    if not evidence.branch_protection_verified:
        reasons.append("branch_protection_not_verified")
    if not evidence.mergeable:
        reasons.append("mergeability_not_verified")
    if not evidence.recovery_verified:
        reasons.append("recovery_not_verified")
    return Decision(eligible=not reasons, risk=risk, reasons=tuple(reasons))
