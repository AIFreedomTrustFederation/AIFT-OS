# AIFT Patch Engine

The Patch Engine is the safe mutation layer of AIFT-OS.

It exists because blindly modifying repositories with string replacement does not scale.

## Current foundation

The initial Patch Engine can:

- inspect patchable source artifacts
- generate a patch plan
- run validation commands
- write machine-readable validation results
- write a human-readable patch plan report

## Commands

- `aiftd patch-engine inspect`
- `aiftd patch-engine plan`
- `aiftd patch-engine validate`

## Future direction

The Patch Engine should evolve toward syntax-aware mutation using Go AST, Tree-sitter, LSP, or other parsers where practical.
