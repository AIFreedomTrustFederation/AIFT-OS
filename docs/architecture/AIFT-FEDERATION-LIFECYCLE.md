# AIFT Federation Lifecycle

`aift lifecycle` is the federation-wide lifecycle planner.

It discovers local federation repositories, classifies their build systems, records dirty state, records manifest state, and writes machine-readable lifecycle reports.

`aift build` remains the permanent build orchestrator.

Lifecycle does not push, delete, or mutate source repositories.

## Commands

    aift lifecycle
    aift build
    aift verify
    aift doctor

## Rules

- Never fake functionality.
- Never hardcode repository names.
- Never delete source code.
- Never overwrite human work.
- Never claim blocked work succeeded.
- Discover reality from disk.
