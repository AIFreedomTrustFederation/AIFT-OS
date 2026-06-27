# Phase 8: Federation Integration Layer

AIFT-OS now integrates sovereign repositories through manifest contracts, repository inspection, federation scanning, workflows, and repo command execution.

Commands:

- `aift repo list`
- `aift repo inspect <name>`
- `aift repo run <name> <command> [args...]`
- `aift workflow list`
- `aift federation scan`
- `aift federation graph`
- `aift federation verify`

Generated files:

- `registry/federation-snapshot.json`
- `registry/workflows.json`

Each repository remains sovereign. AIFT-OS discovers and coordinates through `.aift/` contracts.
