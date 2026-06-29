#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" capabilities scan >/dev/null
"$ROOT/aift" capabilities report >/dev/null
"$ROOT/aift" capabilities repo AIFT-OS >/dev/null
"$ROOT/aift" verify >/dev/null

test -f registry/capabilities.json
test -f reports/capabilities.md
test -f .aift/capabilities.json

echo "OK: capabilities smoke passed"
