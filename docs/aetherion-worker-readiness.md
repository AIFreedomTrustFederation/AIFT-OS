# Aetherion local worker readiness and controlled draft-PR acceptance

**Scope: source-only, opt-in acceptance tooling. No runner, VPS, app credentials, or production deployment is provisioned by this change.** It does not enable scheduled AI jobs, merge code, open public model ports, or create PRs automatically.

The `AIFT-OS` worker already validates bounded local Aetherion edits and separates the model-only `plan` job from the GitHub-hosted App-authenticated `publish` job in [aift-secure-cloud-worker.yml](../.github/workflows/aift-secure-cloud-worker.yml). This change adds an independently auditable readiness probe rather than replacing that boundary.

## What the evidence establishes

The local tester runs the **real** `model_edit` implementation on the same non-sensitive synthetic Markdown fixture, five consecutive times by default. Every resulting complete document must match the expected output, including the final newline; malformed JSON, fenced JSON, wrong fields, multiple lines, missing or extra words, missing model, timeouts and other exceptions are failures. It records pass/fail and elapsed time but **not** model responses, prompts, credentials or complete file content. It never selects a GitHub issue, edits a file, or publishes anything. The result is a limited sample, not a long-term SLA.

Run directly on the machine where Aetherion is reachable at loopback (e.g. the existing Android Ubuntu PRoot environment):

```sh
cd /root/AIFT/AIFT-OS
python3 -B -m unittest discover -s tests -p 'test_aift_aetherion_readiness.py' -v
python3 -B scripts/aift_aetherion_readiness.py --repetitions 5 \
  --model-url http://127.0.0.1:11434/v1 --model aetherion:latest
```

Expected: exit `0`, `status=local_worker_inference_verified`, `passed=5`, `failed=0`. Failure exits nonzero and reports `blocked`. `runner_registration`, `publisher`, and `production_deployment` **always remain** `not_checked`. A model responding on Android loopback proves nothing about a hosted VPS or GitHub runner.

## Separate, manually triggered Actions proof

The dedicated [Aetherion readiness workflow](../.github/workflows/aift-aetherion-readiness.yml) always runs offline tests on GitHub-hosted Ubuntu for PR events. Its optional live model job requires **workflow_dispatch on main** with `run_local_smoke=true` and a preapproved isolated Linux runner labelled `aift-local-model`. The job pins the checkout to `github.sha`, compares the actual checkout against that SHA, and does not receive App write credentials. Evidence is retained for one day. A missing runner leaves the job queued; there is no provider fallback. Do not dispatch this job unless the runner's isolation is independently reviewed.

**Security warning:** GitHub public repositories and persistent self-hosted runners are a risky combination. Do not install a persistent runner on a model/production host merely to satisfy this workflow. Prefer a tightly controlled, ephemeral runner with reviewed job admission, no Docker socket, no unrelated workspace, no write credentials, and no access to sensitive networks. Even trusted-main workflow code must be reviewed before an operator dispatches it. The workflow does not enroll or provision that runner.

## End-to-end draft-PR acceptance (separate operator approval)

After the isolated runtime is approved and live inference passes:

1. Confirm that AIFT-Runtime's model artifact matches the authorized source and the model identity, host isolation, restart behavior, resource limits, and offline failure behavior are documented. Stop if runner registration, deployment, or secrets have not been approved.
2. Independently verify that the **existing** GitHub App can mint a short-lived AIFT-OS-only token with `contents:write` and `pull_requests:write`, without exposing a key to the model or planner. The app key stays only in GitHub Actions secrets on the GitHub-hosted publisher job, never in the inference process.
3. Keep `AIFT_AUTODEV_AI_ENABLED=false`. Create or select **one operator-approved ordinary docs issue** whose body contains one `Target: docs/...md` and one `Instruction: ...`. Manually apply the `aift-autodev-docs` label only after reviewing the target and instructions. Never use governance, trust, identity, security, deployment or credentials files as a first test.
4. On trusted `main`, manually run [AIFT Secure Cloud AI Documentation Worker](../.github/workflows/aift-secure-cloud-worker.yml) with `run_ai_once=true`. Inspect the planning job's actual model outcome, the JSON plan and source SHA, the publisher's rechecks, and the resulting **draft PR only**. If the runner is missing or GitHub App permissions are not ready, stop; do not bypass checks or route inference to a cloud model.
5. Verify draft PR scope, issue provenance, exact-head CI, requested reviews, changed content and branch protections. The smoke test never proves that the publisher worked. Do not merge or deploy as part of this acceptance test.
6. Before any later unattended scheduling, separately verify outage behavior, restart recovery, bounded queue/concurrency, artifact provenance, backups and recovery, incident handling, resource headroom, log redaction, and sustained success across varied documents. Repeated identical fixture success is **not** sufficient for autonomous production operation.

No schedule or AI enablement toggle is changed here. Any future change to runner enrollment, production host configuration, GitHub App secrets, trust policy, high-impact workflows or autonomous merges requires separate operator review and approval.
