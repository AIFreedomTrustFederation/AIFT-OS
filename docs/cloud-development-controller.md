# AIFT Cloud Development Controller — review-first foundation

AIFT-OS coordinates AIFreedomTrustFederation repositories without assuming authority over them. This cloud extension is a **scheduled, bounded, evidence-producing controller** hosted in GitHub Actions. It complements, and does not replace, the local-first AIFT-Forge build farm, AIFT-OS evidence collector, or separate ChatGPT automations.

## Actual capabilities (not a continuously running process)

- Every 15 minutes, GitHub Actions *attempts* a scheduled read-only snapshot. GitHub may delay or drop scheduled runs; the job is not a permanently running worker. The workflow can also run manually from the default branch.
- The observation job samples explicitly configured repositories, open PRs, and exact-head GitHub check runs. It emits bounded JSON evidence as an Actions artifact and a readable Actions summary. **Reported checks are not proof of required-check status, mergeability or review approval.** Missing API access and incomplete sampling remain unverified.
- An **opt-in** proposal job applies gofmt to tracked internal Go test files checked out from trusted main, runs go test ./internal/..., and opens a **draft PR**. It never merges, deploys, edits other repositories, or grants itself permissions. The proposal job is disabled for scheduled runs until variable AIFT_AUTODEV_AUTO_PROPOSE is true.
- No LLM inference, free-form AI-generated code changes, unattended cloud service, or deployment is claimed. Those require separate scoped workers, an approved model/provider, and explicit operating limits.

## Set up unattended routine proposals

1. In GitHub, register a dedicated **GitHub App** under account AIFreedomTrustFederation (or an operator-managed account). Install it **only on AIFT-OS** for this phase. Grant **Contents: Read and write** and **Pull requests: Read and write** repository permissions. No Administration, Actions write, Secrets, Deployments, or account-wide privileges are needed.
2. Under AIFT-OS **Settings → Secrets and variables → Actions**, set variable AIFT_AUTODEV_APP_CLIENT_ID to the App client ID, and secret AIFT_AUTODEV_APP_PRIVATE_KEY to the App PEM private key. Never commit the PEM or send it in a chat. GitHub mints short-lived installation tokens during trusted main-branch jobs.
3. Run the controller manually on **main** with propose_routine_fixes enabled. Review logs and the draft PR if a genuine gofmt discrepancy exists.
4. After reviewing behavior, set AIFT_AUTODEV_AUTO_PROPOSE=true to enable scheduled format-only proposals. Leave unset/false for observation-only operation.
5. Optional AIFT_AUTODEV_REPOS is a comma-separated allowlist of 1–12 repository names. Default: AIFT-OS,AIFT-Runtime,AIFT-Forge,VPS,www.aifreedomtrust.com. Repository-scoped GitHub tokens cannot necessarily inspect other repositories' private Actions or check-run data; errors appear as unavailable, not passing. For private repositories, a separate read-only App token with explicit scope is necessary.

## Trust boundaries

- Pull-request runs execute only offline safety tests; they receive no dedicated App secrets. The privileged proposal job requires refs/heads/main even for manual dispatch.
- Proposed modifications are restricted to pre-existing, tracked Go test files in internal/. A dirty checkout or unexpected modified file fails closed. Unverified PR heads are never executed with write credentials.
- The controller pins the privileged checkout to the event's exact main commit, creates a new proposal branch, refuses to overwrite an existing proposal branch, and opens only **draft** PRs. No branch protection bypass, main push, auto-merge, credential changes or deployment.
- An operator must still review exact-head CI, branch protection, reviews and cross-repository dependencies before merge. Protected trust/identity/governance/security/deployment changes remain outside this job.
- Dependencies are commit-SHA pinned. App installation and repository policy require review before enabling write access.
- The report is bounded: at most 8 PRs sampled per repository, 4 exact-head PR check snapshots per repository, and 100 check runs per inspected PR. It does not claim to represent every open PR.

## Validation and limits

Offline unit tests: python3 -B -m unittest discover -s tests -p 'test_aift_cloud_dev_controller.py' -v.

Canonical repository gate: pnpm install --frozen-lockfile --ignore-scripts; pnpm run typecheck; make verify. Hosted PR CI must establish exact-head success before merging.

GitHub Actions schedules are best-effort and require an enabled workflow on the default branch. For a **true continuously connected 24/7 AI worker**, run an independently supervised service with restart-on-failure, queueing, quotas, model/API credentials and the same approval boundaries. The phone's Termux/Ubuntu setup remains the mobile development console, **not** an always-running server. This controller is a safe cloud foundation, not a claim of completed autonomous coding.
