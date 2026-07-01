#!/usr/bin/env bash
# no-harness: phase bootstrap script; intentionally standalone migration utility
set -euo pipefail

ROOT="${AIFT_ROOT:-$HOME/AIFT}"

find "$ROOT" -mindepth 2 -maxdepth 2 -type d -name .git | sort | while read -r gitdir; do
  repo="$(dirname "$gitdir")"
  status="$(git -C "$repo" status --short 2>/dev/null || true)"

  if [ -n "$status" ]; then
    echo
    echo "===== $repo ====="
    echo "$status"
  fi
done
