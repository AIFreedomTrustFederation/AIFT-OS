#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail
cd "${AIFT_ROOT:-$HOME/AIFT}/AIFT-OS"
git add cmd/aift/main.go internal/doctor/housekeeping.go docs/architecture/PHASE7-AIFT-DOCTOR-GIT-HOUSEKEEPING.md repair/lib repair-phase7.sh
if git diff --cached --quiet; then
  echo "No staged changes."
else
  git commit -m "phase7: add doctor git housekeeping"
fi
git push -u origin HEAD
if command -v gh >/dev/null 2>&1; then
  gh pr create --base main --head "$(git branch --show-current)" --title "phase7: add doctor git housekeeping" --body "Adds AIFT doctor repair, git housekeeping, verification, and generated-state cleanup."
fi
