#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/common.sh"

out="$AIFT_OS_HOME/reports/dependency-graph.md"
mkdir -p "$AIFT_OS_HOME/reports"

{
  echo "# AIFT Dependency Graph"
  echo
  echo "| Repository | Dependencies |"
  echo "|---|---|"

  for repo in $(aift_find_repos); do
    name="$(aift_repo_name "$repo")"
    file="$(aift_manifest_path "$repo")"
    deps="[]"
    if [ -f "$file" ]; then
      deps="$(grep '"dependencies"' "$file" 2>/dev/null || echo '"dependencies": []')"
    fi
    deps="$(printf '%s' "$deps" | sed 's/^[[:space:]]*//; s/[",]//g')"
    echo "| \`$name\` | \`$deps\` |"
  done
} > "$out"

echo "Wrote $out"
