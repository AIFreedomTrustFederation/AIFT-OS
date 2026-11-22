# Aetherion pull-based documentation worker (proposed, opt-in)

This is an alternative to the self-hosted runner group referenced in the existing AIFT-OS workflow. It is designed for the **public** `AIFreedomTrustFederation/AIFT-OS` repository, which is currently owned by a personal GitHub user account, and does not assume organization-level runner groups.

**No model worker registration, deployment, scheduled processing, GitHub credentials, App secret, or automatic publishing is performed by this PR.** Its local worker uses anonymous **read-only** GitHub API calls, local Ollama/llama.cpp at literal loopback, and an owner-only outbox. A separate **manual** GitHub-hosted Actions dispatch carries one reviewed, bounded v2 plan to a GitHub App publisher. This is a secure first vertical slice, **not fully unattended draft publishing**. Any later autonomous relay requires a separately reviewed credential/admission boundary.

## Architecture and trust boundaries

1. `scripts/aift_pull_docs_worker.py once` reads GitHub's public `main` SHA and the already operator-labelled, open `aift-autodev-docs` issue queue (plus open PR markers). It rejects unexpected API routes, redirects, oversized responses, unauthorized paths and stale or dirty local checkouts. It makes at most one Aetherion inference per poll and invokes the existing exact-span `create_plan` validator.
2. It atomically places a maximum of **eight** private `issue-N.json` v2 plans in `~/.local/state/aift-aetherion-pull/outbox/`, outside the Git checkout. The outbox survives process restarts. Concurrent pollers are prevented by `flock`; already-queued/archived issues are excluded. Bounded oldest-first pagination inspects at most four pages of 20 approved issues, so earlier queued issues do not starve later ones. The worker syncs completed files and directory entries, rejects symlinked temporary entries, and cleans only incomplete private temp files older than 24 hours. It never pushes, comments, labels, opens a PR, accesses a GitHub token, or invokes a shell based on model output.
3. The operator reviews the exact `old`/`new` replacement with `inspect --issue N`, including the source SHA and `plan_sha256`, and must supply that digest again to `export --issue N --expected-digest ...`. Base64 is **not encryption**. This artifact only contains a bounded, approved public-docs span; it must not contain secrets or private data. Operator credentials must be kept in a **separate execution/security context** from the credentialless inference process to provide real credential isolation. Running both as the same Android Unix user does not enforce OS-level credential isolation.
4. `.github/workflows/aift-pull-doc-publisher.yml` runs on **GitHub-hosted Ubuntu only**, accepts only manual `workflow_dispatch` on `main` by `AIFreedomTrustFederation` with `operator_approved=true`. It requires the reviewed `plan_sha256` separately from the bounded base64 plan, verifies an exact checksum match using `scripts/aift_pull_plan_handoff.py`, and invokes the existing `preflight_plan` against the exact trusted checkout and freshly retrieved issue **before** obtaining an App write token. After preflight it mints a short-lived AIFT-OS-only GitHub App token and invokes the unchanged draft-only publisher, which revalidates the artifact, issue, exact SHA, file and single exact replacement before Git push.
5. If `main` or the labelled issue changed since local planning, publisher **fails closed**. The publisher now probes the **live** `refs/heads/main` SHA during read-only preflight, again before local mutation, and once immediately before its remote push. These checks reduce stale-head risk but are not an atomic server-side guarantee: GitHub branch protection, exact-head CI, and human review are still necessary. A draft PR still requires human review before merge. The GitHub App must have `contents:write` and `pull_requests:write` on AIFT-OS only. Nothing accesses public inference from a hosted runner.

## Local test and usage on Android Ubuntu PRoot (after merging and reviewing)

On Android Ubuntu PRoot, keep your existing `/root/AIFT/AIFT-OS` checkout and its untracked files untouched. Create a separate clean **detached** worktree pinned to remote `main` instead; the worker permits `main` or detached HEAD, but still verifies the full exact public SHA and clean status. The worktree path must not already exist:

```sh
cd /root/AIFT/AIFT-OS
git fetch origin main
git worktree add --detach /root/AIFT/AIFT-OS-PULL origin/main
cd /root/AIFT/AIFT-OS-PULL
git status --porcelain
git rev-parse HEAD
python3 -B -m unittest discover -s tests -p 'test_aift_pull_docs_worker.py' -v
python3 -B -m unittest discover -s tests -p 'test_aift_pull_publisher.py' -v
python3 -B -m unittest discover -s tests -p 'test_aift_pull_plan_handoff.py' -v
python3 -B scripts/aift_pull_docs_worker.py status
python3 -B scripts/aift_pull_docs_worker.py once \
  --model aetherion:latest --model-url http://127.0.0.1:11434/v1
```

