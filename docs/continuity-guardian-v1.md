# Continuity Guardian V1: evidence-only merge eligibility

The guardian is a pure decision function in `federation_automation/continuity.py`.
It performs no GitHub writes, deploys, merges, secrets access or model inference.
It does **not** change the existing worker queue or enable unattended merging.

## Contract

Supply a candidate with exact 40-character lowercase Git commit IDs, repository,
changed paths, and independently assigned change categories. Supply independently
verified exact-head and base evidence, complete required-check conclusions,
review compliance, branch-rule verification, mergeability, trusted provenance,
and recovery readiness. Unknown or missing evidence fails closed.

Only narrowly classified routine documentation/formatting/test changes may be
eligible. Sensitive categories and protected code/configuration paths are
automatically escalated. An eligible decision is **not** merge authorization:
the executor must re-read all live GitHub protections and exact commit IDs at
the moment of mutation, and must not bypass checks or approvals.

## Verification

```sh
python3 -B -m unittest discover -s tests -p 'test_federation_continuity.py' -v
```

This first slice deliberately does not add an executor. Next slice: integrate
trusted evidence collection and the read-only decision into the existing
SQLite worker queue, with adversarial tests for stale commits, partial check
coverage, branch rules, race conditions, and restart recovery. Only then
consider opt-in low-risk merging after independent review. Deployments,
credentials, trust, safety, and recovery changes always remain separate.
