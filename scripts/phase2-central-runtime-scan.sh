#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="$ROOT/AIFT-OS"
OUT="$OS/registry/repos/discovered-repos.tsv"

mkdir -p "$OS/registry/repos"
mkdir -p "$OS/runtime/logs"

printf "name\tpath\tbranch\tstate\tmanifest\tremote\n" > "$OUT"

find "$ROOT" -mindepth 2 -maxdepth 2 -type d -name .git | sort | while read -r gitdir; do
  repo="$(dirname "$gitdir")"
  name="$(basename "$repo")"
  branch="$(git -C "$repo" branch --show-current 2>/dev/null || true)"
  status="$(git -C "$repo" status --short 2>/dev/null || true)"
  remote="$(git -C "$repo" remote get-url origin 2>/dev/null || true)"

  if [ -z "$branch" ]; then
    branch="unknown"
  fi

  if [ -z "$status" ]; then
    state="clean"
  else
    state="dirty"
  fi

  if [ -f "$repo/aift.repo.json" ] || [ -f "$repo/.aift/module.json" ]; then
    manifest="valid"
  else
    manifest="missing"
  fi

  printf "%s\t%s\t%s\t%s\t%s\t%s\n" "$name" "$repo" "$branch" "$state" "$manifest" "$remote" >> "$OUT"
done

echo "Wrote $OUT"
