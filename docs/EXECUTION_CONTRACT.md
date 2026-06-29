# Script Execution Contract

## Overview

All active shell scripts in AIFT-OS follow a standardized execution contract
via `scripts/lib/aift-run.sh`. This harness provides structured logging,
run directories, failure trapping, and mode-based execution control.

## Usage

Source the harness and call `aift_run`:

```bash
#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/lib/aift-run.sh"

do_work() {
  # your script logic here
  if aift_is_readonly; then
    echo "Read-only mode, skipping changes"
    return
  fi
  # apply changes...
}

aift_run "inspect-only" "my-script" do_work
```

## Execution Modes

| Mode | Reads | Writes | Commits |
|---|---|---|---|
| `inspect-only` | yes | no | no |
| `dry-run` | yes | no | no |
| `apply-local` | yes | yes | no |
| `commit-verified` | yes | yes | yes (if validation passes) |

## Mode Query Functions

- `aift_is_readonly` - true for `inspect-only` and `dry-run`
- `aift_should_apply` - true for `apply-local` and `commit-verified`
- `aift_mode` - prints the current mode string
- `aift_run_dir` - prints the path to the current run directory

## Run Directory Structure

Each execution creates a timestamped directory:

```
reports/runs/<YYYYMMDD-HHMMSS>-<script-name>/
  terminal.log            # all log output
  environment.txt         # env vars, tool versions
  git-before.txt          # git state before execution
  git-after.txt           # git state after execution
  generated-files.txt     # list of files in run dir
  failure-analysis.txt    # only written on failure
  upload.txt              # status summary
```

A `reports/runs/latest/upload.txt` symlink-like copy is maintained with the
most recent run status.

## Path Discovery

The harness sources `scripts/lib/discovery.sh`, which exports:

- `REPO_ROOT` - root of the current repository (auto-detected from `.git` or `go.mod`)
- `AIFT_ROOT` - federation workspace root (defaults to `$HOME/AIFT`)
- `AIFT_OS_HOME` - AIFT-OS installation (defaults to `$AIFT_ROOT/AIFT-OS`)

All paths can be overridden via environment variables.

## Validation

Call `aift_validate` to run the standard validation suite:

1. `go test ./...`
2. `go build ./cmd/aift`
3. `go run ./cmd/aift doctor`
4. `go run ./cmd/aift introspect scan`
5. `go run ./cmd/aift introspect check`
6. `bash -n` for all shell scripts

## Commit Behavior

The `aift_commit "message"` function:

1. Only acts in `commit-verified` mode (no-op otherwise)
2. Runs full validation first
3. Aborts if validation fails
4. Skips if the working tree is clean
5. Stages all changes and commits

## Failure Handling

The harness traps `EXIT` and:

- Writes `failure-analysis.txt` with exit code and context
- Updates `upload.txt` with `status=failed`
- Copies status to `latest/upload.txt`
- Always captures `git-after.txt` state

## Operator Workflow

The recommended operator workflow runs all verification steps in sequence:

```bash
aift verify              # Doctor + manifests + registries + graphs
aift runtime scan        # Evaluate all objects for readiness
aift runtime report      # Print readiness report
```

The `aift operator check` command runs this entire sequence plus architecture
validation in a single invocation:

```bash
aift operator check
```

Steps executed:
1. `verify` - doctor, manifests, registries, capabilities, intelligence, graph, mesh, service-contracts, planner
2. `architecture` - runs `go run ./tools/architecture --ci` to check invariants
3. `runtime scan` - evaluates all federation objects
4. Readiness summary - prints status table

Each step's success or failure is reported. The command fails if any required
step fails, and failures are surfaced in the output (never swallowed).

## Coverage Baseline

The repository enforces a coverage threshold via `scripts/check-coverage.sh`.
The baseline percentage is stored in `coverage-baseline.txt` at the repo root.
CI runs this check automatically; if coverage drops below the baseline the
build fails.

To update the baseline after adding tests:

```bash
echo "30.0" > coverage-baseline.txt
```
