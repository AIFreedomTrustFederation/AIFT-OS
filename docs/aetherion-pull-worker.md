# Aetherion pull-based documentation worker (proposed, opt-in)

This is an alternative to the self-hosted runner group referenced in the existing AIFT-OS workflow. It is designed for the **public** `AIFreedomTrustFederation/AIFT-OS` repository, which is currently owned by a personal GitHub user account, and does not assume organization-level runner groups.

**No model worker registration, deployment, scheduled processing, GitHub credentials, App secret, or automatic publishing is performed by this PR.** Its local worker uses anonymous **read-only** GitHub API calls, local Ollama/llama.cpp at literal loopback, and an owner-only outbox. A separate **manual** GitHub-hosted Actions dispatch carries one reviewed, bounded v2 plan to a GitHub App publisher. This is a secure first vertical slice, **not fully unattended draft publishing**. Any later autonomous relay requires a separately reviewed credential/admission boundary.

## Architecture and trust boundaries

1. `scripts/aift_pull_docs_worker.py once` reads GitHub's public `main` SHA and the already operator-labelled, open `aift-autodev-docs` issue queue (plus open PR markers). It rejects unexpected API routes, redirects, oversized responses, unauthorized paths and stale or dirty local checkouts. It makes at most one Aetherion inference per poll and invokes the existing exact-span `create_plan` validator.
2. It atomically places a maximum of **eight** private `issue-N.json` v2 plans in `~/.local/state/aift-aetherion-pull/outbox/`, outside the Git checkout. The outbox survives process restarts. Concurrent pollers are prevented by `flock`; already-queued/archived issues are excluded. It never pushes, comments, labels, opens a PR, accesses a GitHub token, or invokes a shell based on model output.
3. The operator reviews the plan and uses `export --issue N` to encode it (base64 is **not encryption**). This artifact only contains a bounded, approved public-docs span; it must not contain secrets or private data. Operator credentials must be kept in a **separate execution/security context** from the credentialless inference process to provide real credential isolation. Running both as the same Android Unix user does not enforce OS-level credential isolation.
4. `.github/workflows/aift-pull-doc-publisher.yml` runs on **GitHub-hosted Ubuntu only**, accepts only manual `workflow_dispatch` on `main` by `AIFreedomTrustFederation` with `operator_approved=true`. It refuses malformed or overlong base64, then invokes the existing `preflight_plan` against the exact trusted checkout and freshly retrieved issue **before** obtaining an App write token. After preflight it mints a short-lived AIFT-OS-only GitHub App token and invokes the unchanged draft-only publisher, which revalidates the artifact, issue, exact SHA, file and single exact replacement before Git push.
5. If `main` or the labelled issue changed since local planning, publisher **fails closed**. A draft PR still requires human review before merge. The GitHub App must have `contents:write` and `pull_requests:write` on AIFT-OS only. Nothing accesses public inference from a hosted runner.

## Local test and usage on Android Ubuntu PRoot (after merging and reviewing)

From `/root/AIFT/AIFT-OS`, with a clean `main` checkout fast-forwarded to the exact public GitHub `main` head and loopback Aetherion already available:

```sh
python3 -B -m unittest discover -s tests -p 'test_aift_pull_docs_worker.py' -v
python3 -B -m unittest discover -s tests -p 'test_aift_pull_publisher.py' -v
python3 -B scripts/aift_pull_docs_worker.py status
python3 -B scripts/aift_pull_docs_worker.py once \
  --model aetherion:latest --model-url http://127.0.0.1:11434/v1
```

If no approved issue exists, the poll returns `idle`. If the checkout is stale, dirty, not on `main`, the anonymous API is unavailable, the model fails, or the outbox is full, it refuses to publish and exits without side effects on GitHub. Note that untracked local cache files may make `verify_checkout` reject your checkout; preserve them and use a **clean, separate worktree** for planning rather than deleting them indiscriminately.

To request ongoing foreground polling, **only after operator review**:

```sh
python3 -B scripts/aift_pull_docs_worker.py loop \
  --model aetherion:latest --model-url http://127.0.0.1:11434/v1 \
  --interval 600
```

Intervals are limited to 600–3600 seconds to respect anonymous GitHub rate limits. Failures back off; no automatic `git pull` occurs. Android background process and Doze restrictions can still suspend or terminate polling. Restart manually or supervise separately after review; the outbox persists.

## Manual operator-approved draft handoff

Once a plan is queued and its proposed change is reviewed locally, a **separate operator environment** with GitHub authentication can submit the bounded plan. Do **not** run this relay step inside the credentialless inference daemon or expose its token to model prompts or files.

```sh
# Example: replace 17 with the actual queued issue number.
PLAN_B64="$(python3 -B scripts/aift_pull_docs_worker.py export --issue 17)"
gh workflow run aift-pull-doc-publisher.yml --ref main \
  -f operator_approved=true -f plan_b64="$PLAN_B64"
unset PLAN_B64
```

The workflow must first be merged into the trusted default branch; it cannot be manually dispatched from PR code. Check the exact-head Actions run, GitHub App access, and resulting **draft PR**. Once an actual draft PR is verified, the operator may archive the local outbox item:

```sh
python3 -B scripts/aift_pull_docs_worker.py mark-delivered --issue 17
```

`mark-delivered` merely records an **operator-reported** local acknowledgment, not independently verified GitHub publication. The poller never deletes a plan automatically. A later update to the source or issue will invalidate stale plans; do not bypass the exact-head checks. Use a new approved issue for a new task.

## Threat model and limits

- **Existing protections preserved:** single existing `docs/*.md` target; blocked sensitive governance/trust/identity/security paths; one exact single-line replacement; approval label; source SHA, issue body and update hash checks; branch uniqueness; draft PR only; no direct merge; separate publisher token.
- **External data:** public issues and document text are untrusted, and base64 is not a signature. The GitHub-hosted publisher independently revalidates its source, issue and exact-span edit. A compromised operator account or GitHub App remains high impact; branch/ruleset policy and human review are essential.
- **No unattended relay:** a later autonomous transfer would need a separate limited GitHub dispatch token, an isolated relay identity, and a security review. This PR intentionally leaves that unimplemented. A local process running under the same OS account as a stored privileged `gh` token is not a strong isolation boundary.
- **Availability:** anonymous GitHub API rate limits, stale source, missing model, Android process suspension, and lack of a configured App all cause safe failure. No public model port or cloud inference fallback is used.
- **Operations:** outbox is private 0700 and plans 0600; queue limit eight; archived markers prevent reprocessing the same issue. Hidden interrupted temp files can remain for human inspection after a crash; never automatically process them.

## Merge acceptance

Offline unit tests and exact-head CI must pass on this PR, including Actionlint/Zizmor where applicable. A human security review must check credential separation, GitHub App scope, workflow-dispatch admission and publisher revalidation. This PR only delivers a source-level pull-based alternative; **do not claim live end-to-end draft publication, mobile 24/7 operation or production deployment until separately demonstrated**.
