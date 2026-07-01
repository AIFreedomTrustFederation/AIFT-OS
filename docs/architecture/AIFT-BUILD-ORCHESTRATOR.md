# AIFT Build Orchestrator

`aift build` is the permanent build pipeline for AIFT-OS.

It replaces phase-only scripts with an operating-system-owned workflow.

## Pipeline

1. Compile repository reality.
2. Run doctor.
3. Run verification.
4. Run Go tests.
5. Build native CLI.
6. Write build reports.
7. Stop before unsafe mutation.

## Rules

- Never fake functionality.
- Never hardcode repository names.
- Never delete source code.
- Never overwrite human work.
- Never claim blocked work succeeded.
