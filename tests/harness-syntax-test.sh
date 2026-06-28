#!/usr/bin/env bash
# tests/harness-syntax-test.sh
# Verify the aift-run.sh harness and discovery.sh pass shell syntax checks
# and can be sourced without error.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== Harness syntax tests ==="

echo "1. bash -n scripts/lib/discovery.sh"
bash -n "$REPO_ROOT/scripts/lib/discovery.sh"
echo "   PASS"

echo "2. bash -n scripts/lib/aift-run.sh"
bash -n "$REPO_ROOT/scripts/lib/aift-run.sh"
echo "   PASS"

echo "3. bash -n scripts/coverage.sh"
bash -n "$REPO_ROOT/scripts/coverage.sh"
echo "   PASS"

echo "4. Source discovery.sh and verify exported vars"
(
  export REPO_ROOT="$REPO_ROOT"
  source "$REPO_ROOT/scripts/lib/discovery.sh"
  [ -n "$AIFT_ROOT" ] || { echo "FAIL: AIFT_ROOT not set"; exit 1; }
  [ -n "$AIFT_OS_HOME" ] || { echo "FAIL: AIFT_OS_HOME not set"; exit 1; }
  [ -n "$REPO_ROOT" ] || { echo "FAIL: REPO_ROOT not set"; exit 1; }
  echo "   PASS (AIFT_ROOT=$AIFT_ROOT, REPO_ROOT=$REPO_ROOT)"
)

echo "5. Source aift-run.sh and verify functions exist"
(
  export REPO_ROOT="$REPO_ROOT"
  source "$REPO_ROOT/scripts/lib/aift-run.sh"
  type aift_run >/dev/null 2>&1 || { echo "FAIL: aift_run not defined"; exit 1; }
  type aift_validate >/dev/null 2>&1 || { echo "FAIL: aift_validate not defined"; exit 1; }
  type aift_commit >/dev/null 2>&1 || { echo "FAIL: aift_commit not defined"; exit 1; }
  type aift_is_readonly >/dev/null 2>&1 || { echo "FAIL: aift_is_readonly not defined"; exit 1; }
  type aift_should_apply >/dev/null 2>&1 || { echo "FAIL: aift_should_apply not defined"; exit 1; }
  type aift_mode >/dev/null 2>&1 || { echo "FAIL: aift_mode not defined"; exit 1; }
  type aift_run_dir >/dev/null 2>&1 || { echo "FAIL: aift_run_dir not defined"; exit 1; }
  echo "   PASS"
)

