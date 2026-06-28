#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" manual init-all >/dev/null
"$ROOT/aift" manual scan >/dev/null
"$ROOT/aift" manual report >/dev/null
"$ROOT/aift" manual repo AIFT-OS >/dev/null
"$ROOT/aift" verify >/dev/null

test -f registry/manuals.json
test -f reports/manuals.md
test -f .aift/manual.json
test -d docs/manual/source/man0
test -d docs/manual/source/man7

echo "OK: manual contracts smoke passed"
