#!/data/data/com.termux/files/usr/bin/bash
set -eu

OS="$HOME/AIFT/AIFT-OS"
cd "$OS"

echo "Repairing doctor output compatibility"

mkdir -p internal/doctor

python <<'PY'
from pathlib import Path

p = Path("internal/doctor/housekeeping.go")
if not p.exists():
    raise SystemExit("missing internal/doctor/housekeeping.go")

s = p.read_text()

s = s.replace("cfg.Stdout", "os.Stdout")
s = s.replace("cfg.Stderr", "os.Stderr")

if '"os"' not in s:
    s = s.replace('import (', 'import (\n\t"os"')

p.write_text(s)
PY

gofmt -w internal/doctor/housekeeping.go cmd/aift/main.go

go test ./...

rm -f "$HOME/.local/bin/aift"
go build -o "$HOME/.local/bin/aift" ./cmd/aift
hash -r

aift doctor
aift doctor repair
aift doctor git
aift verify

git status --short

git add internal/doctor/housekeeping.go cmd/aift/main.go docs/architecture/PHASE7-AIFT-DOCTOR-GIT-HOUSEKEEPING.md repair/ scripts/ registry/doctor/ registry/execution-plan.json registry/repos/discovered-repos.tsv 2>/dev/null || true

if git diff --cached --quiet; then
  echo "No staged changes."
else
  git commit -m "phase7: repair doctor output compatibility"
fi

git push -u origin "$(git branch --show-current)"

if gh pr view "$(git branch --show-current)" >/dev/null 2>&1; then
  gh pr view "$(git branch --show-current)" --web
else
  gh pr create \
    --base main \
    --head "$(git branch --show-current)" \
    --title "phase7: repair doctor output compatibility" \
    --body "Fixes the Doctor module so it uses standard output streams instead of nonexistent config fields. Verifies doctor repair, git housekeeping, native CLI build, and federation verification."
fi
