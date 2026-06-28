#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" mesh init-all >/dev/null
"$ROOT/aift" mesh scan >/dev/null
"$ROOT/aift" mesh topics >/dev/null
"$ROOT/aift" mesh subscribers >/dev/null
"$ROOT/aift" mesh publish phase13.test tests "phase 13 smoke event" >/dev/null
"$ROOT/aift" mesh replay phase13.test >/dev/null
"$ROOT/aift" mesh tail >/dev/null
"$ROOT/aift" mesh report >/dev/null
"$ROOT/aift" verify >/dev/null

test -f registry/event-mesh.json
test -f reports/event-mesh.md
test -f .aift/events.json
test -d .aift/events/handlers

echo "OK: event mesh smoke passed"
