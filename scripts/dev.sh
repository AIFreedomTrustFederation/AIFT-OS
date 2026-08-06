#!/usr/bin/env bash
# no-harness: developer verification orchestrator; mutations are limited to declared build/report outputs and optional Git hook installation.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
KEEP_ARTIFACTS="${AIFT_KEEP_ARTIFACTS:-0}"

cd "$REPO_ROOT"

say() {
  printf '\n==> %s\n' "$*"
}

require() {
  command -v "$1" >/dev/null 2>&1 || {
    printf 'ERROR: required command not found: %s\n' "$1" >&2
    exit 1
  }
}

format() {
  say "Formatting Go source"
  gofmt -w cmd internal tools
}

check_format() {
  say "Checking Go formatting"
  local unformatted
  unformatted="$(gofmt -l cmd internal tools)"
  if [[ -n "$unformatted" ]]; then
    printf 'ERROR: these files need gofmt:\n%s\n' "$unformatted" >&2
    return 1
  fi
}

check_shell() {
  say "Checking shell syntax"
  local status=0
  while IFS= read -r -d '' file; do
    if ! bash -n "$file"; then
      printf 'FAIL: %s\n' "$file" >&2
      status=1
    fi
  done < <(find . -type f -name '*.sh'     -not -path './.git/*'     -not -path './node_modules/*'     -print0)
  return "$status"
}

test_all() {
  say "Running all Go tests"
  go test ./...
}

build_to() {
  local output_dir="$1"
  mkdir -p "$output_dir"
  go build -o "$output_dir/aift" ./cmd/aift
  go build -o "$output_dir/aiftd" ./cmd/aiftd
}

build() {
  say "Building commands into bin/"
  build_to "$REPO_ROOT/bin"
}

check_build() {
  say "Building commands in an isolated directory"
  local build_root
  build_root="$(mktemp -d "${TMPDIR:-/tmp}/aift-build.XXXXXX")"
  trap 'rm -rf "$build_root"' RETURN
  build_to "$build_root"
  [[ -x "$build_root/aift" && -x "$build_root/aiftd" ]]
  rm -rf "$build_root"
  trap - RETURN
}

smoke() {
  say "Running MoBox UXI smoke test"
  bash scripts/smoke-mobox-uxi.sh
}

coverage() {
  say "Checking coverage threshold"
  bash scripts/check-coverage.sh
  if [[ "$KEEP_ARTIFACTS" != "1" ]]; then
    rm -f coverage.out
  fi
}

architecture() {
  say "Checking architecture invariants"
  if [[ "$KEEP_ARTIFACTS" == "1" ]]; then
    go run ./tools/architecture --ci
    return
  fi

  local copy_root
  copy_root="$(mktemp -d "${TMPDIR:-/tmp}/aift-architecture.XXXXXX")"
  trap 'rm -rf "$copy_root"' RETURN
  tar -cf - --exclude='./.git' --exclude='./bin' --exclude='./coverage.out' . |
    tar -xf - -C "$copy_root"
  (
    cd "$copy_root"
    go run ./tools/architecture --ci
  )
  rm -rf "$copy_root"
  trap - RETURN
}

check_artifacts() {
  say "Checking repository artifact policy"
  local forbidden
  forbidden="$(git ls-files | awk '
    $0 == "aift" || $0 == "aiftd" || $0 ~ /^bin\// || $0 == "coverage.out" { print }
  ')"
  if [[ -n "$forbidden" ]]; then
    printf 'ERROR: compiled/generated artifacts are tracked:\n%s\n' "$forbidden" >&2
    return 1
  fi
}

install_hooks() {
  say "Installing opt-in pre-push hook"
  local source="$REPO_ROOT/hooks/pre-push"
  local target="$REPO_ROOT/.git/hooks/pre-push"
  [[ -d "$REPO_ROOT/.git" ]] || {
    printf 'ERROR: install-hooks requires a Git worktree\n' >&2
    return 1
  }
  if [[ -e "$target" ]] && ! cmp -s "$source" "$target"; then
    printf 'ERROR: %s already exists and differs; preserve or remove it manually\n' "$target" >&2
    return 1
  fi
  cp "$source" "$target"
  chmod +x "$target"
  printf 'Installed %s\n' "$target"
}

verify_fast() {
  check_format
  check_shell
  test_all
  check_build
  check_artifacts
}

verify() {
  verify_fast
  smoke
  coverage
  architecture
  say "Verification passed"
}

usage() {
  cat <<'EOF'
Usage: bash scripts/dev.sh <command>

Commands:
  format         Apply gofmt to cmd/, internal/, and tools/
  check-format   Fail when Go source is not formatted
  shell          Check every maintained shell script with bash -n
  test           Run go test ./...
  build          Build bin/aift and bin/aiftd
  build-check    Build both commands in an isolated temporary directory
  smoke          Run the isolated MoBox UXI smoke test
  coverage       Enforce the stored coverage baseline
  architecture   Validate architecture without dirtying the local worktree
  artifacts      Reject tracked binaries and generated coverage output
  verify-fast    Formatting, shell, tests, isolated builds, artifact policy
  verify         Complete local verification
  install-hooks  Install the repository's optional pre-push hook
EOF
}

require bash
require go
require git
require tar

case "${1:-}" in
  format) format ;;
  check-format) check_format ;;
  shell) check_shell ;;
  test) test_all ;;
  build) build ;;
  build-check) check_build ;;
  smoke) smoke ;;
  coverage) coverage ;;
  architecture) architecture ;;
  artifacts) check_artifacts ;;
  verify-fast) verify_fast ;;
  verify) verify ;;
  install-hooks) install_hooks ;;
  -h|--help|help|"") usage ;;
  *)
    printf 'ERROR: unknown command: %s\n' "$1" >&2
    usage >&2
    exit 2
    ;;
esac
