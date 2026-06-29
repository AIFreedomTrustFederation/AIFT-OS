#!/usr/bin/env bash
set -euo pipefail

WORKSPACE="${1:-$HOME/AIFT}"

for repo in "$WORKSPACE"/*; do
  [ -d "$repo/.git" ] || continue
  echo
  echo "======================================"
  echo "Pulling $(basename "$repo")"
  echo "======================================"
  cd "$repo"
  git pull --ff-only || echo "Manual merge needed in $(basename "$repo")"
done
