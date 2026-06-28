# Phase 13: Federation Event Mesh

AIFT-OS now has the foundation for asynchronous federation coordination.

## Principle

Repositories are not chained through hard-coded if/then workflows.

They publish events, subscribe to events, and declare handlers through `.aift/events.json`.

## Commands

- `aift mesh init-all`
- `aift mesh scan`
- `aift mesh topics`
- `aift mesh subscribers`
- `aift mesh publish <topic> [source] [message]`
- `aift mesh replay [topic]`
- `aift mesh tail`
- `aift mesh report`

## Per Repo

- `.aift/events.json`
- `.aift/events/handlers/`

## Generated

- `registry/event-mesh.json`
- `reports/event-mesh.md`

## Truth Rule

A topic or subscriber can be planned, detected, ready, v1, broken, deprecated, or disabled.

AIFT-OS records event contracts but does not pretend unverified handlers are running.
