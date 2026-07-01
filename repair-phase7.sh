#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="$ROOT/AIFT-OS"
BRANCH="phase7-doctor-git-housekeeping"

cd "$OS"

mkdir -p repair/lib docs/architecture internal/doctor registry/doctor runtime/logs scripts

cat > repair/lib/00-bootstrap.sh <<'LIB'
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
LIB

cat > repair/lib/10-git-housekeeping.sh <<'LIB'
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
LIB

cat > repair/lib/20-doctor-generator.sh <<'LIB'
#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail
cd "${AIFT_ROOT:-$HOME/AIFT}/AIFT-OS"

cat > docs/architecture/PHASE7-AIFT-DOCTOR-GIT-HOUSEKEEPING.md <<'DOC'
# AIFT-OS Phase 7: Doctor and Git Housekeeping

AIFT Doctor inspects the real local federation workspace.

It repairs safe generated state, checks git status, verifies the native Go CLI, and reports what still needs human review.

AIFT-OS remains the central runtime. Other repositories remain source packages.
DOC

cat > internal/doctor/housekeeping.go <<'GO'
package doctor

import (
"fmt"
"os/exec"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func Git(cfg config.Config) error {
root := cfg.Root
cmd := exec.Command("sh", "-c", `find "`+root+`" -mindepth 2 -maxdepth 2 -type d -name .git | sort | while read gitdir; do repo=$(dirname "$gitdir"); echo "== $repo =="; git -C "$repo" status --short; done`)
cmd.Stdout = cfg.Stdout
cmd.Stderr = cfg.Stderr
return cmd.Run()
}

func Repair(cfg config.Config) error {
fmt.Fprintln(cfg.Stdout, "Repairing generated runtime state")
cmd := exec.Command("sh", "-c", `find "`+cfg.Root+`" -mindepth 2 -maxdepth 2 -type d -name .git | sort | while read gitdir; do repo=$(dirname "$gitdir"); git -C "$repo" restore .aift/capabilities.json .aift/providers.json .aift/workflows.json .aift/repos.json var/events/events.jsonl 2>/dev/null || true; rm -f "$repo/.aift/module.json"; done`)
cmd.Stdout = cfg.Stdout
cmd.Stderr = cfg.Stderr
return cmd.Run()
}

func Full(cfg config.Config) error {
if err := Repair(cfg); err != nil {
return err
}
return Run(cfg)
}
GO
LIB

cat > repair/lib/30-main-patcher.sh <<'LIB'
#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail
cd "${AIFT_ROOT:-$HOME/AIFT}/AIFT-OS"

python <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text()

imp = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/doctor"'
if imp not in s:
    marker = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"'
    s = s.replace(marker, marker + "\n\t" + imp)

if 'case "doctor":' in s and 'doctor.Git' not in s:
    old = '''case "doctor":
\t\terr = doctor.Run(cfg)'''
    new = '''case "doctor":
\t\tif len(args) > 0 && args[0] == "repair" {
\t\t\terr = doctor.Repair(cfg)
\t\t} else if len(args) > 0 && args[0] == "git" {
\t\t\terr = doctor.Git(cfg)
\t\t} else if len(args) > 0 && args[0] == "full" {
\t\t\terr = doctor.Full(cfg)
\t\t} else {
\t\t\terr = doctor.Run(cfg)
\t\t}'''
    s = s.replace(old, new)

if 'doctor [repair|git|full]' not in s:
    s = s.replace('fmt.Println("  doctor")', 'fmt.Println("  doctor [repair|git|full]")')

p.write_text(s)
PY
LIB

cat > repair/lib/40-verify.sh <<'LIB'
#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail
cd "${AIFT_ROOT:-$HOME/AIFT}/AIFT-OS"
gofmt -w cmd/aift/main.go internal/doctor/*.go
go test ./...
go build -o "$HOME/.local/bin/aift" ./cmd/aift
hash -r
aift doctor
aift doctor repair
aift doctor git
aift verify
LIB

cat > repair/lib/50-commit-pr.sh <<'LIB'
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
LIB

chmod +x repair/lib/*.sh

for f in repair/lib/*.sh; do
  echo "Running $f"
  bash "$f"
done
