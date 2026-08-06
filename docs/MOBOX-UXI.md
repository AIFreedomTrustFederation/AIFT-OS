# MoBox UXI

MoBox UXI is the conversation-first operator surface for AIFT-OS. The foundation persists sessions, discovers local repositories from filesystem evidence, connects to an OpenAI-compatible local model, inspects persisted Forge mission state, records plans and approval decisions, and visualizes federation growth and geography without pretending an unavailable model or unexecuted action succeeded.

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

- **Tree of Life** shows observed runtime, infrastructure, application, stewardship, and service repositories with evidence-derived readiness classifications.
- **Tree of Knowledge** shows doctrine, publishing, models, memory, and the federation genome.
- **One Root** represents the AI Freedom Trust Federation.
- **Seven Living Layers** form the permanent trunk: Sovereign Foundation, Knowledge, Intelligence, Federation, Economy, Applications, and Exploration.

Repository leaves receive growth, levels, and evidence XP from observed Git state, manifests, declared capabilities, and ready capability evidence. Missing capability manifests or readiness proof become open quests. These scores are derived views, not authority, currency, execution proof, or permanent identity records.

Selecting a repository leaf reveals its role, layer, branch, evidence, capability records, growth, and an action that returns to the conversation with `/inspect <repository>`. Filters show Life, Knowledge, ready, growing, or blocked branches. The view is mobile-first, keyboard accessible, dependency-free, and read-only.

## Federation World Map

The World Map tab adds a local-first geographic layer without using external map tiles, geocoding services, analytics, or location uploads.

### Device GPS

GPS is requested only after the operator taps **Use my location**. Coordinates remain in browser memory and are forgotten when the page is refreshed, closed, or the operator taps **Forget GPS**. The default display precision is `city`, which rounds latitude and longitude to two decimal places. `region`, `country`, and explicitly selected `exact` precision are also available. Exact precision never earns additional XP.

After GPS permission is granted, repositories that are missing a location declaration may be temporarily grouped at the current device position. This means “the repository is hosted on this phone at the phone's current location,” not that the repository has permanently declared or published that location. Hidden and invalid declarations are never overridden by the temporary device anchor.

### Repository location declarations

A repository may declare its own location in `.aift/location.json`:

```json
{
  "schema": "aift.location.v1",
  "label": "Sacramento, California, USA",
  "latitude": 38.5816,
  "longitude": -121.4944,
  "precision": "city",
  "visibility": "federation",
  "source": "operator-declared",
  "updated_at": "2026-08-06T04:00:00Z"
}
```

The canonical JSON Schema is `schemas/location-v1.schema.json`. Declarations are limited to 64 KiB, reject unknown fields, require latitude and longitude, validate coordinate ranges, and validate an optional `updated_at` as RFC3339.

Supported precision values:

- `country`: rounded to zero decimal places.
- `region`: rounded to one decimal place.
- `city`: rounded to two decimal places and used as the default.
- `exact`: rounded to five decimal places and must be an explicit operator choice.

Supported visibility values:

- `private`: validated locally but never returned as a geographic node.
- `federation`: visible to the local federation map API.
- `public`: available to the local map API for future explicitly published federation views.

If visibility is omitted, it defaults to `private`. If precision is omitted, it defaults to `city`. The map API never returns private coordinates, filesystem paths, or device GPS coordinates.

### Privacy-neutral geographic progression

Each repository receives one evidence-derived mapping quest: **Anchor the repository on Earth**. It is completed by a valid privacy-scoped `.aift/location.json` declaration.

A valid private declaration earns exactly the same mapping XP as a federation-visible or public declaration. Country, region, city, and exact precision also earn the same XP. Visibility controls only whether a marker is displayed; it never changes level, XP, or quest completion. This prevents the game layer from pressuring operators to reveal a location or choose greater precision.

Missing, invalid, hidden, and mapped repositories remain distinct states. Temporary device anchoring is visual context only and does not earn declaration XP or permanently complete a quest.

## Truth, privacy, and governance contract

1. Observed evidence outranks generated claims.
2. AI output is a proposal or explanation, not proof of execution.
3. Tree levels, XP, growth, map levels, and quests are derived from observed evidence and are never proof of execution, wealth, authority, location ownership, or moral value.
4. Device GPS is opt-in, memory-only, and never uploaded by MoBox UXI.
5. Private repository location declarations are never returned by the federation world API.
6. Private, federation-visible, and public declarations earn equal mapping progression.
7. Plans, action proposals, and approval decisions may be recorded locally.
8. Approval changes authorization state but does not start a job.
9. No execution, shell, deployment, wallet, transaction, repository-location write, or external-write endpoint exists in this foundation.
10. Model failure is shown as degraded mode rather than hidden.
11. The daemon refuses non-loopback binding.

## API

Read and conversation:

- `GET /health`
- `GET /v1/system`
- `GET /v1/repositories`
- `GET /v1/federation/tree`
- `GET /v1/federation/world`
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

There is intentionally no action-invocation or GPS-persistence endpoint.

## Persistence and recovery

Sessions are atomically replaced as private JSON files. Every session mutation and its required audit event are staged through a private transaction journal before being applied. Interrupted or transiently failed commits are recovered before later reads or writes, and event IDs prevent duplicate audit records during recovery.

User and assistant turns are committed as one exchange, so a failed Forge inspection or concurrent request cannot leave a dangling user turn. Corrupt session files are skipped rather than hiding healthy sessions, and the failure is recorded in the audit log. Events are decoded as streamed JSON values, avoiding line-length limits.

The Tree of Life and World Map are reconstructed from current repository evidence each time they are requested. They do not maintain separate mutable game or location databases, so stale or fabricated progression cannot outrank the underlying repository state.

## Validation

The GitHub Actions CI gate runs the full Go tests, binary build, formatting check, shell syntax validation, coverage threshold, and architecture invariant checks. Tree-model tests verify deterministic ordering, layer and branch classification, evidence-derived progress, open readiness quests, and the HTTP endpoint contract. World-model tests verify coordinate validation and rounding, private-location non-disclosure, privacy-neutral rewards, declaration-size limits, strict fields, invalid and missing declarations, deterministic ordering, UI composition, and the world endpoint contract.

## Next safe extraction

The next execution phase must add a capability adapter registry, immutable approval scope, snapshots, validation, rollback, and job supervision before any approved action can be invoked. AIFT-Forge must remain behind an adapter rather than being called directly by browser code. A future location-declaration writer must also require explicit approval, show the exact file diff, and preserve the same privacy defaults before it can modify repository manifests.
