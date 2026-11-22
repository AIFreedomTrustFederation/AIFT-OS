#!/usr/bin/env python3
"""Read-only exact-head GitHub PR evidence, never an approval or merge oracle.

AIFT-OS's GITHUB_TOKEN is used only for its own repository. Other public
federation repositories are queried anonymously, with errors reported as
unavailable rather than bypassing authentication or inventing success.
"""
from __future__ import annotations

import argparse
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import re
from typing import Any, Callable
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

OWNER = "AIFreedomTrustFederation"
REPOSITORIES = (
    f"{OWNER}/AIFT-OS",
    f"{OWNER}/www.aifreedomtrust.com",
    f"{OWNER}/AIFT-Runtime",
    f"{OWNER}/VPS",
)
SHA = re.compile(r"^[0-9a-f]{40}$")
SENSITIVE = (
    "security", "identity", "trust", "governance", "constitution",
    "policy", "secret", "credential", "token", "auth", "release",
    "deploy", "signing", "compliance", "financial", "custody", "legal",
    "private", "safety", "workflow", "ruleset", "branch-protection",
)
FAILED = {"failure", "timed_out", "cancelled", "startup_failure"}
EXPECTED_CHECKS = {f"{OWNER}/AIFT-OS": ("test",)}


class ProbeUnavailable(RuntimeError):
    pass


def read_api(path: str, *, repository: str, token: str = "") -> Any:
    """Only GitHub REST, only a statically enumerated repository."""
    if repository not in REPOSITORIES or not path.startswith(f"/repos/{repository}/"):
        raise ProbeUnavailable("Unapproved GitHub API route")
    headers = {
        "Accept": "application/vnd.github+json",
        "User-Agent": "aift-read-only-pr-evidence",
        "X-GitHub-Api-Version": "2022-11-28",
    }
    if token and repository == f"{OWNER}/AIFT-OS":
        headers["Authorization"] = f"Bearer {token}"
    req = Request("https://api.github.com" + path, headers=headers, method="GET")
    try:
        with urlopen(req, timeout=15) as response:
            payload = response.read(1500001)
        if len(payload) > 1500000:
            raise ProbeUnavailable("GitHub evidence response exceeds size limit")
        return json.loads(payload)
    except HTTPError as error:
        raise ProbeUnavailable(f"GitHub read HTTP {error.code}") from error
    except (URLError, TimeoutError, ValueError, OSError) as error:
        raise ProbeUnavailable("GitHub read unavailable or malformed") from error


def classify(paths: list[str], *, changed_files: int) -> str:
    """Conservative path-based TRIAGE, not a semantic security review."""
    if not paths or changed_files != len(paths) or changed_files > 100:
        return "unknown_requires_review"
    for path in paths:
        lower = path.casefold()
        parts = lower.replace("\\", "/").split("/")
        if lower.startswith(".github/") or any(word in lower for word in SENSITIVE):
            return "restricted_review"
        if parts[0] in {"federation_automation", "internal", "cmd", "src"}:
            return "engineering_review"
    if all(path.startswith("docs/") and path.endswith(".md") for path in paths):
        return "routine_docs_candidate"
    if all(path.startswith("tests/") for path in paths):
        return "routine_tests_candidate"
    return "engineering_review"


