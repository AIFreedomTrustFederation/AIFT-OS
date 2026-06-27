#!/usr/bin/env bash
set -euo pipefail

WORKSPACE="${1:-$HOME/AIFT}"

for repo in "$WORKSPACE"/*; do
  [ -d "$repo/.git" ] || continue
  echo
  echo "======================================"
  echo "$(basename "$repo")"
  echo "======================================"
  cd "$repo"
  echo "Branch: $(git branch --show-current)"
  echo "Commit: $(git rev-parse --short HEAD)"
  git status --short
done
