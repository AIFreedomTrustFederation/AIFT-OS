# AIFT Federation Build Engine

`aift federation-build` discovers and builds the federation without hardcoded repository names.

## Design

Every repository is treated as a module.

Every module may be:

- synchronous
- asynchronous
- queued
- blocked
- skipped
- planned
- active after verification

## Rules

- Never fake functionality.
- Never hardcode repository names.
- Never assume package managers.
- Never delete source code.
- Never overwrite human work.
- Never claim blocked work succeeded.
- Discover reality from disk.
- Build only when a real build system is detected.
