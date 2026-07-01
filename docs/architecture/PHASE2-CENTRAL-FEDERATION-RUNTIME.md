# AIFT-OS Phase 2: Central Federation Runtime

AIFT-OS is the single federation runtime.

All other repositories are mounted source packages.

## Rules

- Repositories remain source repositories.
- Runtime state belongs in AIFT-OS/runtime.
- Federation registry data belongs in AIFT-OS/registry.
- Generated reports belong in AIFT-OS/registry/reports.
- Per-repository .aift files are compatibility inputs only.
- aift verify must not leave mounted repositories dirty.
- aift scan discovers reality from the workspace.
- aift graph reads the central registry.
- aift clean removes generated runtime artifacts only.

## Canonical Layout

~/AIFT/
  AIFT-OS/
    runtime/
    registry/
    reports/
    cmd/aift/
  AI-Freedom-Trust/
  AIFT-Forge/
  BookSmith-Federation-OS/
