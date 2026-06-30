#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail
cd "${AIFT_ROOT:-$HOME/AIFT}/AIFT-OS"
git fetch origin
git restore registry/repos/discovered-repos.tsv 2>/dev/null || true
git restore scripts/phase2-central-runtime-scan.sh 2>/dev/null || true
git restore var/events/events.jsonl 2>/dev/null || true
git restore .aift/capabilities.json 2>/dev/null || true
git restore .aift/module.json 2>/dev/null || true
rm -f phase5-upgrade-existing-runtime.sh phase6-runtime-execution-engine.sh repair-phase6-runtime-pr.sh
BRANCH="phase7-doctor-git-housekeeping"
if [ "$(git branch --show-current)" != "$BRANCH" ]; then
  git checkout main
  git pull origin main
  git checkout -B "$BRANCH"
fi
