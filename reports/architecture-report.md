# Architecture Report

Generated: 2026-06-28T17:34:16Z

## Summary

- **Packages**: 40
- **Commands**: 33
- **Dependencies**: 137 internal edges
- **Tested packages**: 19 / 40

## Invariant Check Results

**5 passed, 3 failed**

### PASS: no-circular-imports

### PASS: commands-have-handlers

### PASS: commands-have-help

### FAIL: no-duplicate-commands

- capabilities appears 2 times in help

### FAIL: no-orphaned-packages

- kernel (not imported by any other package)
- scheduler (not imported by any other package)

### PASS: modules-have-source

### PASS: capabilities-have-evidence

### FAIL: service-contracts-have-owner

- ServiceContract struct has no Owner field

## Command Status

- **Active**: 29
- **Planned**: 4

Planned commands:
- `intelligence`
- `manual`
- `mesh`
- `service-contracts`

## Package Categories

- **foundation**: 7 packages
- **runtime**: 6 packages
- **events**: 3 packages
- **analysis**: 4 packages
- **kernel**: 5 packages
- **data**: 5 packages
- **federation**: 3 packages
- **operations**: 2 packages
- **planning**: 3 packages
- **extensions**: 2 packages

