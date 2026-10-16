#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
OS="${AIFT_OS_HOME:-${AIFT_ROOT:+$AIFT_ROOT/AIFT-OS}}"
OS="${OS:-$SCRIPT_DIR}"
BIN="$OS/bin/aift"

cd "$OS" || exit 1

if [ ! -x "$BIN" ]; then
  mkdir -p "$(dirname "$BIN")"
  go build -o "$BIN" ./cmd/aift
  chmod +x "$BIN"
fi

exec "$BIN" "$@"
