#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

echo "Adding permanent aift build command"

MODULE="$(awk '/^module / {print $2; exit}' go.mod)"

mkdir -p internal/builder docs/architecture registry/builds reports

cat > docs/architecture/AIFT-BUILD-ORCHESTRATOR.md <<'DOC'
# AIFT Build Orchestrator

`aift build` is the permanent build pipeline for AIFT-OS.

It replaces phase-only scripts with an operating-system-owned workflow.

## Pipeline

1. Compile repository reality.
2. Run doctor.
3. Run verification.
4. Run Go tests.
5. Build native CLI.
6. Write build reports.
7. Stop before unsafe mutation.

## Rules

- Never fake functionality.
- Never hardcode repository names.
- Never delete source code.
- Never overwrite human work.
- Never claim blocked work succeeded.
DOC

cat > internal/builder/builder.go <<GO
package builder

import (
"encoding/json"
"fmt"
"os"
"os/exec"
"path/filepath"
"time"

"$MODULE/internal/compiler"
"$MODULE/internal/config"
)

type Step struct {
Name   string \`json:"name"\`
Status string \`json:"status"\`
}

type Report struct {
Name     string   \`json:"name"\`
Time     string   \`json:"time"\`
Root     string   \`json:"root"\`
OSHome   string   \`json:"os_home"\`
Verified bool     \`json:"verified"\`
Steps    []Step   \`json:"steps"\`
Blocked  []string \`json:"blocked"\`
}

func Run(cfg config.Config) error {
report := Report{
Name:     "AIFT Build Orchestrator",
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
report.Steps = append(report.Steps, Step{Name: name, Status: status})
}

add("repository compiler", compiler.Run(cfg))
add("doctor", run(cfg.OSHome, "aift", "doctor"))
add("verify", run(cfg.OSHome, "aift", "verify"))
add("go test", run(cfg.OSHome, "go", "test", "./..."))
add("go build", run(cfg.OSHome, "go", "build", "-o", filepath.Join(os.Getenv("HOME"), ".local", "bin", "aift"), "./cmd/aift"))

if err := writeReport(cfg, report); err != nil {
return err
}

fmt.Println("AIFT Build Orchestrator")
fmt.Println("verified:", report.Verified)

for _, step := range report.Steps {
fmt.Printf("%-25s %s\n", step.Name, step.Status)
}

if !report.Verified {
return fmt.Errorf("build completed with blocked work")
}

return nil
}

func run(dir string, name string, args ...string) error {
cmd := exec.Command(name, args...)
cmd.Dir = dir
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
return cmd.Run()
}

func writeReport(cfg config.Config, report Report) error {
buildDir := filepath.Join(cfg.OSHome, "registry", "builds")
reportDir := filepath.Join(cfg.OSHome, "reports")

if err := os.MkdirAll(buildDir, 0755); err != nil {
return err
}
if err := os.MkdirAll(reportDir, 0755); err != nil {
return err
}

b, err := json.MarshalIndent(report, "", "  ")
if err != nil {
return err
}

jsonPath := filepath.Join(buildDir, "build-report.json")
mdPath := filepath.Join(reportDir, "build-report.md")

if err := os.WriteFile(jsonPath, append(b, '\n'), 0644); err != nil {
return err
}

md := "# AIFT Build Report\n\n"
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

python - <<PY
from pathlib import Path

module = "$MODULE"
p = Path("cmd/aift/main.go")
s = p.read_text()

if f'"{module}/internal/builder"' not in s:
    pos = s.find("import (")
    line = s.find("\\n", pos) + 1
    s = s[:line] + f'\\t"{module}/internal/builder"\\n' + s[line:]

if 'case "build":' not in s:
    target = 'case "verify":'
    i = s.find(target)
    if i == -1:
        raise SystemExit('case "verify" not found')
    block = '\\tcase "build":\\n\\t\\tif err := builder.Run(cfg); err != nil {\\n\\t\\t\\tpanic(err)\\n\\t\\t}\\n\\t'
    s = s[:i] + block + s[i:]

if 'fmt.Println("  build")' not in s:
    s = s.replace(
        'fmt.Println("  verify")',
        'fmt.Println("  build")\\n\\tfmt.Println("  verify")'
    )

p.write_text(s)
PY

gofmt -w internal/builder/builder.go cmd/aift/main.go

go test ./...
go build -o "$HOME/.local/bin/aift" ./cmd/aift
hash -r

aift build || true
aift verify

git add internal/builder cmd/aift/main.go docs/architecture/AIFT-BUILD-ORCHESTRATOR.md registry/builds reports/build-report.md 2>/dev/null || true
git add registry/compiler reports/repository-compiler-report.md var/events/events.jsonl 2>/dev/null || true

git commit -m "feat: add permanent AIFT build orchestrator" || true
git push origin main

echo "DONE: aift build added"
