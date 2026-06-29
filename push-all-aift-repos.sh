#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"

echo "== AIFT federation push sweep =="
echo "Root: $ROOT"
echo

for repo in "$ROOT"/*; do
  [ -d "$repo/.git" ] || continue

  name="$(basename "$repo")"
  echo
  echo "== $name =="

  cd "$repo" || continue

  branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo main)"

  echo "-- status --"
  git status --short || true

  if ! git diff --quiet || ! git diff --cached --quiet || [ -n "$(git ls-files --others --exclude-standard)" ]; then
    echo "-- add/commit --"
    git add .

    if git diff --cached --quiet; then
      echo "Nothing staged after add."
    else
      git commit -m "Sync AIFT phase updates"
    fi
  else
    echo "Clean working tree."
  fi

  if git remote get-url origin >/dev/null 2>&1; then
    echo "-- pull/rebase --"
    git pull --rebase origin "$branch" || true

    echo "-- push --"
    git push origin "$branch" || true
  else
    echo "No origin remote."
  fi
done

echo
echo "== Final federation status =="
for repo in "$ROOT"/*; do
  [ -d "$repo/.git" ] || continue
  cd "$repo" || continue
  name="$(basename "$repo")"
  echo
  echo "== $name =="
  git status --short
done

echo
echo "DONE."
