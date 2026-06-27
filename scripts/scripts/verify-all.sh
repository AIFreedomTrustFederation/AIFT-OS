#!/usr/bin/env bash
set -euo pipefail

WORKSPACE="${1:-$HOME/AIFT}"

for repo in "$WORKSPACE"/*; do
  [ -d "$repo/.git" ] || continue
  cd "$repo"

  echo
  echo "======================================"
  echo "Verifying $(basename "$repo")"
  echo "======================================"

  git status --short

  if [ -f "package.json" ] && command -v npm >/dev/null 2>&1; then
    npm run lint --if-present || true
    npm run typecheck --if-present || true
    npm run test --if-present || true
    npm run build --if-present || true
  fi

  if [ -f "Cargo.toml" ] && command -v cargo >/dev/null 2>&1; then
    cargo check || true
  fi

  if [ -f "go.mod" ] && command -v go >/dev/null 2>&1; then
    go test ./... || true
  fi
done
