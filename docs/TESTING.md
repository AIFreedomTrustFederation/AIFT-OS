# Testing

## Running Tests

```bash
go test ./...
```

Run with verbose output:

```bash
go test ./... -v
```

Run a specific package:

```bash
go test ./internal/modules/...
```

## Coverage

Generate a coverage report:

```bash
bash scripts/coverage.sh
```

Or manually:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out > reports/coverage.txt
```

View HTML coverage in a browser:

```bash
go tool cover -html=coverage.out
```

The coverage report is written to `reports/coverage.txt`.

## Coverage Baseline

A coverage threshold gate prevents regressions. The baseline is stored in
`coverage-baseline.txt` at the repo root. CI runs `scripts/check-coverage.sh`
which fails if total coverage drops below the baseline.

To update the baseline intentionally (e.g. after adding tests), edit
`coverage-baseline.txt` with the new percentage.

```bash
# Check coverage against baseline locally
bash scripts/check-coverage.sh
```

## Packages With Tests

| Package | Tests | Coverage |
|---|---:|---:|
| `internal/config` | 4 | 100.0% |
| `internal/fsutil` | 9 | 100.0% |
| `internal/sliceutil` | 13 | 100.0% |
| `internal/events` | 9 | 86.7% |
| `internal/jsonfile` | 11 | 86.2% |
| `internal/eventbus` | 15 | 53.4% |
| `internal/manifests` | 14 | 46.5% |
| `internal/discoveryengine` | 12 | 45.7% |
| `internal/capabilities` | 24 | 45.4% |
| `internal/intelligence` | 14 | 42.1% |
| `internal/modules` | 18 | 38.6% |
| `internal/kernelregistry` | 18 | 36.4% |
| `internal/planner` | 12 | 32.0% |
| `internal/patchengine` | 8 | 27.8% |
| `internal/repo` | 7 | 17.7% |
| `internal/eventmesh` | 18 | 17.0% |
| `internal/servicecontracts` | 16 | 13.3% |
| `internal/api` | 11 | 11.7% |
| `internal/graph` | 15 | 11.0% |
| `internal/readiness` | 35 | — |

## Packages Needing Coverage

The following packages have 0% test coverage:

- `internal/daemon`
- `internal/doctor`
- `internal/federation`
- `internal/gitx`
- `internal/jobs`
- `internal/kernel`
- `internal/kernelruntime`
- `internal/manual`
- `internal/plugins`
- `internal/providers`
- `internal/registry`
- `internal/reports`
- `internal/runtime`
- `internal/scheduler`
- `internal/services`
- `internal/state`
- `internal/supervisor`
- `internal/sync`
- `internal/workflow`
- `internal/workspace`

## Test Design

All tests are self-contained:

- Use `t.TempDir()` for filesystem isolation (auto-cleaned)
- No network access required
- No GitHub credentials or external services needed
- No real federation repositories required

## Integration Tests

CLI integration tests live in `tests/integration/` and exercise the compiled
`aiftd` binary end-to-end via `os/exec`. Each test creates an isolated workspace
using `t.TempDir()` with a minimal fake repo structure.

```bash
go test ./tests/integration/ -v
```

Command families covered:
- `help`, `version` (output format)
- `doctor` (health check)
- `status` (repo listing)
- `manifest` (creates `.aift/repo.json`)
- `registry` (generates `registry/repos.json`)
- `events`, `event-bus` (publish, list, report)
- `capabilities` (scan, report)
- `modules` (scan, list, init-all)
- `graph` (federation graph)
- `runtime` (scan, status, ready, blocked, report)
- `verify` (validation)

Contract tests:
- Every command in `help` has a matching `case` in `main.go`
- Every `case` in `main.go` appears in `help` output
- Planned commands return "planned" error messages
- No command panics on empty args
- No duplicate commands in help output
- Every `.sh` file passes `bash -n`
- Non-exempt scripts under `scripts/` source `aift-run.sh`

## Shell Harness Tests

The script execution harness has its own test suite:

```bash
bash tests/harness-syntax-test.sh
```

This validates 13 scenarios including syntax checks, path discovery, mode
queries, run directory creation, upload.txt timing, failure trap behavior,
inspect-only immutability, and commit-verified refusal on validation failure.
