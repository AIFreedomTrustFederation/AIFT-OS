#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$BIN" help >/dev/null
"$BIN" doctor >/dev/null
"$ROOT/aift" help >/dev/null
"$ROOT/aift" doctor >/dev/null
"$ROOT/aift" verify >/dev/null

echo "OK: AIFT-OS v2 smoke passed"
