# Scheduler Plan

AIFT-OS generates a truthful execution plan from runtime readiness data. The scheduler is **planning only** — it does not execute jobs.

## How It Works

The scheduler consumes:
- `registry/runtime-readiness.json` — all evaluated federation objects
- `registry/architecture.json` — command handler/help status (optional)

It produces:
- `registry/scheduler-plan.json` — machine-readable plan
- `reports/scheduler-plan.md` — human-readable report

## Job Statuses

| Status | Meaning |
|---|---|
| `pending` | Not yet evaluated or requires prerequisites |
| `runnable` | All prerequisites met; can be executed |
| `blocked` | Cannot proceed; missing dependency or prerequisite not ready |
| `running` | Currently executing (not used in planning phase) |
| `succeeded` | Completed successfully (not used in planning phase) |
| `failed` | Execution failed (not used in planning phase) |
| `skipped` | Deprecated or removed; will not execute |

## Dependency Logic

- **Commands and scripts** are blocked if no repository is runnable
- **Capabilities** are blocked if their source repository is not runnable
- **Services** are blocked if their source repository is not runnable
- **Modules** are blocked if their source repository is not runnable
- **Planned commands** are always pending (no executable exists)
- The scheduler never pretends a command or script exists

## Status Mapping from Readiness

| Readiness Status | Scheduler Status |
|---|---|
| `ready` | `runnable` |
| `active` | `runnable` |
| `planned` | `pending` |
| `detected` | `pending` |
| `blocked` | `blocked` |
| `deprecated` | `skipped` |
| `removed` | `skipped` |

## CLI Commands

```
aift scheduler plan      # Build plan, write registry + report
aift scheduler ready     # Print only runnable jobs
aift scheduler blocked   # Print only blocked jobs
aift scheduler report    # Print the plan report (markdown)
```

## Recommended Workflow

```bash
aift verify              # Populate registries
aift runtime scan        # Evaluate readiness
aift scheduler plan      # Generate execution plan
aift scheduler report    # Review the plan
```

Or via the aggregate operator command:

```bash
aift operator check      # Runs verify + architecture + runtime scan
aift scheduler plan      # Generate plan from populated registries
```

## Running Tests

```bash
# Unit tests for job building, dependency resolution, status mapping
go test ./internal/schedulerplan/ -v

# Integration tests for CLI commands
go test ./tests/integration/ -v -run TestScheduler
```
