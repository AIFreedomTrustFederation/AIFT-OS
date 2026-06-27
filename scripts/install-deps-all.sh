#!/usr/bin/env bash
set -euo pipefail

WORKSPACE="${1:-$HOME/AIFT}"

for repo in "$WORKSPACE"/*; do
  [ -d "$repo/.git" ] || continue
  cd "$repo"

  echo
  echo "======================================"
  echo "Installing deps: $(basename "$repo")"
  echo "======================================"

  if [ -f "pnpm-lock.yaml" ] && command -v pnpm >/dev/null 2>&1; then
    pnpm install || true
  elif [ -f "bun.lockb" ] && command -v bun >/dev/null 2>&1; then
    bun install || true
  elif [ -f "package-lock.json" ] && command -v npm >/dev/null 2>&1; then
    npm ci || npm install || true
  elif [ -f "package.json" ] && command -v npm >/dev/null 2>&1; then
    npm install || true
  fi

  if [ -f "Cargo.toml" ] && command -v cargo >/dev/null 2>&1; then
    cargo fetch || true
  fi

  if [ -f "go.mod" ] && command -v go >/dev/null 2>&1; then
    go mod download || true
  fi
done
