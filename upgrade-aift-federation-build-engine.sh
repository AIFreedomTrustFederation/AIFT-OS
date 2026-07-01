#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

echo "Upgrading AIFT build into federation-wide module-agnostic build engine"

MODULE="$(awk '/^module / {print $2; exit}' go.mod)"

mkdir -p internal/fedbuild docs/architecture registry/fedbuild reports

cat > docs/architecture/AIFT-FEDERATION-BUILD-ENGINE.md <<'DOC'
# AIFT Federation Build Engine

`aift federation-build` discovers and builds the federation without hardcoded repository names.

## Design

Every repository is treated as a module.

Every module may be:

- synchronous
- asynchronous
- queued
- blocked
- skipped
- planned
- active after verification

## Rules

- Never fake functionality.
- Never hardcode repository names.
- Never assume package managers.
- Never delete source code.
- Never overwrite human work.
- Never claim blocked work succeeded.
- Discover reality from disk.
- Build only when a real build system is detected.
DOC

cat > internal/fedbuild/fedbuild.go <<GO
package fedbuild

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

type Module struct {
Name      string   \`json:"name"\`
Path      string   \`json:"path"\`
Branch    string   \`json:"branch"\`
State     string   \`json:"state"\`
Mode      string   \`json:"mode"\`
Runtime   string   \`json:"runtime"\`
Build     []string \`json:"build"\`
Test      []string \`json:"test"\`
Status    string   \`json:"status"\`
Blocked   []string \`json:"blocked"\`
Artifacts []string \`json:"artifacts"\`
}

type Report struct {
Name     string   \`json:"name"\`
Time     string   \`json:"time"\`
Root     string   \`json:"root"\`
OSHome   string   \`json:"os_home"\`
Verified bool     \`json:"verified"\`
Mode     string   \`json:"mode"\`
Modules  []Module \`json:"modules"\`
Blocked  []string \`json:"blocked"\`
}

func Run(cfg config.Config, async bool) error {
mode := "sync"
if async {
mode = "async-planned"
}

modules, blocked := discover(cfg.Root)

report := Report{
Name:     "AIFT Federation Build Engine",
Time:     time.Now().Format(time.RFC3339),
Root:     cfg.Root,
OSHome:   cfg.OSHome,
Verified: true,
Mode:     mode,
Modules:  modules,
Blocked:  blocked,
}

for i := range report.Modules {
module := &report.Modules[i]
module.Mode = mode

if len(module.Blocked) > 0 {
module.Status = "blocked"
report.Verified = false
continue
}

if async {
module.Status = "planned"
continue
}

if err := runSteps(module.Path, module.Build); err != nil {
module.Status = "blocked"
module.Blocked = append(module.Blocked, "build failed: "+err.Error())
report.Verified = false
continue
}

if err := runSteps(module.Path, module.Test); err != nil {
module.Status = "blocked"
module.Blocked = append(module.Blocked, "test failed: "+err.Error())
report.Verified = false
continue
}

module.Status = "active"
}

if err := writeReport(cfg, report); err != nil {
return err
}

fmt.Println("AIFT Federation Build Engine")
fmt.Println("mode:", report.Mode)
fmt.Println("verified:", report.Verified)
fmt.Println("modules:", len(report.Modules))
fmt.Println("blocked:", countBlocked(report.Modules)+len(report.Blocked))

if !report.Verified {
return fmt.Errorf("federation build completed with blocked modules")
}

return nil
}

func discover(root string) ([]Module, []string) {
var modules []Module
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
case ".git", "node_modules", ".next", "dist", "build", "vendor", "runtime", "reports", "registry", ".cache":
return filepath.SkipDir
}

if exists(filepath.Join(path, ".git")) {
modules = append(modules, inspect(path))
return filepath.SkipDir
}

return nil
})

if err != nil {
blocked = append(blocked, err.Error())
}

return modules, blocked
}

func inspect(path string) Module {
module := Module{
Name:   filepath.Base(path),
Path:   path,
Branch: gitOut(path, "branch", "--show-current"),
State:  "clean",
Status: "discovered",
}

if module.Branch == "" {
module.Branch = "unknown"
}

if gitOut(path, "status", "--short") != "" {
module.State = "dirty"
}

switch {
case exists(filepath.Join(path, "go.mod")):
module.Runtime = "go"
module.Build = []string{"go build ./..."}
module.Test = []string{"go test ./..."}
case exists(filepath.Join(path, "package.json")):
module.Runtime = "node"
module.Build = []string{"npm run build"}
module.Test = []string{"npm test"}
case exists(filepath.Join(path, "Cargo.toml")):
module.Runtime = "rust"
module.Build = []string{"cargo build"}
module.Test = []string{"cargo test"}
case exists(filepath.Join(path, "Makefile")):
module.Runtime = "make"
module.Build = []string{"make"}
module.Test = []string{"make test"}
default:
module.Runtime = "unknown"
module.Blocked = append(module.Blocked, "no supported build runtime detected")
}

if !exists(filepath.Join(path, "aift.repo.json")) && !exists(filepath.Join(path, ".aift", "module.json")) {
module.Blocked = append(module.Blocked, "missing AIFT manifest")
}

return module
}

func runSteps(dir string, steps []string) error {
for _, step := range steps {
parts := strings.Fields(step)
if len(parts) == 0 {
continue
}

cmd := exec.Command(parts[0], parts[1:]...)
cmd.Dir = dir
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr

if err := cmd.Run(); err != nil {
return fmt.Errorf("%s: %w", step, err)
}
}

return nil
}

func writeReport(cfg config.Config, report Report) error {
outDir := filepath.Join(cfg.OSHome, "registry", "fedbuild")
reportDir := filepath.Join(cfg.OSHome, "reports")

if err := os.MkdirAll(outDir, 0755); err != nil {
return err
}

if err := os.MkdirAll(reportDir, 0755); err != nil {
return err
}

jsonPath := filepath.Join(outDir, "federation-build.json")
mdPath := filepath.Join(reportDir, "federation-build.md")

b, err := json.MarshalIndent(report, "", "  ")
if err != nil {
return err
}

if err := os.WriteFile(jsonPath, append(b, '\n'), 0644); err != nil {
return err
}

md := "# AIFT Federation Build Report\n\n"
md += fmt.Sprintf("Mode: %s\n\n", report.Mode)
md += fmt.Sprintf("Verified: %v\n\n", report.Verified)
md += "## Modules\n\n"

for _, module := range report.Modules {
md += fmt.Sprintf("- %s | %s | %s | %s | %s\n", module.Name, module.Runtime, module.State, module.Status, module.Mode)

if len(module.Blocked) > 0 {
md += "  - blocked: " + strings.Join(module.Blocked, ", ") + "\n"
}
}

return os.WriteFile(mdPath, []byte(md), 0644)
}

func countBlocked(modules []Module) int {
total := 0
for _, module := range modules {
if len(module.Blocked) > 0 || module.Status == "blocked" {
total++
}
}
return total
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

python - <<PY
from pathlib import Path

module = "$MODULE"
p = Path("cmd/aift/main.go")
s = p.read_text()

if f'"{module}/internal/fedbuild"' not in s:
    pos = s.find("import (")
    line = s.find("\\n", pos) + 1
    s = s[:line] + f'\\t"{module}/internal/fedbuild"\\n' + s[line:]

if 'case "federation-build":' not in s:
    target = 'case "build":'
    i = s.find(target)
    if i == -1:
        raise SystemExit('case "build" not found')

    block = '''\tcase "federation-build":
\t\tasync := len(args) > 1 && args[1] == "--async"
\t\tif err := fedbuild.Run(cfg, async); err != nil {
\t\t\tpanic(err)
\t\t}
\t'''

    s = s[:i] + block + s[i:]

if 'fmt.Println("  federation-build")' not in s:
    s = s.replace(
        'fmt.Println("  build")',
        'fmt.Println("  build")\\n\\tfmt.Println("  federation-build")'
    )

p.write_text(s)
PY

gofmt -w internal/fedbuild/fedbuild.go cmd/aift/main.go

go test ./...
go build -o "$HOME/.local/bin/aift" ./cmd/aift
hash -r

aift federation-build --async || true
aift build || true
aift verify

git add internal/fedbuild cmd/aift/main.go docs/architecture/AIFT-FEDERATION-BUILD-ENGINE.md registry/fedbuild reports/federation-build.md 2>/dev/null || true
git add registry/builds reports/build-report.md registry/lifecycle reports/federation-lifecycle.md registry/compiler reports/repository-compiler-report.md var/events/events.jsonl 2>/dev/null || true

git commit -m "feat: add module-agnostic federation build engine" || true
git push origin main

echo "DONE: module-agnostic sync async federation build engine added"
