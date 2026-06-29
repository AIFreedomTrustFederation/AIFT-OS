# Phase 9: Truthful Federation Capabilities

AIFT-OS now audits what each sovereign repository can actually do.

Capability statuses:

- `planned` — intended, but not proven executable
- `detected` — inferred from repo files
- `ready` — executable exists and local verification passes
- `v1` — ready capability has been explicitly promoted/versioned
- `broken` — capability was expected or promoted but current verification fails
- `missing` — not present

Commands:

- `aift capabilities scan`
- `aift capabilities report`
- `aift capabilities repo <repo>`
- `aift capabilities promote <repo> <capability>`

Generated files:

- `registry/capabilities.json`
- `reports/capabilities.md`
- `<repo>/.aift/capabilities.json`

AIFT-OS should orchestrate only capabilities marked `ready` or `v1`.
