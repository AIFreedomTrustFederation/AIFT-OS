# Runtime Readiness

AIFT-OS evaluates federation objects for operational readiness. Every object is assigned a status backed by evidence.

## Object Kinds

| Kind | Source |
|---|---|
| repository | `workspace.FindRepos` (.git directories) |
| module | `registry/modules.json` |
| capability | `registry/capabilities.json` |
| service | `registry/service-contracts.json` |
| event | `var/events/event-bus.jsonl` |
| command | `registry/architecture.json` |
| script | `scripts/**/*.sh` files |

## Status Model

Every object has exactly one status. Transitions are validated.

| Status | Meaning |
|---|---|
| `planned` | Declared in a manifest or roadmap; no implementation |
| `detected` | Discovered on disk but not validated as operational |
| `ready` | Validated: has evidence of correct implementation |
| `active` | Running or producing events |
| `blocked` | Cannot proceed; missing dependency or broken |
| `deprecated` | Scheduled for removal |
| `removed` | No longer present (terminal state) |

### Transition Rules

```
planned    -> detected, ready, removed
detected   -> ready, blocked, deprecated, removed
ready      -> active, blocked, deprecated
active     -> blocked, deprecated
blocked    -> detected, ready, removed
deprecated -> removed
removed    -> (terminal, no transitions out)
```

## Evidence

Every status assignment includes an `evidence` string explaining why the object has that status. Examples:

- `".aift/repo.json manifest"` - repository has AIFT manifest
- `"sources aift-run.sh harness"` - script integrates with execution harness
- `"planned command stub"` - command returns "not implemented" error
- `"case in main.go switch + help entry"` - command has handler and help text

## CLI Commands

```
aift runtime scan      # Evaluate all objects, write registry + report
aift runtime status    # Print all objects with status
aift runtime ready     # Print only ready/active objects
aift runtime blocked   # Print only blocked objects
aift runtime report    # Print the readiness report (markdown)
```

## Operator Workflow

Runtime readiness is part of the standard operator workflow. The recommended
sequence is:

```bash
aift verify              # Run all verification steps
aift runtime scan        # Evaluate readiness
aift runtime report      # Review results
```

Or use the aggregate command:

```bash
aift operator check      # Runs verify + architecture + runtime scan + summary
```

Readiness is intentionally NOT part of `verify`. It is a meta-analysis layer
that reads from registries populated by verify's sub-commands. Running it
inside verify would create a circular dependency or produce stale results.

## Output Files

| File | Format | Content |
|---|---|---|
| `registry/runtime-readiness.json` | JSON | All objects with status, evidence, summary |
| `reports/runtime-readiness.md` | Markdown | Human-readable report with tables |

## Running Tests

```bash
# Unit tests for status model, transitions, scanners
go test ./internal/readiness/ -v

# Integration tests for CLI commands
go test ./tests/integration/ -v -run TestRuntime

# Integration tests for operator check
go test ./tests/integration/ -v -run TestOperator
```
