# MoBox UXI

MoBox UXI is the conversation-first operator surface for AIFT-OS. The foundation persists sessions, discovers local repositories from filesystem evidence, connects to an OpenAI-compatible local model, inspects persisted Forge mission state, records plans and approval decisions, and visualizes federation growth without pretending an unavailable model or unexecuted action succeeded.

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

## Federation Tree of Life

The Tree of Life tab transforms the same repository evidence used by `/inspect` into an interactive dual-tree world:

- **Tree of Life** shows runtime, infrastructure, applications, stewardship, and living services.
- **Tree of Knowledge** shows doctrine, publishing, models, memory, and the federation genome.
- **One Root** represents the AI Freedom Trust Federation.
- **Seven Living Layers** form the permanent trunk: Sovereign Foundation, Knowledge, Intelligence, Federation, Economy, Applications, and Exploration.

Repository leaves receive growth, levels, and evidence XP from observed Git state, manifests, declared capabilities, and ready capability evidence. Missing capability manifests or readiness proof become open quests. These scores are derived views, not authority, currency, execution proof, or permanent identity records.

Selecting a repository leaf reveals its role, layer, branch, evidence, capability records, growth, and an action that returns to the conversation with `/inspect <repository>`. Filters show Life, Knowledge, ready, growing, or blocked branches. The view is mobile-first, keyboard accessible, dependency-free, and read-only.

## Truth and governance contract

1. Observed evidence outranks generated claims.
2. AI output is a proposal or explanation, not proof of execution.
3. Tree levels, XP, growth, and quests are derived from observed evidence and are never proof of execution, wealth, authority, or moral value.
4. Plans, action proposals, and approval decisions may be recorded locally.
5. Approval changes authorization state but does not start a job.
6. No execution, shell, deployment, wallet, transaction, or external-write endpoint exists in this foundation.
7. Model failure is shown as degraded mode rather than hidden.
8. The daemon refuses non-loopback binding.

## API

Read and conversation:

- `GET /health`
- `GET /v1/system`
- `GET /v1/repositories`
- `GET /v1/federation/tree`
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

The Tree of Life is reconstructed from current repository evidence each time it is requested. It does not maintain a separate mutable game database, so stale or fabricated progression cannot outrank the underlying repository state.

## Validation

The GitHub Actions CI gate runs the full Go tests, binary build, formatting check, shell syntax validation, coverage threshold, and architecture invariant checks. Tree-model tests verify deterministic ordering, layer and branch classification, evidence-derived progress, open readiness quests, and the HTTP endpoint contract.

## Next safe extraction

The next execution phase must add a capability adapter registry, immutable approval scope, snapshots, validation, rollback, and job supervision before any approved action can be invoked. AIFT-Forge must remain behind an adapter rather than being called directly by browser code.