If the worktree already exists, inspect its status; do not overwrite or delete it. When `main` advances, explicitly fetch and fast-forward/recreate a **clean** detached worktree after operator review. Do not implement a silent `git reset --hard` or force-upgrade. If directory fsync is unsupported by the Android/PRoot filesystem, the worker fails closed; test local durability before unattended use.

If no approved issue exists, the poll returns `idle`. If the checkout is stale, dirty, on any branch other than `main` or detached exact-main, the anonymous API is unavailable, the model fails, or the outbox is full, it refuses to publish and exits without side effects on GitHub. Note that untracked local cache files may make `verify_checkout` reject your checkout; preserve them and use the **clean, separate detached worktree** shown above rather than deleting them indiscriminately.

To request ongoing foreground polling, **only after operator review**:

```sh
python3 -B scripts/aift_pull_docs_worker.py loop \
  --model aetherion:latest --model-url http://127.0.0.1:11434/v1 \
  --interval 600
```

Intervals are limited to 600–3600 seconds to respect anonymous GitHub rate limits. Failures back off; no automatic `git pull` occurs. Android background process and Doze restrictions can still suspend or terminate polling. Restart manually or supervise separately after review; the outbox persists.

## Manual operator-approved draft handoff

Once a plan is queued and its proposed change is reviewed locally, a **separate operator environment** with GitHub authentication can submit the bounded plan. Do **not** run this relay step inside the credentialless inference daemon or expose its token to model prompts or files.

Review the escaped JSON preview **before export**:

```sh
# Example only: replace 17 with the real queued issue number.
python3 -B scripts/aift_pull_docs_worker.py inspect --issue 17
```

Inspect `issue`, `target`, `old`, `new`, `summary`, `head_sha`, and the complete `plan_sha256`. Never authorize a plan containing secrets or unapproved edits. **Still inside the credentialless Aetherion polling environment**, export the exact artifact into a restricted file:

```sh
PLAN_SHA256="<exact 64-hex digest from inspected plan>"
umask 077
python3 -B scripts/aift_pull_docs_worker.py export \
  --issue 17 --expected-digest "$PLAN_SHA256" \
  > "$HOME/issue-17-reviewed.b64"
chmod 600 "$HOME/issue-17-reviewed.b64"
```

Transfer **only** the reviewed public-documentation encoded file and its SHA-256 digest to a separately protected operator environment. Keep the polling outbox `0700`; **never** transfer your GH token or GitHub App key into the poller, and never loosen the outbox permissions to make the export readable. In that *separate operator environment*, reverify the transferred artifact and request exactly one draft:

```sh
PLAN_SHA256="<the same reviewed digest>"
PLAN_B64="$(tr -d '\r\n' < ./issue-17-reviewed.b64)"
test "$(base64 --decode ./issue-17-reviewed.b64 | sha256sum | awk '{print $1}')" = "$PLAN_SHA256"
gh workflow run aift-pull-doc-publisher.yml --repo AIFreedomTrustFederation/AIFT-OS --ref main \
  -f operator_approved=true -f plan_sha256="$PLAN_SHA256" -f plan_b64="$PLAN_B64"
unset PLAN_B64 PLAN_SHA256
```

The CLI invocation is an **explicit authenticated operator action**, not a part of `once` or `loop`.

The `gh workflow run` example is **only for a separate operator context that is authorized to publish**. Do not connect GitHub write credentials to the model's long-lived polling process or leave a privileged `gh` token accessible to the model process. The publisher independently validates the digest, then the current issue and source, **before** minting an App token.

The workflow must first be merged into the trusted default branch; it cannot be manually dispatched from PR code. Check the exact-head Actions run, GitHub App access, and resulting **draft PR**. Once an actual draft PR is verified, the operator may archive the local outbox item:

```sh
python3 -B scripts/aift_pull_docs_worker.py mark-delivered --issue 17
```

