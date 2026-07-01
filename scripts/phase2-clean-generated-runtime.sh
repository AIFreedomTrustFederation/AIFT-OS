#!/usr/bin/env bash
# no-harness: phase bootstrap script; intentionally standalone cleanup utility
set -euo pipefail

ROOT="${AIFT_ROOT:-$HOME/AIFT}"

find "$ROOT" -mindepth 2 -maxdepth 2 -type d -name .git | sort | while read -r gitdir; do
  repo="$(dirname "$gitdir")"

  git -C "$repo" restore .aift/capabilities.json 2>/dev/null || true
  git -C "$repo" restore var/events/events.jsonl 2>/dev/null || true

  rm -f "$repo/.aift/module.json" 2>/dev/null || true
done

echo "Cleaned generated repo-local runtime artifacts."
