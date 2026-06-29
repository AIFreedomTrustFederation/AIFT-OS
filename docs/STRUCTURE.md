# AIFT-OS Structure

AIFT-OS is the Federation Control Plane.

It orchestrates sovereign repositories under `~/AIFT` without absorbing them.

## Main folders

- `app/` — future real application source
- `app/cli/` — command-line interface layer
- `app/core/` — control-plane core logic
- `app/providers/` — provider interfaces and adapters
- `app/plugins/` — plugin discovery and execution logic
- `app/reports/` — report generation logic
- `app/tests/` — application-level tests
- `commands/` — current shell command modules
- `runtime/` — current shell runtime libraries
- `providers/` — current shell provider adapters
- `plugins/` — built-in AIFT-OS plugins
- `config/` — local/default configuration
- `schemas/` — JSON schemas
- `tests/` — smoke tests and regression tests
- `ci/` — CI scripts and workflow helpers
- `tools/` — developer utilities
- `registry/` — generated federation registry
- `reports/` — generated federation reports
- `docs/` — documentation
- `logs/` — local logs
- `var/` — local runtime state
