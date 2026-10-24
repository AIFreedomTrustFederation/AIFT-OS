#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT="${AIFT_PUSH_SWEEP_SCRIPT:-$REPO_ROOT/push-all-aift-repos.sh}"
REAL_GIT="$(command -v git)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

mkdir -p "$TMP/mock-bin"
cat > "$TMP/mock-bin/git" <<'WRAPPER'
#!/usr/bin/env bash
set -euo pipefail
case "${1:-}" in
  pull|push)
    printf '%s %s\n' "${PWD##*/}" "$1" >> "$GIT_CALL_LOG"
    if [[ "${PWD##*/}" == "${INJECT_REPO:-}" && "$1" == "${INJECT_COMMAND:-}" ]]; then
      echo "injected $1 failure for ${PWD##*/}" >&2
      exit 42
    fi
    ;;
esac
exec "$REAL_GIT" "$@"
WRAPPER
chmod +x "$TMP/mock-bin/git"
export REAL_GIT

new_workspace() {
  local workspace="$1" name bare
  mkdir -p "$workspace"
  for name in first second; do
    bare="$TMP/${name}-${workspace##*/}.git"
    "$REAL_GIT" init -q --bare --initial-branch=main "$bare"
    "$REAL_GIT" init -q --initial-branch=main "$workspace/$name"
    "$REAL_GIT" -C "$workspace/$name" config user.email 'audit@example.invalid'
    "$REAL_GIT" -C "$workspace/$name" config user.name 'AIFT Audit'
    printf '%s\n' "$name" > "$workspace/$name/tracked.txt"
    "$REAL_GIT" -C "$workspace/$name" add tracked.txt
    "$REAL_GIT" -C "$workspace/$name" commit -qm initial
    "$REAL_GIT" -C "$workspace/$name" remote add origin "$bare"
    "$REAL_GIT" -C "$workspace/$name" push -q -u origin main
  done
  printf "unpublished update\n" >> "$workspace/first/tracked.txt"
}

run_scenario() {
  local name="$1" fail_repo="$2" fail_command="$3" expect_success="$4"
  local log="$TMP/$name.log" workspace="$TMP/workspace-$name"
  new_workspace "$workspace"
  : > "$TMP/$name.calls"
  local status=0
  AIFT_ROOT="$workspace" GIT_CALL_LOG="$TMP/$name.calls" INJECT_REPO="$fail_repo" INJECT_COMMAND="$fail_command" \
    PATH="$TMP/mock-bin:$PATH" sh "$SCRIPT" > "$log" 2>&1 || status=$?
  if [[ "$expect_success" == yes ]]; then
    if (( status != 0 )); then echo "FAIL: $name unexpectedly returned $status" >&2; cat "$log" >&2; exit 1; fi
    grep -q '^DONE\.$' "$log"
  else
    if (( status == 0 )); then echo "FAIL: $name incorrectly returned success" >&2; cat "$log" >&2; exit 1; fi
    if grep -q '^DONE\.$' "$log"; then echo "FAIL: $name reported DONE despite failure" >&2; exit 1; fi
    grep -q "injected $fail_command failure for $fail_repo" "$log"
    grep -q 'FAIL:' "$log"
  fi
  if [[ "$fail_command" == pull ]]; then
    if grep -q '^first push$' "$TMP/$name.calls"; then
      echo "FAIL: push attempted after failed pull" >&2; exit 1
    fi
  fi
  grep -q '^second push$' "$TMP/$name.calls" || { echo "FAIL: sweep did not continue to second repo" >&2; exit 1; }
  local remote_sha local_sha
  remote_sha="$("$REAL_GIT" --git-dir="$TMP/first-${workspace##*/}.git" rev-parse refs/heads/main)"
  local_sha="$("$REAL_GIT" -C "$workspace/first" rev-parse HEAD)"
  if [[ "$expect_success" == yes && "$remote_sha" != "$local_sha" ]]; then
    echo "FAIL: successful run did not publish first repo" >&2; exit 1
  elif [[ "$expect_success" == no && "$remote_sha" == "$local_sha" ]]; then
    echo "FAIL: failed run published first repo" >&2; exit 1
  fi
  echo "PASS: $name"
}

run_scenario successful '' '' yes
run_scenario failed-pull first pull no
run_scenario failed-push first push no
