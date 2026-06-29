#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" plan build >/dev/null
"$ROOT/aift" plan summary >/dev/null
"$ROOT/aift" plan repo AIFT-OS >/dev/null
"$ROOT/aift" plan ready >/dev/null
"$ROOT/aift" plan blocked >/dev/null
"$ROOT/aift" plan report >/dev/null
"$ROOT/aift" verify >/dev/null

test -f registry/execution-plan.json
test -f reports/execution-plan.md
test -f reports/execution-blockers.md

echo "OK: planner smoke passed"
