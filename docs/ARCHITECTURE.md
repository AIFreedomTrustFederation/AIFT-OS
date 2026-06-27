# AIFT-OS Architecture

AIFT-OS is the Federation Control Plane for AI Freedom Trust Federation.

## Current architecture

- `cmd/aift` — CLI entry point.
- `internal/` — Go kernel packages.
- `bin/aift` — built executable launcher.
- `install/` — system bootstrap scripts.
- `tests/` — smoke tests.
- `schemas/` — federation schemas.
- `registry/` — generated registry output.
- `reports/` — generated dashboard and graph output.
- `legacy/shell/` — archived shell-era implementation kept for reference.

AIFT-OS discovers and orchestrates sovereign repositories. It does not absorb them.
