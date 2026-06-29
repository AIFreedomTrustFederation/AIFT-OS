#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"

cd "$OS" || exit 1

echo "AIFT-OS inspect"
echo "ROOT=$ROOT"
echo "OS=$OS"
echo

echo "== launchers =="
ls -la "$ROOT/aift" aift-os.sh bin/aiftd 2>/dev/null || true
echo

echo "== launcher content =="
sed -n '1,80p' "$ROOT/aift" 2>/dev/null || true
echo
sed -n '1,120p' aift-os.sh 2>/dev/null || true
echo

echo "== go =="
go version
go test ./...
echo

echo "== commands =="
"$ROOT/aift" help
"$ROOT/aift" doctor
echo

echo "== git =="
git status --short
git branch -vv || true
