# AIFT Federation Automation: existing App worker

This branch implements a **phone-independent worker foundation** for the already registered **AIFT Federation Automation** GitHub App, ID **5861093**. It does **not** register a new App. This is code on an unmerged branch, **not a deployed or operational 24/7 service**. It complements the separate unmerged AIFT-OS controller PR #73 and preserves the existing read-only federation evidence collector.

## Implemented

- 8-minute RS256 JWTs signed from an owner-private PEM using OpenSSL (key contents never enter arguments or logs); authenticated `/app` identity must equal 5861093 and the installed account must equal AIFreedomTrustFederation.
- Short-lived installation tokens restricted to a single repository and exactly requested read or write scopes, with token response readback. No personal GitHub CLI token is used.
- Signature-verified HMAC-SHA256 webhook endpoint on **127.0.0.1 only**; deduplicated delivery and exact-head events; bot and non-default push filtering.
- SQLite WAL-backed durable jobs, bounded queue (300 active), four-attempt exponential retries, expired lease recovery, structured no-secret audit records, initial paused state, local pause/resume/status CLI, systemd example for independent hosting.
- Read-only operator console and evidence JSON at `http://127.0.0.1:8787/` and `/api/jobs` on an approved running host; live job records are displayed without inferred merge readiness. Remote administrative controls are intentionally absent. Commands can be copied for authorized host-shell execution.
- Read-only, bounded open-PR and exact-head check-run sampling. **Merge readiness is always unverified**: branch rules, review requirements, and conflicts are not independently evaluated.
- Opt-in deterministic whitespace-only formatting of narrowly allowed files under `docs/` into isolated branches and **draft PRs only**. Requires both `AIFT_ENABLE_DOC_PROPOSALS=1` and membership in `AIFT_WRITE_REPOS`; never auto-merges.

## Not yet implemented or verified

No AI model or autonomous arbitrary code-writing agent, remote administrative console/approvals, cross-repository dependency solver, browser acceptance, production deployment, live App authentication, or live webhook verification is present. A scheduled audit is not equivalent to an operational AI coding system. Exact-head CI and a real draft PR created by this worker require live verification. There is no uptime claim.

## Safe operator runbook

1. Review the branch, run the offline regression workflow, and verify full AIFT-OS canonical tests. Do not merge without review.
2. On an **approved independent host**, install Python 3.13+, OpenSSL, and a reviewed working tree at `/opt/aift/AIFT-OS`. Preserve existing files and key backups. The Android key is not accessible to this worker or ChatGPT; securely provision the already downloaded PEM using an approved method and set permissions to 0600. The revised `scripts/bootstrap_local_app_key.py --apply` searches accessible Downloads, cryptographically checks App ID 5861093, discovers Client ID, and requires `.github` plus `AIFT-Runtime` read-only producer API access before configuring the existing Actions evidence secret and variable. Pass `--downloads-dir` when Android storage is not automatically visible; no private key is pasted or uploaded to chat. Do not paste or upload it into chat.
3. Set `AIFT_APP_PRIVATE_KEY_FILE` to a private path and execute `python3 -m federation_automation.service check-auth`; review the returned App ID, Client ID, installation ID, and permission names. GitHub App metadata cannot be read by a personal token. If permissions are missing, obtain approval before changing the App configuration.
4. Configure `AIFT_WEBHOOK_SECRET_FILE` with a private random 32–256 byte secret, and an HTTPS reverse proxy forwarding authenticated webhook POSTs to the loopback-only `/webhook`. Do not expose `/healthz` publicly. Configure push and pull_request subscriptions in the existing App only after authorization.
5. Optionally adapt `deploy/aift-federation-worker.service` and manage the PEM and webhook secret using systemd credentials or another approved secret store. This file is an **example, not an installed service**. Set monitored restart policies, backup/encryption of the SQLite DB, alerting, TLS verification, and tested recovery procedures.
6. Initially the queue is **paused**. Resume through the local CLI, not a public endpoint. Use the same `--db` for service and controls:

```sh
python3 -m federation_automation.service --db /var/lib/aift-federation-automation/jobs.db status
python3 -m federation_automation.service --db /var/lib/aift-federation-automation/jobs.db resume
python3 -m federation_automation.service --db /var/lib/aift-federation-automation/jobs.db enqueue-audit --repo AIFreedomTrustFederation/AIFT-OS
python3 -m federation_automation.service --db /var/lib/aift-federation-automation/jobs.db pause
python3 -B -m unittest discover -s tests -p test_federation_automation_core.py -v
```

7. To opt in to one safe formatting test, explicitly configure `AIFT_ENABLE_DOC_PROPOSALS=1` and an appropriate `AIFT_WRITE_REPOS` allowlist, resume, then enqueue a `docs/<non-sensitive>.md` formatting job. No supported operation merges or deploys.

## Acceptance evidence still required

- [ ] `/app` authenticates as App 5861093 and installation ID/repository coverage is verified.
- [ ] Collector separately validates read-only `contents:read`/`actions:read` scopes on `.github` and `AIFT-Runtime`.
- [ ] Approved independent host remains running after Android disconnects.
- [ ] Webhook signature negative tests, delivery and dedup, recovery and pause/resume work on host.
- [ ] A safe worker job produces a **real draft PR**, with validated exact-head CI; no merge/deployment.
- [ ] Branch protection, rulesets, reviews, approval gating, logs and restoration verified.

**No changes to GitHub App registration, branch protections, secrets, production routes or public deployments are performed in this branch.**
