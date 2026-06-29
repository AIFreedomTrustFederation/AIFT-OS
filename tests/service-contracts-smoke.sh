#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" service-contracts init-all >/dev/null
"$ROOT/aift" service-contracts scan >/dev/null
"$ROOT/aift" service-contracts list >/dev/null
"$ROOT/aift" service-contracts repo AIFT-OS >/dev/null
"$ROOT/aift" service-contracts report >/dev/null
"$ROOT/aift" verify >/dev/null

test -f registry/service-contracts.json
test -f reports/service-contracts.md
test -f .aift/services.json
test -d .aift/services

echo "OK: service contracts smoke passed"
