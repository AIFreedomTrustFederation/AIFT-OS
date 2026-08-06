# MoBox UXI

MoBox UXI is the conversation-first operator surface for AIFT-OS. The foundation persists sessions, discovers local repositories from filesystem evidence, connects to an OpenAI-compatible local model, inspects persisted Forge mission state, and records plans and approval decisions without pretending an unavailable model or unexecuted action succeeded.

## Run

```bash
bash scripts/mobox-uxi.sh
```

Or:

```bash
go run ./cmd/aiftd
```

Open `http://127.0.0.1:8787`.

Configuration:

- `AIFT_ROOT`: federation workspace root. Defaults to `~/AIFT` when present.
- `AIFT_UXI_HOME`: persistent UXI state. Defaults to `~/.aift/uxi`.
- `AIFT_UXI_ADDR`: loopback address. Defaults to `127.0.0.1:8787`.
- `AIFT_MODEL_URL`: local OpenAI-compatible endpoint. Defaults to `http://127.0.0.1:8080/v1`.
- `AIFT_MODEL_NAME`: model identifier. Defaults to `local`.

## Conversation commands

- `/inspect` lists observed repositories.
- `/inspect <repository>` returns evidence-backed repository status.
- `/forge` inspects only the persisted `AIFT-Forge/.forge/mission.json` record. It does not report Forge's generated in-memory default as observed repository state.

## Truth and governance contract

1. Observed evidence outranks generated claims.
2. AI output is a proposal or explanation, not proof of execution.
3. Plans, action proposals, and approval decisions may be recorded locally.
4. Approval changes authorization state but does not start a job.
5. No execution, shell, deployment, wallet, transaction, or external-write endpoint exists in this foundation.
6. Model failure is shown as degraded mode rather than hidden.
7. The daemon refuses non-loopback binding.

## API

Read and conversation:

- `GET /health`
- `GET /v1/system`
- `GET /v1/repositories`
- `GET /v1/adapters/forge/mission`
- `GET /v1/sessions`
- `POST /v1/sessions`
- `GET /v1/sessions/{id}`
- `POST /v1/sessions/{id}/messages`
- `GET /v1/events`

Governance records:

- `POST /v1/sessions/{id}/plans`
- `POST /v1/sessions/{id}/actions`
- `POST /v1/sessions/{id}/actions/{actionID}/decision`

There is intentionally no action-invocation endpoint.

## Persistence and recovery

Sessions are atomically replaced as private JSON files. Every session mutation and its required audit event are staged through a private transaction journal before being applied. Interrupted or transiently failed commits are recovered before later reads or writes, and event IDs prevent duplicate audit records during recovery.

User and assistant turns are committed as one exchange, so a failed Forge inspection or concurrent request cannot leave a dangling user turn. Corrupt session files are skipped rather than hiding healthy sessions, and the failure is recorded in the audit log. Events are decoded as streamed JSON values, avoiding line-length limits.

## Validation

The GitHub Actions CI gate runs the full Go tests, binary build, formatting check, shell syntax validation, coverage threshold, and architecture invariant checks. The hardened persistence and concurrency tests also pass in an isolated module reconstruction before publication.

## Next safe extraction

The next execution phase must add a capability adapter registry, immutable approval scope, snapshots, validation, rollback, and job supervision before any approved action can be invoked. AIFT-Forge must remain behind an adapter rather than being called directly by browser code.
