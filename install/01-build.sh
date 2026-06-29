#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

gofmt -w cmd internal
go test ./...
rm -f "$BIN"
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

echo "Built $BIN"
