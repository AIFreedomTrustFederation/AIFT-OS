#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"

cd "$OS" || exit 1

go test ./...
sh install/01-build.sh >/dev/null

"$ROOT/aift" help >/dev/null
"$ROOT/aift" doctor >/dev/null
"$ROOT/aift" status >/dev/null
"$ROOT/aift" manifest >/dev/null
"$ROOT/aift" registry >/dev/null
"$ROOT/aift" dashboard >/dev/null
"$ROOT/aift" deps >/dev/null
"$ROOT/aift" plugins >/dev/null
"$ROOT/aift" sync --safe >/dev/null
"$ROOT/aift" verify >/dev/null

echo "OK: AIFT-OS v2 launcher smoke tests passed"