def evaluate(
    repository: str, pr: dict[str, Any],
    get: Callable[[str], Any],
) -> dict[str, Any]:
    """Inspect source head, files, CI, Action approval, review, and ref freshness."""
    number = pr.get("number")
    if type(number) is not int or not 1 <= number <= 1000000000:
        raise ProbeUnavailable("Invalid pull request number")
    base = f"/repos/{repository}"
    detail = get(f"{base}/pulls/{number}")
    sha = detail.get("head", {}).get("sha")
    if not isinstance(sha, str) or not SHA.fullmatch(sha):
        raise ProbeUnavailable("Missing exact pull request head SHA")
    count = detail.get("changed_files")
    if type(count) is not int or count < 1 or count > 100:
        paths = []
    else:
        files = get(f"{base}/pulls/{number}/files?per_page=100")
        if not isinstance(files, list):
            raise ProbeUnavailable("Changed-file evidence malformed")
        paths = [row.get("filename", "") for row in files]
        if not all(isinstance(path, str) and path for path in paths):
            raise ProbeUnavailable("Invalid changed-file paths")
    checks_response = get(f"{base}/commits/{sha}/check-runs?per_page=100")
    workflow_response = get(f"{base}/actions/runs?head_sha={sha}&per_page=100")
    reviews = get(f"{base}/pulls/{number}/reviews?per_page=100")
    if (not isinstance(checks_response, dict) or
            not isinstance(workflow_response, dict) or not isinstance(reviews, list)):
        raise ProbeUnavailable("GitHub CI or review evidence malformed")
    checks = checks_response.get("check_runs")
    workflows = workflow_response.get("workflow_runs")
    if not isinstance(checks, list) or not isinstance(workflows, list):
        raise ProbeUnavailable("GitHub check-run or workflow evidence missing")
    if checks_response.get("total_count", 0) > 100 or workflow_response.get("total_count", 0) > 100:
        raise ProbeUnavailable("CI evidence truncated; cannot conclude readiness")
    for check in checks:
        if check.get("head_sha") not in (None, sha):
            raise ProbeUnavailable("Check belongs to a different commit")
    workflows = [w for w in workflows if w.get("head_sha") == sha]
    if len(reviews) == 100:
        raise ProbeUnavailable("Review pagination limit reached")
    latest_reviews: dict[str, str] = {}
    for review in reviews:
        actor = review.get("user") or {}
        login = actor.get("login")
        if isinstance(login, str) and actor.get("type") != "Bot":
            latest_reviews[login] = review.get("state", "")
    human_approvals = sum(state == "APPROVED" for state in latest_reviews.values())
    failed = sorted(set(
        [c.get("name", "unnamed") for c in checks if c.get("conclusion") in FAILED] +
        [w.get("name", "unnamed") for w in workflows if w.get("conclusion") in FAILED]
    ))
    action_required = sorted(set(
        w.get("name", "unnamed") for w in workflows if w.get("conclusion") == "action_required"
    ))
    expected = EXPECTED_CHECKS.get(repository)
    required_check_state = "not_configured"
    if expected:
        matched = [c for c in checks if c.get("name") in expected]
        if len(matched) != len(expected):
            required_check_state = "missing_or_pending"
        elif all(c.get("status") == "completed" and c.get("conclusion") == "success" for c in matched):
            required_check_state = "observed_success"
        elif any(c.get("conclusion") in FAILED for c in matched):
            required_check_state = "observed_failure"
        else:
            required_check_state = "pending_or_unverified"
    current = get(f"{base}/pulls/{number}")
    if current.get("head", {}).get("sha") != sha:
        raise ProbeUnavailable("PR head advanced during evidence collection")
    risk = classify(paths, changed_files=count)
    if action_required:
        gate = "workflow_approval_required"
    elif failed or required_check_state == "observed_failure":
        gate = "checks_failed"
    elif required_check_state in {"missing_or_pending", "pending_or_unverified"}:
        gate = "ci_not_verified"
    elif current.get("draft"):
        gate = "draft_requires_review"
    elif risk != "routine_docs_candidate" and risk != "routine_tests_candidate":
        gate = "additional_risk_review"
    else:
        gate = "human_or_policy_review"
    return {
        "number": number,
        "url": current.get("html_url") or f"https://github.com/{repository}/pull/{number}",
        "head_sha": sha,
        "base_sha": (current.get("base") or {}).get("sha"),
        "draft": bool(current.get("draft")),
        "mergeable": current.get("mergeable"),  # None = unverified, never "ready"
        "risk_triage": risk,
        "changed_paths": paths[:100],
        "human_approvals_observed": human_approvals,
        "expected_check_state": required_check_state,
        "workflow_approval_required": action_required,
        "failing_checks": failed,
        "check_run_count": len(checks),
        "gate": gate,
        "merge_readiness": "not_authorized_or_verified",
        "auto_merge_eligible": False,
    }


