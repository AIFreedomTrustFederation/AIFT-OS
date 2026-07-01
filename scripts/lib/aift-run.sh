#!/usr/bin/env bash
# scripts/lib/aift-run.sh
# Standard execution harness for AIFT-OS scripts.
#
# Usage:
#   source "$(dirname "$0")/lib/aift-run.sh"
#   aift_run <mode> <script-name> <function-to-run>
#
# Modes:
#   inspect-only     - read-only analysis, no changes
#   dry-run          - show what would happen, no changes
#   apply-local      - apply changes locally, no commit
#   commit-verified  - apply and commit if validation passes
#
# This harness:
#   - Discovers repo/workspace paths via scripts/lib/discovery.sh
#   - Creates a run directory under reports/runs/<timestamp>-<script>/
#   - Writes structured logs: terminal.log, environment.txt, git-before.txt,
#     git-after.txt, generated-files.txt, failure-analysis.txt, upload.txt
#   - Creates latest/upload.txt before risky operations
#   - Traps failures and always updates latest/upload.txt
#   - Only commits when mode=commit-verified and validation passes

set -euo pipefail

HARNESS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib/discovery.sh
source "${HARNESS_DIR}/discovery.sh"

# ── State ─────────────────────────────────────────────────────────────

_AIFT_RUN_DIR=""
_AIFT_LATEST_DIR=""
_AIFT_SCRIPT_NAME=""
_AIFT_MODE=""
_AIFT_STARTED=false
_AIFT_STATUS="pending"

# ── Logging ───────────────────────────────────────────────────────────

_log() {
  local msg="[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $*"
  echo "$msg"
  if [ -n "$_AIFT_RUN_DIR" ] && [ -d "$_AIFT_RUN_DIR" ]; then
    echo "$msg" >> "${_AIFT_RUN_DIR}/terminal.log"
  fi
}

# ── Capture helpers ───────────────────────────────────────────────────

_write_environment() {
  {
    echo "AIFT_ROOT=${AIFT_ROOT}"
    echo "AIFT_OS_HOME=${AIFT_OS_HOME}"
    echo "REPO_ROOT=${REPO_ROOT}"
    echo "MODE=${_AIFT_MODE}"
    echo "SCRIPT=${_AIFT_SCRIPT_NAME}"
    echo "USER=$(whoami)"
    echo "HOSTNAME=$(hostname)"
    echo "DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "GO_VERSION=$(go version 2>/dev/null || echo 'not installed')"
    echo "GIT_VERSION=$(git --version 2>/dev/null || echo 'not installed')"
    echo "BASH_VERSION=${BASH_VERSION}"
  } > "${_AIFT_RUN_DIR}/environment.txt"
}

_write_git_state() {
  local out="$1"
  (
    cd "$REPO_ROOT"
    {
      echo "# Branch"
      git branch --show-current 2>/dev/null || echo "(detached)"
      echo ""
      echo "# Status"
      git status --short 2>/dev/null || echo "(not a git repo)"
      echo ""
      echo "# HEAD"
      git log -1 --oneline 2>/dev/null || echo "(no commits)"
    } > "$out"
  )
}

_write_generated_files() {
  if [ -d "$_AIFT_RUN_DIR" ]; then
    find "$_AIFT_RUN_DIR" -type f | sort > "${_AIFT_RUN_DIR}/generated-files.txt"
  fi
}

