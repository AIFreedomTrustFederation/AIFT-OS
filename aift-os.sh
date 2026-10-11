#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$SCRIPT_DIR}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

if [ ! -x "$BIN" ]; then
  mkdir -p "$(dirname "$BIN")"
  go build -o "$BIN" ./cmd/aift
  chmod +x "$BIN"
fi

exec "$BIN" "$@"
