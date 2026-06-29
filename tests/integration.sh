#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" version >/dev/null
"$ROOT/aift" doctor >/dev/null
"$ROOT/aift" status >/dev/null
"$ROOT/aift" manifest >/dev/null
"$ROOT/aift" registry >/dev/null
"$ROOT/aift" dashboard >/dev/null
"$ROOT/aift" deps >/dev/null
"$ROOT/aift" providers >/dev/null
"$ROOT/aift" events >/dev/null
"$ROOT/aift" services >/dev/null
"$ROOT/aift" start >/dev/null
"$ROOT/aift" tick >/dev/null
"$ROOT/aift" verify >/dev/null

test -f registry/repos.json
test -f registry/providers.json
test -f reports/dashboard.md
test -f reports/dependency-graph.md
test -f var/events/events.jsonl

echo "OK: integration test passed"
