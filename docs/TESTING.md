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

## Packages With Tests

| Package | Tests | Coverage |
|---|---:|---:|
| `internal/fsutil` | 9 | 100.0% |
| `internal/sliceutil` | 13 | 100.0% |
| `internal/jsonfile` | 11 | 86.2% |
| `internal/eventbus` | 15 | 50.0% |
| `internal/discoveryengine` | 12 | 37.0% |
| `internal/modules` | 18 | 31.1% |
| `internal/kernelregistry` | 18 | 27.7% |
| `internal/planner` | 12 | 27.2% |
| `internal/patchengine` | 8 | 25.7% |
| `internal/intelligence` | 14 | 23.4% |
| `internal/eventmesh` | 18 | 19.4% |
| `internal/servicecontracts` | 14 | 10.8% |
| `internal/graph` | 15 | 9.7% |

## Packages Needing Coverage

The following packages have 0% test coverage:

- `internal/api`
- `internal/capabilities`
- `internal/config`
- `internal/daemon`
- `internal/doctor`
- `internal/events`
- `internal/federation`
- `internal/gitx`
- `internal/jobs`
- `internal/kernel`
- `internal/kernelruntime`
- `internal/manifests`
- `internal/manual`
- `internal/plugins`
- `internal/providers`
- `internal/registry`
- `internal/repo`
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