_write_upload() {
  local status="${1:-$_AIFT_STATUS}"
  {
    echo "script=${_AIFT_SCRIPT_NAME}"
    echo "mode=${_AIFT_MODE}"
    echo "status=${status}"
    echo "run_dir=${_AIFT_RUN_DIR}"
    echo "timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  } > "${_AIFT_RUN_DIR}/upload.txt"

  # Also write to latest/
  mkdir -p "$_AIFT_LATEST_DIR"
  cp "${_AIFT_RUN_DIR}/upload.txt" "${_AIFT_LATEST_DIR}/upload.txt"
}

# ── Trap handler ──────────────────────────────────────────────────────

_on_exit() {
  local exit_code=$?
  if [ "$_AIFT_STARTED" != "true" ]; then
    return
  fi

  if [ $exit_code -ne 0 ]; then
    _AIFT_STATUS="failed"
    {
      echo "Exit code: $exit_code"
      echo "Script: $_AIFT_SCRIPT_NAME"
      echo "Mode: $_AIFT_MODE"
      echo "Time: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
      echo ""
      echo "Check terminal.log for details."
    } > "${_AIFT_RUN_DIR}/failure-analysis.txt"
    _log "FAILED with exit code $exit_code"
  else
    _AIFT_STATUS="success"
    _log "Completed successfully"
  fi

  _write_git_state "${_AIFT_RUN_DIR}/git-after.txt"
  _write_generated_files
  _write_upload "$_AIFT_STATUS"
}

# ── Validation ────────────────────────────────────────────────────────

aift_validate() {
  _log "Running validation..."
  local fail=0

  _log "  go test ./..."
  if ! (cd "$REPO_ROOT" && go test ./... >> "${_AIFT_RUN_DIR}/terminal.log" 2>&1); then
    _log "  FAIL: go test"
    fail=1
  fi

  _log "  go build -o bin/aift ./cmd/aift"
  if ! (cd "$REPO_ROOT" && mkdir -p bin && go build -o bin/aift ./cmd/aift >> "${_AIFT_RUN_DIR}/terminal.log" 2>&1); then
    _log "  FAIL: go build"
    fail=1
  fi

  if command -v go >/dev/null 2>&1; then
    _log "  ./bin/aift status"
    if ! (cd "$REPO_ROOT" && ./bin/aift status >> "${_AIFT_RUN_DIR}/terminal.log" 2>&1); then
      _log "  WARN: aift status returned non-zero (non-blocking)"
    fi

    _log "  ./bin/aift verify"
    if ! (cd "$REPO_ROOT" && ./bin/aift verify >> "${_AIFT_RUN_DIR}/terminal.log" 2>&1); then
      _log "  WARN: aift verify returned non-zero (non-blocking)"
    fi

    _log "  ./bin/aift registry"
    if ! (cd "$REPO_ROOT" && ./bin/aift registry >> "${_AIFT_RUN_DIR}/terminal.log" 2>&1); then
      _log "  WARN: aift registry returned non-zero (non-blocking)"
    fi
  fi

  _log "  bash -n (shell syntax check)"
  local sh_fail=0
  for f in "$REPO_ROOT"/tests/*.sh "$REPO_ROOT"/scripts/**/*.sh "$REPO_ROOT"/*.sh; do
    [ -f "$f" ] || continue
    if ! bash -n "$f" 2>> "${_AIFT_RUN_DIR}/terminal.log"; then
      _log "  FAIL: bash -n $f"
      sh_fail=1
    fi
  done
  if [ $sh_fail -ne 0 ]; then
    fail=1
  fi

  if [ $fail -ne 0 ]; then
    _log "Validation FAILED"
    return 1
  fi

  _log "Validation PASSED"
  return 0
}

# ── Commit helper ─────────────────────────────────────────────────────

aift_commit() {
  local message="${1:-aift-run: automated commit}"

  if [ "$_AIFT_MODE" != "commit-verified" ]; then
    _log "Skipping commit (mode=$_AIFT_MODE)"
    return 0
  fi

  if ! aift_validate; then
    _log "Commit aborted: validation failed"
    return 1
  fi

  (
    cd "$REPO_ROOT"
    if [ -z "$(git status --porcelain)" ]; then
      _log "Nothing to commit (working tree clean)"
      return 0
    fi

    git add -A
    git commit -m "$message"
    _log "Committed: $message"
  )
}

# ── Main entry point ─────────────────────────────────────────────────

aift_run() {
  local mode="${1:?Usage: aift_run <mode> <script-name> <function>}"
  local script_name="${2:?Usage: aift_run <mode> <script-name> <function>}"
  local func="${3:?Usage: aift_run <mode> <script-name> <function>}"

  case "$mode" in
    inspect-only|dry-run|apply-local|commit-verified) ;;
    *) echo "Unknown mode: $mode (use inspect-only|dry-run|apply-local|commit-verified)" >&2; return 1 ;;
  esac

  _AIFT_MODE="$mode"
  _AIFT_SCRIPT_NAME="$script_name"

  local timestamp
  timestamp="$(date -u +%Y%m%d-%H%M%S)"
  _AIFT_RUN_DIR="${REPO_ROOT}/reports/runs/${timestamp}-${script_name}"
  _AIFT_LATEST_DIR="${REPO_ROOT}/reports/runs/latest"

  mkdir -p "$_AIFT_RUN_DIR"
  mkdir -p "$_AIFT_LATEST_DIR"

  # Write latest/upload.txt immediately before any risky operation
  _write_upload "running"

  trap _on_exit EXIT

  _AIFT_STARTED=true
  _log "Starting: $script_name (mode=$mode)"

  _write_environment
  _write_git_state "${_AIFT_RUN_DIR}/git-before.txt"

  # Execute the user function
  "$func"

  # Write post-run state (trap also writes on failure)
  _AIFT_STATUS="success"
  _log "Completed successfully"
  _write_git_state "${_AIFT_RUN_DIR}/git-after.txt"
  _write_generated_files
  _write_upload "success"
  _AIFT_STARTED=false
}

# ── Mode queries ──────────────────────────────────────────────────────

aift_is_readonly() {
  [ "$_AIFT_MODE" = "inspect-only" ] || [ "$_AIFT_MODE" = "dry-run" ]
}

aift_should_apply() {
  [ "$_AIFT_MODE" = "apply-local" ] || [ "$_AIFT_MODE" = "commit-verified" ]
}

aift_mode() {
  echo "$_AIFT_MODE"
}

aift_run_dir() {
  echo "$_AIFT_RUN_DIR"
}
