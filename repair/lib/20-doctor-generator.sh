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
