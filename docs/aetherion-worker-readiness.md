# Aetherion local worker readiness and controlled draft-PR acceptance

**Scope: source-only, opt-in acceptance tooling. No runner, VPS, app credentials, or production deployment is provisioned by this change.** It does not enable scheduled AI jobs, merge code, open public model ports, or create PRs automatically.

The `AIFT-OS` worker already validates bounded local Aetherion edits and separates the model-only `plan` job from the GitHub-hosted App-authenticated `publish` job in [aift-secure-cloud-worker.yml](../.github/workflows/aift-secure-cloud-worker.yml). This change adds an independently auditable readiness probe rather than replacing that boundary.

## What the evidence establishes

The local tester runs the **real** `model_edit` implementation on the same non-sensitive synthetic Markdown fixture, five consecutive times by default. An opt-in `--varied-fixtures` mode cycles through three distinct, bounded single-span documentation examples; its actual live results must be verified separately. Every resulting complete document must match the expected output, including the final newline; malformed JSON, fenced JSON, wrong fields, multiple lines, missing or extra words, missing model, timeouts and other exceptions are failures. It records pass/fail and elapsed time but **not** model responses, prompts, credentials or complete file content. It never selects a GitHub issue, edits a file, or publishes anything. The result is a limited sample, not a long-term SLA.

Run directly on the machine where Aetherion is reachable at loopback (e.g. the existing Android Ubuntu PRoot environment):

```sh
cd /root/AIFT/AIFT-OS
python3 -B -m unittest discover -s tests -p 'test_aift_aetherion_readiness.py' -v
python3 -B -m unittest discover -s tests -p 'test_aift_publisher_artifact.py' -v
python3 -B scripts/aift_aetherion_readiness.py --repetitions 5 --varied-fixtures \
  --model-url http://127.0.0.1:11434/v1 --model aetherion:latest
```

Expected: exit `0`, `status=local_worker_inference_verified`, `passed=5`, `failed=0`. Failure exits nonzero and reports `blocked`. `runner_registration`, `publisher`, and `production_deployment` **always remain** `not_checked`. A model responding on Android loopback proves nothing about a hosted VPS or GitHub runner.

## Separate, manually triggered Actions proof

The dedicated [Aetherion readiness workflow](../.github/workflows/aift-aetherion-readiness.yml) always runs offline tests on GitHub-hosted Ubuntu for PR events. Its optional live model job requires **workflow_dispatch on main** with `run_local_smoke=true` and an independently approved, ephemeral Linux runner in the GitHub Actions runner group `aift-isolated-ephemeral`, additionally labelled `self-hosted`, `linux`, and `aift-local-model`. The job pins the checkout to `github.sha`, compares the actual checkout against that SHA, and does not receive App write credentials. Evidence is retained for one day. A missing runner **or missing approved runner group** leaves the job queued; there is no provider fallback. A runner-group selector is enforced by Actions, but by itself does **not** prove that a host is ephemeral or safely isolated. Do not dispatch this job unless the runner's isolation is independently reviewed.

**Required operator preconditions:** Provision and independently review the runner group in the organization's GitHub settings before attempting dispatch; verify its repository access policy, ephemeral lifecycle, workload admission, workspace destruction, network restrictions, and separation from GitHub publisher credentials. Set the exact **Actions variables** `AIFT_MODEL_URL=http://127.0.0.1:11434/v1` and `AIFT_MODEL_NAME=aetherion:latest` **only if those values match the real approved runner's localhost service**. The workflows now **fail closed** if either variable is missing, rather than reverting to `8080/local`. The model-worker will reject non-loopback endpoints and invalid model identifiers. No secrets are required for inference.

**Security warning:** GitHub public repositories and persistent self-hosted runners are a risky combination. Do not install a persistent runner on a model/production host merely to satisfy this workflow. Prefer a tightly controlled, ephemeral runner with reviewed job admission, no Docker socket, no unrelated workspace, no write credentials, and no access to sensitive networks. Even trusted-main workflow code must be reviewed before an operator dispatches it. The workflow does not enroll or provision that runner.

## GitHub-hosted workflow security lint

The readiness PR adds a read-only `workflow-lint` job on GitHub-hosted Ubuntu, using SHA-pinned GitHub Actions, `actionlint` v1.7.12 installed through the checksummed Go module ecosystem, and `zizmor` v1.32.0 via a SHA-pinned `zizmor-action` v0.8.1. It checks the two Aetherion workflows only, reports high-severity/high-confidence security findings, and has **no local model access or publishing credentials**. A passing static audit is *not* runner isolation proof. Inspect findings and exact-head Actions results before merging.

## Independent publisher verification: artifact v2

The planner now stores only `schema_version=2`, exact source/issue bindings, the approved `old`/`new` single-line replacement, a short summary, a changed-line count, and the SHA-256 of the **expected** resulting document. It does **not** send `updated_content` or grant the model publisher credentials. On the GitHub-hosted publisher, the artifact is untrusted: the script rejects extra keys, legacy v1, invalid IDs, oversized files and bad hashes; fetches the still-labelled issue; rechecks its body and target against the exact source checkout; then independently applies `old`/`new` through the same strict validator and compares the resulting hash/line count *before* modifying a tracked file. Altering the plan's displayed hash cannot by itself confer trust: the issue approval, content scope and human PR review remain required. This does not cryptographically attest the runner's integrity; provenance hardening remains further work.

The new `tests/test_aift_publisher_artifact.py` exercises valid reconstruction and corrupted or adversarial artifacts without network calls or Git writes. Artifact v1 is intentionally incompatible and must not be republished.

## End-to-end draft-PR acceptance (separate operator approval)

After the isolated runtime is approved and live inference passes:

1. Confirm that AIFT-Runtime's model artifact matches the authorized source and the model identity, host isolation, restart behavior, resource limits, and offline failure behavior are documented. Stop if runner registration, deployment, or secrets have not been approved.
2. Independently verify that the **existing** GitHub App can mint a short-lived AIFT-OS-only token with `contents:write` and `pull_requests:write`, without exposing a key to the model or planner. The app key stays only in GitHub Actions secrets on the GitHub-hosted publisher job, never in the inference process.
3. Keep `AIFT_AUTODEV_AI_ENABLED=false`. Create or select **one operator-approved ordinary docs issue** whose body contains one `Target: docs/...md` and one `Instruction: ...`. Manually apply the `aift-autodev-docs` label only after reviewing the target and instructions. Never use governance, trust, identity, security, deployment or credentials files as a first test.
4. On trusted `main`, manually run [AIFT Secure Cloud AI Documentation Worker](../.github/workflows/aift-secure-cloud-worker.yml) with `run_ai_once=true`. Inspect the planning job's actual model outcome, the JSON plan and source SHA, the publisher's rechecks, and the resulting **draft PR only**. If the runner is missing or GitHub App permissions are not ready, stop; do not bypass checks or route inference to a cloud model.
5. Verify draft PR scope, issue provenance, exact-head CI, requested reviews, changed content and branch protections. The smoke test never proves that the publisher worked. Do not merge or deploy as part of this acceptance test.
6. Before any later unattended scheduling, separately verify outage behavior, restart recovery, bounded queue/concurrency, artifact provenance, backups and recovery, incident handling, resource headroom, log redaction, and sustained success across varied documents. Repeated identical fixture success is **not** sufficient for autonomous production operation.

No schedule or AI enablement toggle is changed here. Any future change to runner enrollment, production host configuration, GitHub App secrets, trust policy, high-impact workflows or autonomous merges requires separate operator review and approval.