echo "6. Test inspect-only mode creates run directory and logs"
(
  export REPO_ROOT="$REPO_ROOT"
  source "$REPO_ROOT/scripts/lib/aift-run.sh"

  test_inspect() {
    _log "Inspect-only test run"
    if ! aift_is_readonly; then
      echo "FAIL: inspect-only should be readonly"
      exit 1
    fi
    if aift_should_apply; then
      echo "FAIL: inspect-only should not apply"
      exit 1
    fi
  }

  aift_run "inspect-only" "harness-test" test_inspect

  run_dir=$(ls -d "$REPO_ROOT"/reports/runs/*-harness-test 2>/dev/null | tail -1)
  [ -f "$run_dir/terminal.log" ] || { echo "FAIL: terminal.log missing"; exit 1; }
  [ -f "$run_dir/environment.txt" ] || { echo "FAIL: environment.txt missing"; exit 1; }
  [ -f "$run_dir/git-before.txt" ] || { echo "FAIL: git-before.txt missing"; exit 1; }
  [ -f "$run_dir/git-after.txt" ] || { echo "FAIL: git-after.txt missing"; exit 1; }
  [ -f "$run_dir/upload.txt" ] || { echo "FAIL: upload.txt missing"; exit 1; }
  [ -f "$run_dir/generated-files.txt" ] || { echo "FAIL: generated-files.txt missing"; exit 1; }
  [ -f "$REPO_ROOT/reports/runs/latest/upload.txt" ] || { echo "FAIL: latest/upload.txt missing"; exit 1; }

  grep -q "status=success" "$run_dir/upload.txt" || { echo "FAIL: upload.txt should show success"; exit 1; }
  echo "   PASS"
)

echo "7. Test dry-run mode is readonly"
(
  export REPO_ROOT="$REPO_ROOT"
  source "$REPO_ROOT/scripts/lib/aift-run.sh"

  test_dryrun() {
    if ! aift_is_readonly; then
      echo "FAIL: dry-run should be readonly"
      exit 1
    fi
  }

  aift_run "dry-run" "dryrun-test" test_dryrun
  echo "   PASS"
)

echo "8. Test apply-local mode should-apply"
(
  export REPO_ROOT="$REPO_ROOT"
  source "$REPO_ROOT/scripts/lib/aift-run.sh"

  test_apply() {
    if aift_is_readonly; then
      echo "FAIL: apply-local should not be readonly"
      exit 1
    fi
    if ! aift_should_apply; then
      echo "FAIL: apply-local should apply"
      exit 1
    fi
  }

  aift_run "apply-local" "apply-test" test_apply
  echo "   PASS"
)

echo "9. Test invalid mode rejected"
(
  export REPO_ROOT="$REPO_ROOT"
  source "$REPO_ROOT/scripts/lib/aift-run.sh"
  if aift_run "bad-mode" "test" echo 2>/dev/null; then
    echo "FAIL: bad mode should be rejected"
    exit 1
  fi
  echo "   PASS"
)

echo "10. Test latest/upload.txt is created before user function runs"
(
  export REPO_ROOT="$REPO_ROOT"
  source "$REPO_ROOT/scripts/lib/aift-run.sh"

  test_upload_timing() {
    # By the time the user function runs, latest/upload.txt should exist
    local latest="$REPO_ROOT/reports/runs/latest/upload.txt"
    [ -f "$latest" ] || { echo "FAIL: latest/upload.txt not created before execution"; exit 1; }
    grep -q "status=running" "$latest" || { echo "FAIL: latest/upload.txt should show status=running"; exit 1; }
  }

  aift_run "inspect-only" "upload-timing-test" test_upload_timing
  echo "   PASS"
)

echo "11. Test failure trap updates upload.txt on failure"
# Run the failing harness in a nested bash so we can inspect files after it exits
bash -c '
  set -euo pipefail
  export REPO_ROOT="'"$REPO_ROOT"'"
  source "$REPO_ROOT/scripts/lib/aift-run.sh"
  test_failure() { _log "About to fail"; return 1; }
  aift_run "inspect-only" "failure-trap-test" test_failure
' 2>/dev/null && {
  echo "FAIL: expected non-zero exit from failing function"
  exit 1
} || true
run_dir=$(ls -d "$REPO_ROOT"/reports/runs/*-failure-trap-test 2>/dev/null | tail -1)
[ -f "$run_dir/failure-analysis.txt" ] || { echo "FAIL: failure-analysis.txt not created"; exit 1; }
[ -f "$run_dir/upload.txt" ] || { echo "FAIL: upload.txt missing after failure"; exit 1; }
grep -q "status=failed" "$run_dir/upload.txt" || { echo "FAIL: upload.txt should show status=failed"; exit 1; }
[ -f "$REPO_ROOT/reports/runs/latest/upload.txt" ] || { echo "FAIL: latest/upload.txt missing after failure"; exit 1; }
grep -q "status=failed" "$REPO_ROOT/reports/runs/latest/upload.txt" || { echo "FAIL: latest/upload.txt should show status=failed"; exit 1; }
echo "   PASS"

echo "12. Test inspect-only mode never mutates working tree"
(
  export REPO_ROOT="$REPO_ROOT"
  source "$REPO_ROOT/scripts/lib/aift-run.sh"

  # Capture git status before
  git_before=$(cd "$REPO_ROOT" && git status --porcelain -- internal/ cmd/ scripts/lib/ 2>/dev/null)

  test_inspect_immutable() {
    if aift_should_apply; then
      echo "FAIL: inspect-only should not allow apply"
      exit 1
    fi
    if ! aift_is_readonly; then
      echo "FAIL: inspect-only should be readonly"
      exit 1
    fi
    # Verify mode string
    local m
    m=$(aift_mode)
    [ "$m" = "inspect-only" ] || { echo "FAIL: mode = $m, want inspect-only"; exit 1; }
  }

  aift_run "inspect-only" "immutable-test" test_inspect_immutable

  # Capture git status after
  git_after=$(cd "$REPO_ROOT" && git status --porcelain -- internal/ cmd/ scripts/lib/ 2>/dev/null)

  if [ "$git_before" != "$git_after" ]; then
    echo "FAIL: inspect-only mode changed working tree"
    echo "  before: $git_before"
    echo "  after:  $git_after"
    exit 1
  fi
  echo "   PASS"
)

echo "13. Test commit-verified refuses commit when validation would fail"
(
  export REPO_ROOT="$REPO_ROOT"
  source "$REPO_ROOT/scripts/lib/aift-run.sh"

  test_commit_refuses() {
    # aift_commit in commit-verified mode calls aift_validate first.
    # We override aift_validate to simulate failure.
    aift_validate() { _log "Simulated validation failure"; return 1; }
    if aift_commit "should not commit" 2>/dev/null; then
      echo "FAIL: commit should fail when validation fails"
      exit 1
    fi
    _log "commit correctly refused"
  }

  aift_run "commit-verified" "commit-refuse-test" test_commit_refuses
  echo "   PASS"
)

# Clean up test run directories
rm -rf "$REPO_ROOT"/reports/runs/*-harness-test "$REPO_ROOT"/reports/runs/*-dryrun-test \
  "$REPO_ROOT"/reports/runs/*-apply-test "$REPO_ROOT"/reports/runs/*-upload-timing-test \
  "$REPO_ROOT"/reports/runs/*-failure-trap-test "$REPO_ROOT"/reports/runs/*-immutable-test \
  "$REPO_ROOT"/reports/runs/*-commit-refuse-test "$REPO_ROOT"/reports/runs/latest

echo ""
echo "=== All harness tests passed ==="
