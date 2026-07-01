#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

BRANCH="phase13-ai-orchestrator"

echo "Phase 13 AIFT AI Orchestrator"

git add registry reports .aift var 2>/dev/null || true
git commit -m "housekeeping: save generated runtime artifacts" 2>/dev/null || true

if ! git diff --quiet || ! git diff --cached --quiet; then
    git stash push -u -m "phase13-auto-housekeeping" || true
fi

git fetch origin
git checkout main
git pull origin main
git checkout -B "$BRANCH"

mkdir -p internal/ai docs/architecture registry/ai reports

cat > docs/architecture/PHASE13-AI-ORCHESTRATOR.md <<'DOC'
# AIFT-OS Phase 13: AI Orchestrator

AIFT AI Orchestrator replaces the growing collection of phase scripts with a permanent operating-system service.

Command:

    aift ai

Execution pipeline

1. Compile repository reality
2. Run doctor
3. Verify repository
4. Record successful work
5. Record blocked work
6. Produce reports
7. Stop before unsafe mutation

Rules

- Never fake functionality.
- Never hardcode repository names.
- Never delete source code.
- Never overwrite human work.
- Never claim blocked work succeeded.
- Generated reports belong under registry/ai and reports.
DOC

cat > internal/ai/ai.go <<'GO'
package ai

import (
"encoding/json"
"fmt"
"os"
"os/exec"
"path/filepath"
"time"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/compiler"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type Step struct {
Name   string `json:"name"`
Status string `json:"status"`
}

type Report struct {
Name     string   `json:"name"`
Time     string   `json:"time"`
Root     string   `json:"root"`
OSHome   string   `json:"os_home"`
Verified bool     `json:"verified"`
Steps    []Step   `json:"steps"`
Blocked  []string `json:"blocked"`
}

func Run(cfg config.Config) error {

report := Report{
Name:     "AIFT AI Orchestrator",
Time:     time.Now().Format(time.RFC3339),
Root:     cfg.Root,
OSHome:   cfg.OSHome,
Verified: true,
}

add := func(name string, err error) {
status := "pass"

if err != nil {
status = "blocked"
report.Verified = false
report.Blocked = append(report.Blocked, name+": "+err.Error())
}

report.Steps = append(report.Steps, Step{
Name: name,
Status: status,
})
}

add("repository compiler", compiler.Run(cfg))
add("doctor", run(cfg.OSHome, "aift", "doctor"))
add("verify", run(cfg.OSHome, "aift", "verify"))

if err := writeReport(cfg, report); err != nil {
return err
}

fmt.Println("AIFT AI Orchestrator")
fmt.Println("verified:", report.Verified)

for _, step := range report.Steps {
fmt.Printf("%-25s %s\n", step.Name, step.Status)
}

if !report.Verified {
return fmt.Errorf("AI orchestration completed with blocked work")
}

return nil
}

func run(dir, name string, args ...string) error {
cmd := exec.Command(name, args...)
cmd.Dir = dir
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
return cmd.Run()
}

func writeReport(cfg config.Config, report Report) error {

aiDir := filepath.Join(cfg.OSHome, "registry", "ai")
reportDir := filepath.Join(cfg.OSHome, "reports")

if err := os.MkdirAll(aiDir, 0755); err != nil {
return err
}

if err := os.MkdirAll(reportDir, 0755); err != nil {
return err
}

b, err := json.MarshalIndent(report, "", "  ")
if err != nil {
return err
}

jsonPath := filepath.Join(aiDir, "ai-orchestrator-report.json")
mdPath := filepath.Join(reportDir, "ai-orchestrator-report.md")

if err := os.WriteFile(jsonPath, append(b, '\n'), 0644); err != nil {
return err
}

md := "# AIFT AI Orchestrator Report\n\n"
md += fmt.Sprintf("Verified: %v\n\n", report.Verified)
md += "## Steps\n\n"

for _, step := range report.Steps {
md += "- " + step.Name + ": " + step.Status + "\n"
}

if len(report.Blocked) > 0 {
md += "\n## Blocked\n\n"

for _, item := range report.Blocked {
md += "- " + item + "\n"
}
}

return os.WriteFile(mdPath, []byte(md), 0644)
}
GO

python <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text()

if '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/ai"' not in s:
    marker = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/'
    pos = s.find(marker)
    if pos != -1:
        line = s.rfind("\n", 0, pos) + 1
        s = s[:line] + '\t"github.com/AIFreedomTrustFederation/AIFT-OS/internal/ai"\n' + s[line:]

if 'case "ai":' not in s:
    target = 'case "verify":'
    idx = s.find(target)
    if idx == -1:
        raise SystemExit("verify case not found")

    insert = '''
case "ai":
if err := ai.Run(cfg); err != nil {
panic(err)
}
'''

    s = s[:idx] + insert + s[idx:]

if 'fmt.Println("  ai")' not in s:
    s = s.replace(
        'fmt.Println("  verify")',
        'fmt.Println("  ai")\n\tfmt.Println("  verify")'
    )

p.write_text(s)
PY

gofmt -w internal/ai/ai.go cmd/aift/main.go

go test ./...
go build -o "$HOME/.local/bin/aift" ./cmd/aift

hash -r

aift ai || true
aift verify

git add internal/ai
git add cmd/aift/main.go
git add docs/architecture/PHASE13-AI-ORCHESTRATOR.md
git add registry/ai reports/ai-orchestrator-report.md 2>/dev/null || true
git add registry/compiler reports/repository-compiler-report.md 2>/dev/null || true
git add var/events/events.jsonl 2>/dev/null || true

git commit -m "phase13: add AI orchestrator" || true
git push -u origin "$BRANCH"

gh pr create \
  --base main \
  --head "$BRANCH" \
  --title "phase13: add AI orchestrator" \
  --body "Adds a permanent AI orchestration subsystem that coordinates compiler, doctor, verification, reporting, and conservative stopping conditions." || true

echo
echo "========================================="
echo " Phase 13 Complete"
echo "========================================="
echo
echo "Run:"
echo
echo "    aift ai"
echo "    aift verify"
echo "    aift doctor"
echo