`mark-delivered` is **only** for plans with an operator-verified, actually created draft PR. **Never** use it to free space for an unpublished, failed, or stale plan. When the eight-plan limit is reached, `loop` keeps running but returns `queue_full` until capacity is freed. For a rejected/unpublished plan, inspect it, then explicitly quarantine it with the new `discard` command and its exact reviewed digest:

```sh
python3 -B scripts/aift_pull_docs_worker.py inspect --issue 17
python3 -B scripts/aift_pull_docs_worker.py discard \
  --issue 17 --expected-digest "<digest from inspect>"
```

`discard` records `discarded_unpublished` and moves the item to a private `discarded` directory. It does not pretend the item was published; future polling may reconsider the same still-approved issue. Review such reprocessing rather than silently deleting evidence. A later source or issue change invalidates the old plan; do not bypass exact-head checks.

## Threat model and limits

- **Existing protections preserved:** single existing `docs/*.md` target; blocked sensitive governance/trust/identity/security paths; one exact single-line replacement; approval label; source SHA, issue body and update hash checks; branch uniqueness; draft PR only; no direct merge; separate publisher token.
- **External data:** public issues and document text are untrusted, and base64 is not a signature. The GitHub-hosted publisher independently revalidates its source, issue and exact-span edit. A compromised operator account or GitHub App remains high impact; branch/ruleset policy and human review are essential.
- **No unattended relay:** a later autonomous transfer would need a separate limited GitHub dispatch token, an isolated relay identity, and a security review. This PR intentionally leaves that unimplemented. A local process running under the same OS account as a stored privileged `gh` token is not a strong isolation boundary.
- **Availability:** anonymous GitHub API rate limits, stale source, missing model, Android process suspension, and lack of a configured App all cause safe failure. No public model port or cloud inference fallback is used.
- **Operations:** outbox is private 0700 and plans 0600; queue limit eight; archived markers prevent reprocessing the same issue. Interrupted private `.pending-*.json` files younger than 24 hours are retained, while older incomplete files alone are removed under the queue lock. Atomic replacement and `fsync` protect completed queue files. A plan that becomes stale after remote `main` advances must be addressed through an explicit operator decision; the worker never silently deletes committed plans.

## Bash and Termux operator helper (opt-in)

After a human-reviewed merge of this repair PR, use the provided `scripts/aift_termux_operator.sh` inside Ubuntu PRoot. It is **not** an Android background service and never silently resets your worktree.

```sh
cd /root/AIFT/AIFT-OS
git fetch origin main
git merge --ff-only origin/main
bash scripts/aift_termux_operator.sh prepare
bash scripts/aift_termux_operator.sh verify
bash scripts/aift_termux_operator.sh status
bash scripts/aift_termux_operator.sh once
# If a labelled issue was queued:
bash scripts/aift_termux_operator.sh inspect 17
# Copy the digest from inspect:
bash scripts/aift_termux_operator.sh export 17 "<reviewed 64-character digest>"
```

Do not run `protection-apply` until an independent reviewer can approve your PRs; it requires `AIFT_REVIEWER_READY=yes`. Only perform `publish` in a separately protected GitHub operator context, with `AIFT_OPERATOR_SEPARATE=yes` and `AIFT_OPERATOR_APPROVED=yes` after reviewing the exact artifact. Those environment flags are explicit operator acknowledgments, **not** technical proof of OS-level isolation. The helper never merges or deploys. If your primary checkout has local edits, do not fast-forward it forcibly; inspect the edits and create a separately reviewed worktree.

## Branch protection and review obligations

At implementation time, GitHub's public branch metadata reports **`main` is not protected**. Do not treat green CI or the ability to merge as authorization. The owner should enable GitHub branch protection or a ruleset that prevents direct pushes to `main`, requires the applicable CI and publisher offline checks, requires reviews, and rejects bypasses where available. The connector used to author this PR does not expose an action to change branch-protection settings, so **this PR does not claim to configure those settings**.

CodeRabbit's requested full review must complete and actionable findings must be addressed, with a separate human/security review of local credentials, approval digest and publisher App isolation. No automated checker can supply that human approval.

## Merge acceptance

Offline unit tests and exact-head CI must pass on this PR, including Actionlint/Zizmor where applicable. A human security review must check credential separation, GitHub App scope, workflow-dispatch admission and publisher revalidation. This PR only delivers a source-level pull-based alternative; **do not claim live end-to-end draft publication, mobile 24/7 operation or production deployment until separately demonstrated**.
