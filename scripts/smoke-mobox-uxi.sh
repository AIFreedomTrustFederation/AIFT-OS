#!/usr/bin/env bash
# End-to-end smoke test for the local MoBox UXI service.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SMOKE_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/aift-uxi-smoke.XXXXXX")"
PORT="${AIFT_UXI_SMOKE_PORT:-18787}"
BASE_URL="http://127.0.0.1:$PORT"
LOG_FILE="$SMOKE_ROOT/aiftd.log"
SERVER_PID=""

cleanup() {
  if [[ -n "$SERVER_PID" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill -TERM "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  rm -rf "$SMOKE_ROOT"
}
trap cleanup EXIT INT TERM

require_command() {
  command -v "$1" >/dev/null 2>&1 || {
    printf 'missing required command: %s\n' "$1" >&2
    exit 1
  }
}

require_command curl
require_command go
require_command python3

request() {
  local method="$1"
  local path="$2"
  local expected="$3"
  local body="${4:-}"
  local output="$SMOKE_ROOT/response.json"
  local status

  if [[ -n "$body" ]]; then
    status="$(curl --silent --show-error --output "$output" --write-out '%{http_code}'       --request "$method" --header 'Content-Type: application/json'       --data "$body" "$BASE_URL$path")"
  else
    status="$(curl --silent --show-error --output "$output" --write-out '%{http_code}'       --request "$method" "$BASE_URL$path")"
  fi

  if [[ "$status" != "$expected" ]]; then
    printf 'FAIL %s %s: expected HTTP %s, got %s\n' "$method" "$path" "$expected" "$status" >&2
    sed -n '1,80p' "$output" >&2
    exit 1
  fi
  printf 'PASS %s %s (%s)\n' "$method" "$path" "$status"
}

json_value() {
  local expression="$1"
  python3 - "$SMOKE_ROOT/response.json" "$expression" <<'PY'
import json
import sys

with open(sys.argv[1], encoding="utf-8") as handle:
    value = json.load(handle)
for key in sys.argv[2].split("."):
    if key:
        value = value[key]
print(value)
PY
}

cd "$REPO_ROOT"
AIFT_UXI_HOME="$SMOKE_ROOT/data" AIFT_UXI_ADDR="127.0.0.1:$PORT" go run ./cmd/aiftd >"$LOG_FILE" 2>&1 &
SERVER_PID="$!"

for _ in $(seq 1 60); do
  if curl --silent --fail "$BASE_URL/health" >/dev/null 2>&1; then
    break
  fi
  if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    printf 'aiftd exited during startup\n' >&2
    sed -n '1,120p' "$LOG_FILE" >&2
    exit 1
  fi
  if curl --silent --fail "$BASE_URL/health" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

if ! kill -0 "$SERVER_PID" 2>/dev/null; then
  printf 'aiftd is not running after startup\n' >&2
  sed -n '1,120p' "$LOG_FILE" >&2
  exit 1
fi

request GET /health 200
[[ "$(json_value status)" == "pass" ]] || { printf 'FAIL health status\n' >&2; exit 1; }

for path in /v1/system /v1/repositories /v1/adapters /v1/sources   /v1/federation/tree /v1/federation/world /v1/adapters/forge/mission /v1/sessions /v1/events; do
  request GET "$path" 200
done

for path in / /tree /world; do
  request GET "$path" 200
done

request POST /v1/sessions 201 '{"title":"MoBox UXI smoke test"}'
SESSION_ID="$(json_value id)"
[[ -n "$SESSION_ID" ]] || { printf 'FAIL session id missing\n' >&2; exit 1; }

request GET "/v1/sessions/$SESSION_ID" 200
request POST "/v1/sessions/$SESSION_ID/plans" 201   '{"objective":"Verify governed UXI flow","steps":[{"title":"Observe smoke-test evidence","status":"pending"}]}'
request POST "/v1/sessions/$SESSION_ID/actions" 201   '{"kind":"inspect","target":"AIFT-OS","risk":"low","approval_required":true,"parameters":{"source":"smoke-test"}}'

printf '\nMoBox UXI smoke test passed. Session: %s\n' "$SESSION_ID"
