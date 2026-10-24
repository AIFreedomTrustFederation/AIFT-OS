#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
failures=0

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
    if git pull --rebase origin "$branch"; then
      echo "-- push --"
      if git push origin "$branch"; then
        echo "PASS: $name synchronized"
      else
        echo "FAIL: $name push failed" >&2
        failures=$((failures + 1))
      fi
    else
      echo "FAIL: $name pull/rebase failed; push skipped" >&2
      failures=$((failures + 1))
    fi
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
if [ "$failures" -ne 0 ]; then
  echo "FAIL: federation push sweep encountered $failures synchronization failure(s)" >&2
  exit 1
fi

echo "DONE."
