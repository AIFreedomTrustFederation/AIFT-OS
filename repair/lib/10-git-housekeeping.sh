#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail
ROOT="${AIFT_ROOT:-$HOME/AIFT}"
cd "$ROOT"
find "$ROOT" -mindepth 2 -maxdepth 2 -type d -name .git | sort | while read -r gitdir; do
  repo="$(dirname "$gitdir")"
  cd "$repo" || continue
  cat >> .gitignore <<'GITIGNORE'
.aift/capabilities.json
.aift/providers.json
.aift/workflows.json
.aift/repos.json
.aift/federation-snapshot.json
var/events/
reports/
registry/*.json
registry/*.dot
registry/*.rdf
registry/*.graphml
registry/*.cypher
GITIGNORE
  sort -u .gitignore -o .gitignore
  git restore .aift/capabilities.json 2>/dev/null || true
  git restore .aift/providers.json 2>/dev/null || true
  git restore .aift/workflows.json 2>/dev/null || true
  git restore .aift/repos.json 2>/dev/null || true
  git restore var/events/events.jsonl 2>/dev/null || true
  rm -f .aift/module.json 2>/dev/null || true
done
