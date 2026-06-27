#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" repo list >/dev/null
"$ROOT/aift" repo inspect AIFT-OS >/dev/null
"$ROOT/aift" workflow list >/dev/null
"$ROOT/aift" federation scan >/dev/null
"$ROOT/aift" federation graph >/dev/null
"$ROOT/aift" federation verify >/dev/null

test -f registry/federation-snapshot.json
test -f registry/workflows.json

echo "OK: federation integration smoke passed"
