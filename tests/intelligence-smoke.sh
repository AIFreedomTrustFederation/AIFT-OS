#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" intelligence scan >/dev/null
"$ROOT/aift" intelligence report >/dev/null
"$ROOT/aift" intelligence repo AIFT-OS >/dev/null
"$ROOT/aift" intelligence roadmap >/dev/null
"$ROOT/aift" verify >/dev/null

test -f registry/intelligence.json
test -f reports/intelligence.md

echo "OK: intelligence smoke passed"
