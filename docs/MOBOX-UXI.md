# MoBox UXI

MoBox UXI is the conversation-first operator surface for AIFT-OS. It persists governed sessions, discovers local repositories from filesystem evidence, connects to an OpenAI-compatible local model, inspects persisted Forge mission state, and supervises a deliberately small registry of read-only adapters.

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

## Registered adapters

The default registry contains exactly two adapters:

- `repository.inspect` — reads one discovered repository and produces an evidence artifact.
- `forge.mission.inspect` — reads the persisted Forge mission and produces an evidence artifact.

Both descriptors declare `mutating: false`. The execution supervisor refuses every adapter that declares `mutating: true` before a job is created.

## Truth and governance contract

1. Observed evidence outranks generated claims.
2. AI output is a proposal or explanation, not proof of execution.
3. Plans, action proposals, and approval decisions may be recorded locally.
4. Approval changes authorization state but does not itself start a job.
5. Only registered read-only adapters can be invoked.
6. An invocation creates a supervised job and a digest-addressed private JSON artifact.
7. Adapter failure is persisted on both the action and job.
8. No shell, deployment, wallet, transaction, file-mutation, or external-write adapter is registered.
9. Model failure is shown as degraded mode rather than hidden.
10. The daemon refuses non-loopback binding.

## API

Read and conversation:

- `GET /health`
- `GET /v1/system`
- `GET /v1/repositories`
- `GET /v1/adapters`
- `GET /v1/adapters/forge/mission`
- `GET /v1/artifacts/{id}`
- `GET /v1/sessions`
- `POST /v1/sessions`
- `GET /v1/sessions/{id}`
- `POST /v1/sessions/{id}/messages`
- `GET /v1/events`

Governance records:

- `POST /v1/sessions/{id}/plans`
- `POST /v1/sessions/{id}/actions`
- `POST /v1/sessions/{id}/actions/{actionID}/decision`

Supervised read-only invocation:

- `POST /v1/sessions/{id}/actions/{actionID}/invoke`

The invocation endpoint accepts only an action whose kind resolves to a registered non-mutating adapter. Approval-required actions must be approved first.

## Persistence

- Sessions are atomically replaced as private JSON files.
- Events are appended to a private JSONL log.
- Adapter artifacts are private JSON files stored under `~/.aift/uxi/artifacts` by default.
- Every artifact record contains a SHA-256 digest.
- Jobs record adapter kind, status, start/end times, result artifact, and error text.

## Browser boundary

The UXI is dependency-free and embedded into the Go binary as separate HTML, CSS, and JavaScript assets. The Content Security Policy is self-only and contains no `unsafe-inline` exception. Browser code can propose, approve, and request a registered read-only invocation; it cannot submit arbitrary commands.

## Next safe extraction

Any future mutating adapter must add immutable approval scope, pre-action snapshots, validation, rollback, idempotency, concurrency controls, and adapter-specific tests before registration. AIFT-Forge must remain behind a typed adapter rather than being called directly by browser code.
