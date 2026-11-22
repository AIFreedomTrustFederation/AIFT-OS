# AIFT read-only federation pull request evidence

This capability collects **observed**, bounded, exact-head GitHub evidence. It is
not a reviewer, merge-authority engine, deployment agent, or a production
guarantee. Preserve AIFT's roots-as-repositories model without inferring live
federation health from static PR metadata.

## Scope and cadence

The workflow `.github/workflows/aift-pr-evidence-audit.yml` is triggered on
its own PR changes for **offline tests only**. Once the workflow is reviewed
and merged onto `main`, it will also run on a six-hour schedule and by
operator-initiated `workflow_dispatch`. Scheduled audits sample at most ten
recent open PRs per configured repository, with a visible `truncated` flag.

The initially enumerated public repositories are AIFT-OS,
www.aifreedomtrust.com, AIFT-Runtime, and VPS. This is not a claim that these
are the complete federation or that the GitHub token has cross-repository
access. Only AIFT-OS receives the repository-scoped Actions read token.
Other public repositories are queried anonymously; missing or rate-limited
data are reported as `evidence_unavailable`, never assumed passing.

For each sampled PR, the evidence includes the exact source head, base SHA,
file-based risk triage, current draft state, raw check-run conclusions,
workflow `action_required` status, non-bot GitHub review counts, and
uncertain or failed verification gates. A second PR fetch detects a changed
head during sampling. The collector does not verify every branch rule,
CODEOWNERS policy, independent model evaluation, dependency, browser
journey, or the semantic safety of a change.

**Every record returns `auto_merge_eligible: false` and
`merge_readiness: not_authorized_or_verified` by design.** Path-based
classifications are triage only. Do not treat a passing `test` check or a
human-approval count as authority to bypass review, rulesets, or the
operator's higher-risk approval boundaries.

## Run manually after a reviewed merge

On GitHub, use Actions → **AIFT Read-Only Federation PR Evidence** →
Run workflow, selecting `all` or one listed repository.

From a protected operator's Ubuntu shell with `gh` already authenticated:

```bash
gh workflow run aift-pr-evidence-audit.yml \
  --repo AIFreedomTrustFederation/AIFT-OS --ref main -f repository=all

gh run list --repo AIFreedomTrustFederation/AIFT-OS \
  --workflow aift-pr-evidence-audit.yml --limit 5
```

The completed audit run uploads `aift-pr-evidence` JSON and writes a Markdown
summary. Any failures or unavailable remote evidence must remain visible.
A workflow on an unmerged branch is not an active periodic service.

## GitHub Copilot PR #90 manual workflow approval

PR #90 is a separate Copilot-generated one-line change in
`docs/TESTING.md`; the controller is not its author. On its exact
`f31d588422ada3e68646496d00d549173fd9f6fb` head, three workflow
runs initially returned `action_required`, with zero jobs executed.
**Moving a PR from draft to ready-for-review does not approve these runs.**
An authorized repository operator must inspect the change and use the PR's
**Approve and run workflows** control. This is not a CI pass or merge. Do
not auto-approve arbitrary contributors or disable approval requirements as
a workaround without a separate security decision.

## PR lifecycle and sensitive actions

- Routine documentation/test changes: collect evidence, suggest edits,
  run checks; propose native auto-merge only via a future separately reviewed
  policy with explicit labels and real branch-rule verification.
- UXI/architecture changes: open draft PRs, verify dependencies and real
  browser acceptance, escalate unresolved checks.
- Security, governance, trust, identity, signing, credentials, destruction
  and deployment: require separate operator authorization.
- Never mint a write token, update settings, merge, alter a deployment,
  register a runner, or expose private state from this collector.

The Android Aetherion loopback test and local read-only poller are independent
of this hosted metadata auditor. Neither proves that a local AI-generated
plan has been published. The first Aetherion-controlled draft must be verified
separately with its exact reviewed digest and a protected operator handoff.
