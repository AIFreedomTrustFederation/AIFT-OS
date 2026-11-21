# AIFT-Native AI Documentation Worker — No External AI API Keys

This review-only change replaces the OpenAI Responses API planner with
**AIFT-Runtime Aetherion**. No third-party model, OpenAI key, or cloud
AI fallback is needed.

AIFT-OS already supports an OpenAI-**compatible local JSON format** in
`internal/uxi/inference.go`, defaulting to `http://127.0.0.1:8080/v1`.
That protocol name does not mean the model is operated by OpenAI.
AIFT-Runtime's `runtime/ai/start.sh` starts a loopback `llama-server`
with `runtime/ai/models/aetherion.gguf`.

This design is not live until the model, dedicated runner, and draft-PR
end-to-end test are actually verified.

## Execution architecture

| Stage | Host | Credentials |
| --- | --- | --- |
| Offline tests | GitHub-hosted Ubuntu | No model or App secret |
| Plan | Dedicated AIFT-hosted Linux runner labelled `aift-local-model` | Read-only GitHub token and local model |
| Publish | GitHub-hosted Ubuntu | Short-lived AIFT-OS-scoped GitHub App write token, no model access |
| Independent persistent worker | Approved always-on AIFT host | Existing App 5861093; separately deployed |

GitHub-hosted runner loopback is **not** the phone or AIFT VPS loopback.
The planner must share a host/network namespace with the Aetherion server.
Its HTTP client accepts only literal `http://127.0.0.1:<port>/v1`.
It has no DNS discovery, redirects, remote inference or external API fallback.

## Activation checklist

1. Review this draft PR and pass exact-head CI before merging or deploying.
2. On an approved model host, load the real Aetherion GGUF and launch the
   expected llama.cpp service. Existing AIFT-Runtime status and startup
   messages are *not* proof that the intended model can generate.
   Test an actual completion:

   ```bash
   curl -fsS http://127.0.0.1:8080/v1/models
   curl -fsS http://127.0.0.1:8080/v1/chat/completions \
     -H 'Content-Type: application/json' \
     -d '{"model":"local","messages":[{"role":"user","content":"Reply READY"}],"stream":false}'
   ```

   If the model rejects `local`, use the identifier shown by /v1/models.
3. Register a **dedicated repo-scoped self-hosted Linux Actions runner**
   on that same host with label `aift-local-model`. Isolate the runner
   from other repositories and untrusted PR workloads. Run only the
   explicitly trusted-main planning job on it. A missing runner leaves
   the job queued rather than selecting an external AI provider.
4. Optional AIFT-OS Actions *variables* (not secrets):
   - `AIFT_MODEL_URL=http://127.0.0.1:8080/v1`
   - `AIFT_MODEL_NAME=local` (or actual loaded model ID)
   - `AIFT_AUTODEV_AI_ENABLED=false` initially

   Do not create `AIFT_AUTODEV_OPENAI_API_KEY`. The previous
   `AIFT_AUTODEV_MODEL` variable no longer controls this planner.
5. The publisher still needs **GitHub** identity, not a model API key.
   Reuse the existing AIFT Federation Automation GitHub App ID 5861093.
   The observed evidence setup proved producer **read** access only:
   validate that a single AIFT-OS installation token can be minted
   with `contents:write` and `pull_requests:write` before activating.
   Set `AIFT_AUTODEV_APP_CLIENT_ID` (Actions variable) and
   `AIFT_AUTODEV_APP_PRIVATE_KEY` (Actions secret) for the *same* App.
   Never copy PEM contents into Git, chat, logs, model prompts or artifacts.
6. Add the operator-controlled label `aift-autodev-docs`. Label exactly
   one approved ordinary documentation issue, for example:

   ```text
   Target: docs/validation.md
   Instruction: Clarify this existing example without inventing test results.
   ```

   That target must already be tracked on main. Label approval authorizes
   only one draft proposal, not a merge.
7. Manually dispatch **AIFT Secure Cloud AI Documentation Worker** on
   trusted `main` with `run_ai_once=true`. Verify the planner's local
   inference, plan validation, separate publisher and resulting draft PR.
   Verify the draft's exact-head CI, issue provenance, diff and reviewer
   requirements before any merge.
8. Only after model, runner, permission boundaries and smoke tests pass
   may an operator set `AIFT_AUTODEV_AI_ENABLED=true` for scheduled
   issue processing. GitHub schedules are best effort; this does not
   prove permanent worker uptime.

## Security and boundaries

- The planner runs trusted main code only, with no GitHub write token.
- No model-issued shell commands, arbitrary file edits or tool calls.
- Labelled issues may change only one existing non-sensitive `docs/*.md`.
- The local OpenAI-compatible chat request uses `response_format` with
  `type: json_schema`, a strict three-field `{old, new, summary}` schema, and
  temperature `0`. This is a **local protocol feature**, not an external AI
  service. Aetherion proposes exactly one original text span and its replacement,
  not a full rewritten document. The worker verifies that the old span occurs
  exactly once in the actual source, that both spans are single-line and
  nonempty, and that the replacement passes bounded change validation.
  It deterministically applies the substitution so headings, other lines and
  the original final newline are preserved verbatim. It refuses ambiguous,
  missing, multiline, oversized, no-op or malformed proposals. It cannot
  guarantee semantic correctness: reviewers must inspect every draft PR.
- A truncated or non-JSON reply blocks publication; there is no heuristic
  repair and no cloud fallback. Small GGUFs can still fail to follow
  instructions, or select the wrong span, even under constrained decoding.
- Strict JSON, output-size and changed-line limits. Invalid or malformed
  local output blocks publication.
- Publisher rechecks label, issue body, exact main commit, source hash,
  single changed file, Git diff and repository-scoped permissions.
- Publisher alone can push a new branch and create a **draft PR**.
- No automatic merge, deployment, branch protection bypass or sensitive
  trust, identity, governance or credentials modifications.
- Inference offline or responding incorrectly means failure; there is
  explicitly **no cloud AI fallback**.

## Verification

```bash
python3 -B -m unittest discover -s tests -p test_aift_secure_cloud_worker.py -v
pnpm install --frozen-lockfile --ignore-scripts
pnpm run typecheck
make verify
git diff --check
```

Still required: real Aetherion model response and identity, dedicated runner
security, existing App write-scope check, one genuine draft PR, exact-head
hosted CI, outage behavior, isolated worker supervision, restart recovery,
and proof the model host runs while Android is disconnected. This change
does not provision a host or deploy a runner.
