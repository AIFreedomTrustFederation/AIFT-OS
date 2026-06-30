#!/data/data/com.termux/files/usr/bin/bash
set -u

OUT="AIFT-OS-LOCAL-INSPECTION-$(date +%Y%m%d-%H%M%S).txt"

{
echo "=== LOCATION ==="
pwd

echo
echo "=== GIT ==="
git remote -v
git branch -vv
git status --short
git rev-parse HEAD

echo
echo "=== GO VERSION ==="
go version

echo
echo "=== ROOT FILES ==="
ls -la

echo
echo "=== CMD/AIFT FILES ==="
ls -la cmd/aift
for f in cmd/aift/*.go; do
  echo
  echo "===== $f ====="
  sed -n '1,260p' "$f"
done

echo
echo "=== INTERNAL PACKAGES ==="
find internal -maxdepth 2 -type f -name '*.go' | sort

echo
echo "=== BUILD OUTPUT ==="
go build ./cmd/aift 2>&1 || true

echo
echo "=== PACKAGE REFERENCES ==="
grep -R "func .*List\|func .*Run\|func .*Scan\|func .*Generate\|func .*EnsureAll\|package " internal cmd/aift -n | head -n 500

} > "$OUT"

echo "Created inspection report:"
echo "$OUT"
