# AIFT-OS Roadmap

## Completed

- Repository structure
- Go CLI kernel
- Launcher stabilization
- Registry generation
- Manifest generation
- Dashboard and dependency reports
- Event log
- Provider registry
- Runtime state
- Service list
- Supervisor and job runner
- Local API skeleton
- AI Code Training archive

## v0.3.0-services

This release is the first stable runtime baseline.

It includes:

- `aift version`
- `aift doctor`
- `aift status`
- `aift manifest`
- `aift registry`
- `aift dashboard`
- `aift deps`
- `aift providers`
- `aift events`
- `aift services`
- `aift start`
- `aift tick`
- `aift verify`
- `aift daemon :8787`

## Next: Phase 8 Plugin Runtime

Planned commands:

- `aift plugin list`
- `aift plugin run <repo> <command>`
- `aift plugin enable <repo> <command>`
- `aift plugin disable <repo> <command>`

## Later

- Full provider interfaces
- Service manager
- Local UI API
- Forge integration
- Federation node discovery
- Cross-node sync
- AI orchestration
