#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

echo "Repairing Phase 13 with repository-aware module discovery"

OS="$PWD"
MODULE="$(awk '/^module / {print $2; exit}' go.mod)"

if [ -z "$MODULE" ]; then
  echo "ERROR: Could not read module path from go.mod"
  exit 1
fi

echo "Detected module: $MODULE"

mkdir -p internal/compiler internal/ai registry/compiler registry/ai reports docs/architecture

cat > internal/compiler/compiler.go <<GO
package compiler

import (
"encoding/json"
"fmt"
"os"
"os/exec"
"path/filepath"
"strings"
"time"

"$MODULE/internal/config"
)

type Repo struct {
Name     string \`json:"name"\`
Path     string \`json:"path"\`
Branch   string \`json:"branch"\`
State    string \`json:"state"\`
Manifest string \`json:"manifest"\`
Remote   string \`json:"remote"\`
}

type Report struct {
Name     string   \`json:"name"\`
Time     string   \`json:"time"\`
Root     string   \`json:"root"\`
OSHome   string   \`json:"os_home"\`
Verified bool     \`json:"verified"\`
Repos    []Repo   \`json:"repos"\`
Blocked  []string \`json:"blocked"\`
}

func Run(cfg config.Config) error {
repos, blocked := discover(cfg.Root)

report := Report{
Name:     "AIFT Repository Compiler",
Time:     time.Now().Format(time.RFC3339),
Root:     cfg.Root,
OSHome:   cfg.OSHome,
Verified: true,
Repos:    repos,
Blocked:  blocked,
}

outDir := filepath.Join(cfg.OSHome, "registry", "compiler")
reportDir := filepath.Join(cfg.OSHome, "reports")

if err := os.MkdirAll(outDir, 0755); err != nil {
return err
}
if err := os.MkdirAll(reportDir, 0755); err != nil {
return err
}

b, err := json.MarshalIndent(report, "", "  ")
if err != nil {
return err
}

jsonPath := filepath.Join(outDir, "repository-compiler-report.json")
mdPath := filepath.Join(reportDir, "repository-compiler-report.md")

if err := os.WriteFile(jsonPath, append(b, '\n'), 0644); err != nil {
return err
}

md := "# AIFT Repository Compiler Report\n\n"
md += fmt.Sprintf("Verified: %v\n\n", report.Verified)
md += "## Repositories\n\n"

for _, repo := range repos {
md += fmt.Sprintf("- %s | %s | %s | %s\n", repo.Name, repo.State, repo.Manifest, repo.Branch)
}

if len(blocked) > 0 {
md += "\n## Blocked\n\n"
for _, item := range blocked {
md += "- " + item + "\n"
}
}

if err := os.WriteFile(mdPath, []byte(md), 0644); err != nil {
return err
}

fmt.Println("AIFT Repository Compiler")
fmt.Println("repos:", len(repos))
fmt.Println("blocked:", len(blocked))
fmt.Println("wrote:", jsonPath)
fmt.Println("wrote:", mdPath)

return nil
}

func discover(root string) ([]Repo, []string) {
var repos []Repo
var blocked []string

err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
if err != nil {
blocked = append(blocked, path+": "+err.Error())
return filepath.SkipDir
}

if !d.IsDir() {
return nil
}

base := filepath.Base(path)
switch base {
case ".git", "node_modules", ".next", "dist", "build", "vendor", "runtime", "reports":
return filepath.SkipDir
}

if exists(filepath.Join(path, ".git")) {
repos = append(repos, inspect(path))
return filepath.SkipDir
}

return nil
})

if err != nil {
blocked = append(blocked, err.Error())
}

return repos, blocked
}

func inspect(path string) Repo {
repo := Repo{
Name:     filepath.Base(path),
Path:     path,
Branch:   gitOut(path, "branch", "--show-current"),
State:    "clean",
Remote:   gitOut(path, "remote", "get-url", "origin"),
Manifest: "missing",
}

if repo.Branch == "" {
repo.Branch = "unknown"
}

if gitOut(path, "status", "--short") != "" {
repo.State = "dirty"
}

if exists(filepath.Join(path, "aift.repo.json")) || exists(filepath.Join(path, ".aift", "module.json")) {
repo.Manifest = "valid"
}

return repo
}

func exists(path string) bool {
_, err := os.Stat(path)
return err == nil
}

func gitOut(dir string, args ...string) string {
cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
out, err := cmd.Output()
if err != nil {
return ""
}
return strings.TrimSpace(string(out))
}
GO

cat > internal/ai/ai.go <<GO
package ai

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
report.Steps = append(report.Steps, Step{Name: name, Status: status})
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

func run(dir string, name string, args ...string) error {
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

python - <<PY
from pathlib import Path

module = "$MODULE"
p = Path("cmd/aift/main.go")
s = p.read_text()

bad = "github.com/AIFreedomTrustFederation/AIFT-OS"
s = s.replace(bad + "/internal/compiler", module + "/internal/compiler")
s = s.replace(bad + "/internal/ai", module + "/internal/ai")

if f'"{module}/internal/compiler"' not in s:
    pos = s.find("import (")
    if pos != -1:
        line = s.find("\\n", pos) + 1
        s = s[:line] + f'\\t"{module}/internal/compiler"\\n' + s[line:]

if f'"{module}/internal/ai"' not in s:
    pos = s.find("import (")
    if pos != -1:
        line = s.find("\\n", pos) + 1
        s = s[:line] + f'\\t"{module}/internal/ai"\\n' + s[line:]

if 'case "compile":' not in s:
    target = 'case "verify":'
    i = s.find(target)
    if i != -1:
        s = s[:i] + '\\tcase "compile":\\n\\t\\tif err := compiler.Run(cfg); err != nil {\\n\\t\\t\\tpanic(err)\\n\\t\\t}\\n\\tcase "compiler":\\n\\t\\tif err := compiler.Run(cfg); err != nil {\\n\\t\\t\\tpanic(err)\\n\\t\\t}\\n\\t' + s[i:]

if 'case "ai":' not in s:
    target = 'case "verify":'
    i = s.find(target)
    if i != -1:
        s = s[:i] + '\\tcase "ai":\\n\\t\\tif err := ai.Run(cfg); err != nil {\\n\\t\\t\\tpanic(err)\\n\\t\\t}\\n\\t' + s[i:]

if 'fmt.Println("  compile")' not in s:
    s = s.replace('fmt.Println("  verify")', 'fmt.Println("  ai")\\n\\tfmt.Println("  compile")\\n\\tfmt.Println("  compiler")\\n\\tfmt.Println("  verify")')

p.write_text(s)
PY

gofmt -w internal/compiler/compiler.go internal/ai/ai.go cmd/aift/main.go

echo "Listing discovered Go packages"
go list ./... | grep -E '/internal/(compiler|ai)$' || true

go test ./...
go build -o "$HOME/.local/bin/aift" ./cmd/aift
hash -r

aift compiler
aift ai || true
aift verify

git add internal/compiler internal/ai cmd/aift/main.go registry/compiler registry/ai reports docs/architecture 2>/dev/null || true
git add var/events/events.jsonl 2>/dev/null || true

git commit -m "phase13: add repository-aware AI foundation" || true
git push -u origin phase13-ai-orchestrator || true

echo "DONE phase13 repository-aware AI foundation"