def collect(
    repositories: tuple[str, ...],
    get: Callable[[str, str], Any],
    *, max_per_repo: int = 10,
) -> dict[str, Any]:
    if not 1 <= max_per_repo <= 20:
        raise ValueError("Bounded per-repository limit required")
    report: dict[str, Any] = {
        "schema_version": 1,
        "observed_at_utc": datetime.now(timezone.utc).isoformat(),
        "mode": "read_only_evidence",
        "merge_authority": "none",
        "repositories": [],
    }
    for repo in repositories:
        entry: dict[str, Any] = {"repository": repo, "pull_requests": [], "truncated": False}
        report["repositories"].append(entry)
        if repo not in REPOSITORIES:
            entry["error"] = "Repository outside explicit read-only allowlist"
            continue
        prefix = f"/repos/{repo}"
        try:
            branch = get(repo, f"{prefix}/branches/main")
            entry["main_sha"] = (branch.get("commit") or {}).get("sha")
            entry["main_protected_observed"] = branch.get("protected")
            listed = get(repo, f"{prefix}/pulls?state=open&sort=updated&direction=desc&per_page={max_per_repo + 1}")
            if not isinstance(listed, list):
                raise ProbeUnavailable("Open PR listing malformed")
            entry["truncated"] = len(listed) > max_per_repo
            for pr in listed[:max_per_repo]:
                try:
                    entry["pull_requests"].append(evaluate(repo, pr, lambda path: get(repo, path)))
                except (ProbeUnavailable, KeyError, TypeError, AttributeError, ValueError) as error:
                    entry["pull_requests"].append({
                        "number": pr.get("number"),
                        "gate": "evidence_unavailable",
                        "error": str(error)[:160],
                        "merge_readiness": "not_authorized_or_verified",
                        "auto_merge_eligible": False,
                    })
        except (ProbeUnavailable, KeyError, TypeError, AttributeError) as error:
            entry["error"] = str(error)[:160]
    return report


def markdown(report: dict[str, Any]) -> str:
    lines = [
        "# AIFT read-only PR evidence",
        "",
        f"Observed: {report['observed_at_utc']}",
        "",
        "This is **not** merge authorization. Missing or stale evidence is never success.",
        "",
        "| Repository | PR | Head | Path risk triage | Verified gate |",
        "|---|---|---|---|---|",
    ]
    for repo in report["repositories"]:
        name = repo["repository"]
        if repo.get("error"):
            lines.append(f"| {name} | — | — | unverified | {repo['error']} |")
            continue
        for pr in repo["pull_requests"]:
            n = pr.get("number")
            gate = pr["gate"]
            if gate == "evidence_unavailable":
                lines.append(f"| {name} | #{n} | unknown | unverified | evidence_unavailable |")
                continue
            lines.append(
                f"| {name} | [#{n}]({pr['url']}) | "
                f"{pr['head_sha'][:12]} | {pr['risk_triage']} | {gate} |"
            )
        if repo["truncated"]:
            lines.append(f"| {name} | — | — | unverified | list truncated |")
        if not repo["pull_requests"] and not repo.get("error"):
            lines.append(f"| {name} | — | — | — | no open PRs in sampled page |")
    return "\n".join(lines) + "\n"


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repository", choices=("all",) + REPOSITORIES, default="all")
    parser.add_argument("--max-per-repo", type=int, default=10)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    selected = REPOSITORIES if args.repository == "all" else (args.repository,)
    token = os.environ.get("AIFT_AUDIT_TOKEN", "")
    result = collect(selected, lambda repo, path: read_api(path, repository=repo, token=token),
                     max_per_repo=args.max_per_repo)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    summary = markdown(result)
    print(summary, end="")
    summary_path = os.environ.get("GITHUB_STEP_SUMMARY")
    if summary_path:
        with open(summary_path, "a", encoding="utf-8") as stream:
            stream.write(summary)


if __name__ == "__main__":
    main()
