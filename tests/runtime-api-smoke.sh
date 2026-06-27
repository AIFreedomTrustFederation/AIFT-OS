#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" start >/dev/null
"$ROOT/aift" tick >/dev/null
"$ROOT/aift" services >/dev/null
"$ROOT/aift" verify >/dev/null

echo "OK: runtime/API smoke passed"
