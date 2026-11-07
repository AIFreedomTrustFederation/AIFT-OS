#!/usr/bin/env python3
"""Audit and optionally reconcile AIFT repository-level required-check rulesets.

Uses only Python stdlib. Never applies changes unless --apply is supplied.
Requires GH_TOKEN with Administration:write on target repositories.
"""
import argparse
import json
import os
import sys
import urllib.error
import urllib.request

OWNER = "AIFreedomTrustFederation"
GROUPS = {
    "validate": ["www.aifreedomtrust.com", "AIFT-Runtime", "AIFT-Genesis", "TheMindofAll", "AI-Freedom-Trust", "mobox"],
    "verify": ["BookSmith-Federation-OS", "AetherianGovernance", "aifreedomtrustfederation.github.io", "capital-city-provisions"],
    "qa": ["AIFT-Forge", "biozone-harmony-boost"],
    "test": ["AIFT-OS"],
    "local-qa": ["VPS"],
    "contracts": ["OpenMontage"],
    "build": ["booksmith-ai"],
    "Test": ["Aether_Coin"],
}
RULE_PREFIX = "aift-required-pr-check"
EXCLUDED = {"chktex", "c-848263", "tastycutz", "repo-brainstorm-backend-forge"}

def api(method, path, payload=None):
    token = os.environ.get("GH_TOKEN")
    if not token:
        raise RuntimeError("GH_TOKEN is required")
    data = None if payload is None else json.dumps(payload).encode()
    req = urllib.request.Request("https://api.github.com" + path, data=data, method=method,
        headers={"Authorization": "Bearer " + token, "Accept": "application/vnd.github+json",
                 "X-GitHub-Api-Version": "2022-11-28", "User-Agent": "aift-ruleset-controller"})
    try:
        with urllib.request.urlopen(req, timeout=25) as res:
            return json.load(res)
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode(errors="replace")[:500]
        raise RuntimeError(f"{method} {path}: HTTP {exc.code}: {detail}") from exc

def preflight(repo, check):
    base = f"/repos/{OWNER}/{repo}"
    metadata = api("GET", base)
    if metadata.get("private"):
        raise RuntimeError("Private repository excluded from this policy")
    if metadata.get("archived"):
        raise RuntimeError("Archived repository")
    if metadata.get("allow_auto_merge") is not True:
        raise RuntimeError("Enable Allow auto-merge in repository settings first")
    if metadata.get("allow_squash_merge") is not True:
        raise RuntimeError("Enable squash merging in repository settings first")
    branch = metadata["default_branch"]
    prs = api("GET", base + "/pulls?state=all&per_page=100")
    eligible = [p for p in prs if p.get("base", {}).get("ref") == branch and not p.get("draft")]
    # PR-head evidence is mandatory. Never infer check availability from main branch.
    found = False
    for pr in eligible[:30]:
        sha = pr["head"]["sha"]
        checks = api("GET", base + f"/commits/{sha}/check-runs?per_page=100")
        names = {c["name"] for c in checks.get("check_runs", [])}
        statuses = api("GET", base + f"/commits/{sha}/status")
        names.update(s["context"] for s in statuses.get("statuses", []))
        if check in names:
            found = True
            break
    if not found:
        raise RuntimeError(f"No PR-head evidence for exact check {check!r}; manual validation required")
    return branch, base

def reconcile(repo, check, apply):
    branch, base = preflight(repo, check)
    name = f"{RULE_PREFIX}-{check}"
    rules = api("GET", base + "/rulesets?per_page=100")
    matching = [r for r in rules if r.get("name") == name]
    if len(matching) > 1:
        raise RuntimeError("Duplicate managed rulesets: refusing to proceed")
    payload = {
        "name": name, "target": "branch", "enforcement": "active",
        "conditions": {"ref_name": {"include": ["~DEFAULT_BRANCH"], "exclude": []}},
        "rules": [
            {"type": "pull_request", "parameters": {
                "dismiss_stale_reviews_on_push": False,
                "require_code_owner_review": False,
                "require_last_push_approval": False,
                "required_approving_review_count": 0,
                "required_review_thread_resolution": True}},
            {"type": "required_status_checks", "parameters": {
                "strict_required_status_checks_policy": True,
                "do_not_enforce_on_create": False,
                "required_status_checks": [{"context": check}]}}
        ],
        "bypass_actors": []
    }
    existing = matching[0] if matching else None
    print(f"{repo}: default={branch}, check={check}, action={'update' if existing else 'create'}, apply={apply}")
    if not apply:
        return
    if existing:
        # Avoid accidental overwrite of human-edited policy: require manual review.
        current = api("GET", base + f"/rulesets/{existing['id']}")
        comparable = {k: current.get(k) for k in payload}
        if comparable != payload:
            raise RuntimeError("Managed ruleset already exists with different policy; review manually")
        print("  already matches desired policy")
    else:
        api("POST", base + "/rulesets", payload)
        print("  created")

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--apply", action="store_true", help="Explicitly create rulesets after preflight")
    parser.add_argument("--repo", help="Limit to one repository")
    args = parser.parse_args()
    desired = [(repo, check) for check, repos in GROUPS.items() for repo in repos]
    if args.repo:
        desired = [(r, c) for r, c in desired if r == args.repo]
        if not desired:
            parser.error("Unknown or excluded repository")
    failures = 0
    for repo, check in desired:
        try:
            reconcile(repo, check, args.apply)
        except Exception as exc:
            failures += 1
            print(f"BLOCKED {repo}: {exc}", file=sys.stderr)
    print(f"Checked {len(desired)} repositories; blocked {failures}")
    return 1 if failures else 0

if __name__ == "__main__":
    sys.exit(main())
