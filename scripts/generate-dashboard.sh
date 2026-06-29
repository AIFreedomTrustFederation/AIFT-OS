#!/usr/bin/env bash
set -euo pipefail

WORKSPACE="${1:-$HOME/AIFT}"
REPORT="$WORKSPACE/reports/federation-dashboard.md"
MANIFEST="$WORKSPACE/manifests/workspace.json"

mkdir -p "$WORKSPACE/reports" "$WORKSPACE/manifests"

echo "# AIFT Federation Dashboard" > "$REPORT"
echo >> "$REPORT"
echo "Generated: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "$REPORT"
echo >> "$REPORT"

echo "{" > "$MANIFEST"
echo '  "workspace": "'"$WORKSPACE"'",' >> "$MANIFEST"
echo '  "generatedAt": "'"$(date -u +"%Y-%m-%dT%H:%M:%SZ")"'",' >> "$MANIFEST"
echo '  "repositories": [' >> "$MANIFEST"

FIRST=1

for repo in "$WORKSPACE"/*; do
  [ -d "$repo/.git" ] || continue
  cd "$repo"

  NAME="$(basename "$repo")"
  BRANCH="$(git branch --show-current || echo unknown)"
  COMMIT="$(git rev-parse --short HEAD || echo unknown)"
  DIRTY="$(git status --short | wc -l | tr -d ' ')"

  echo "## $NAME" >> "$REPORT"
  echo >> "$REPORT"
  echo "- Branch: \`$BRANCH\`" >> "$REPORT"
  echo "- Commit: \`$COMMIT\`" >> "$REPORT"

  if [ "$DIRTY" = "0" ]; then
    echo "- Status: clean" >> "$REPORT"
  else
    echo "- Status: $DIRTY changed files" >> "$REPORT"
  fi

  if [ -f "package.json" ]; then
    echo "- Runtime: Node / JavaScript / TypeScript" >> "$REPORT"
  fi

  if [ -f "Cargo.toml" ]; then
    echo "- Runtime: Rust" >> "$REPORT"
  fi

  if [ -f "go.mod" ]; then
    echo "- Runtime: Go" >> "$REPORT"
  fi

  echo >> "$REPORT"

  if [ "$FIRST" = "0" ]; then
    echo "," >> "$MANIFEST"
  fi

  FIRST=0

  cat >> "$MANIFEST" <<JSON
    {
      "name": "$NAME",
      "path": "$repo",
      "branch": "$BRANCH",
      "commit": "$COMMIT",
      "changedFiles": $DIRTY
    }
JSON
done

echo >> "$MANIFEST"
echo "  ]" >> "$MANIFEST"
echo "}" >> "$MANIFEST"

echo "Dashboard written to: $REPORT"
echo "Manifest written to: $MANIFEST"
