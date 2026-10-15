#!/usr/bin/env bash
set -euo pipefail

WORKSPACE="${1:-$HOME/AIFT}"

if [[ ! -d "$WORKSPACE" ]]; then
  printf 'FAIL: workspace directory not found: %s\n' "$WORKSPACE" >&2
  exit 1
fi

repositories=0
failures=0

for repo in "$WORKSPACE"/*; do
  [[ -d "$repo/.git" ]] || continue
  repositories=$((repositories + 1))

  # Continue inspecting other repositories even when one check fails.
  (
    cd "$repo"
    printf '\n======================================\nVerifying %s\n======================================\n' "${repo##*/}"
    status=0

    run_check() {
      local label="$1"
      shift
      if "$@"; then
        printf 'PASS: %s\n' "$label"
      else
        printf 'FAIL: %s\n' "$label" >&2
        status=1
      fi
    }

    run_check 'git status' git status --short

    if [[ -f package.json ]]; then
      if command -v npm >/dev/null 2>&1; then
        run_check 'npm lint (if present)' npm run lint --if-present
        run_check 'npm typecheck (if present)' npm run typecheck --if-present
        run_check 'npm test (if present)' npm run test --if-present
        run_check 'npm build (if present)' npm run build --if-present
      else
        printf 'FAIL: npm unavailable for package.json\n' >&2
        status=1
      fi
    fi

    if [[ -f Cargo.toml ]]; then
      if command -v cargo >/dev/null 2>&1; then
        run_check 'cargo check' cargo check
      else
        printf 'FAIL: cargo unavailable for Cargo.toml\n' >&2
        status=1
      fi
    fi

    if [[ -f go.mod ]]; then
      if command -v go >/dev/null 2>&1; then
        run_check 'go test' go test ./...
      else
        printf 'FAIL: go unavailable for go.mod\n' >&2
        status=1
      fi
    fi

    exit "$status"
  ) || failures=$((failures + 1))
done

if (( repositories == 0 )); then
  printf 'FAIL: no Git repositories found under %s\n' "$WORKSPACE" >&2
  exit 1
fi

if (( failures > 0 )); then
  printf 'FAIL: verification failed for %s of %s repositories\n' "$failures" "$repositories" >&2
  exit 1
fi

printf 'PASS: verification completed for %s repositories\n' "$repositories"
