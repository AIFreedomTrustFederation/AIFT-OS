#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

echo "Adding capability-aware federation build engine"

MODULE="$(awk '/^module / {print $2; exit}' go.mod)"

mkdir -p internal/capability internal/fedbuild docs/architecture registry/capabilities registry/fedbuild reports

cat > docs/architecture/AIFT-CAPABILITY-AWARE-BUILD.md <<'DOC'
# AIFT Capability-Aware Build

AIFT must never fail a federation build simply because the current machine lacks a runtime.

Instead, it discovers local capabilities first, then classifies modules honestly:

- active: runnable here
- planned: valid but waiting for runtime/tooling
- blocked: invalid or unsafe
- unsupported: no provider can handle it

This keeps AIFT module-agnostic, provider-agnostic, runtime-agnostic, and sync/async capable.
DOC

cat > internal/capability/capability.go <<GO
package capability

import (
"encoding/json"
"fmt"
"os"
"os/exec"
"path/filepath"
"time"

"$MODULE/internal/config"
)

type Capability struct {
Name      string \`json:"name"\`
Command   string \`json:"command"\`
Installed bool   \`json:"installed"\`
Path      string \`json:"path"\`
}

type Report struct {
Name         string       \`json:"name"\`
Time         string       \`json:"time"\`
Root         string       \`json:"root"\`
OSHome       string       \`json:"os_home"\`
Verified     bool         \`json:"verified"\`
Capabilities []Capability \`json:"capabilities"\`
}

func Discover(cfg config.Config) Report {
names := []string{
"go",
"git",
"node",
"npm",
"pnpm",
"bun",
"python",
"python3",
"pip",
"pip3",
"cargo",
"rustc",
"make",
"docker",
"java",
"mvn",
"gradle",
"zig",
}

report := Report{
Name:     "AIFT Capability Discovery",
Time:     time.Now().Format(time.RFC3339),
Root:     cfg.Root,
OSHome:   cfg.OSHome,
Verified: true,
}

for _, name := range names {
path, err := exec.LookPath(name)
capability := Capability{
Name:      name,
Command:   name,
Installed: err == nil,
Path:      path,
}
report.Capabilities = append(report.Capabilities, capability)
}

return report
}

func Run(cfg config.Config) error {
report := Discover(cfg)
if err := Write(cfg, report); err != nil {
return err
}

fmt.Println("AIFT Capability Discovery")
for _, cap := range report.Capabilities {
status := "missing"
if cap.Installed {
status = "installed"
}
fmt.Printf("%-10s %s\n", cap.Name, status)
}
return nil
}

func Write(cfg config.Config, report Report) error {
outDir := filepath.Join(cfg.OSHome, "registry", "capabilities")
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

jsonPath := filepath.Join(outDir, "capabilities.json")
mdPath := filepath.Join(reportDir, "capabilities.md")

if err := os.WriteFile(jsonPath, append(b, '\n'), 0644); err != nil {
return err
}

md := "# AIFT Capability Discovery Report\n\n"
for _, cap := range report.Capabilities {
status := "missing"
if cap.Installed {
status = "installed"
}
md += "- " + cap.Name + ": " + status + "\n"
}

return os.WriteFile(mdPath, []byte(md), 0644)
}

func Has(report Report, name string) bool {
for _, cap := range report.Capabilities {
if cap.Name == name && cap.Installed {
return true
}
}
return false
}
GO

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

"$MODULE/internal/capability"
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
Waiting   []string \`json:"waiting"\`
Artifacts []string \`json:"artifacts"\`
}

type Report struct {
Name         string                  \`json:"name"\`
Time         string                  \`json:"time"\`
Root         string                  \`json:"root"\`
OSHome       string                  \`json:"os_home"\`
Verified     bool                    \`json:"verified"\`
Mode         string                  \`json:"mode"\`
Modules      []Module                \`json:"modules"\`
Blocked      []string                \`json:"blocked"\`
Capabilities capability.Report       \`json:"capabilities"\`
}

func Run(cfg config.Config, async bool) error {
caps := capability.Discover(cfg)
_ = capability.Write(cfg, caps)

mode := "sync"
if async {
mode = "async-planned"
}

modules, blocked := discover(cfg.Root, caps)

report := Report{
Name:         "AIFT Federation Build Engine",
Time:         time.Now().Format(time.RFC3339),
Root:         cfg.Root,
OSHome:       cfg.OSHome,
Verified:     true,
Mode:         mode,
Modules:      modules,
Blocked:      blocked,
Capabilities: caps,
}

for i := range report.Modules {
module := &report.Modules[i]
module.Mode = mode

if len(module.Blocked) > 0 {
module.Status = "blocked"
report.Verified = false
continue
}

if len(module.Waiting) > 0 {
module.Status = "planned"
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
fmt.Println("planned:", countPlanned(report.Modules))

if !report.Verified {
return fmt.Errorf("federation build completed with blocked modules")
}

return nil
}

func discover(root string, caps capability.Report) ([]Module, []string) {
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
modules = append(modules, inspect(path, caps))
return filepath.SkipDir
}

return nil
})

if err != nil {
blocked = append(blocked, err.Error())
}

return modules, blocked
}

func inspect(path string, caps capability.Report) Module {
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
require(&module, caps, "go")

case exists(filepath.Join(path, "pnpm-lock.yaml")):
module.Runtime = "node-pnpm"
module.Build = []string{"pnpm build"}
module.Test = []string{"pnpm test"}
require(&module, caps, "node")
require(&module, caps, "pnpm")

case exists(filepath.Join(path, "package-lock.json")):
module.Runtime = "node-npm"
module.Build = []string{"npm run build"}
module.Test = []string{"npm test"}
require(&module, caps, "node")
require(&module, caps, "npm")

case exists(filepath.Join(path, "package.json")):
module.Runtime = "node"
module.Build = []string{"npm run build"}
module.Test = []string{"npm test"}
require(&module, caps, "node")
require(&module, caps, "npm")

case exists(filepath.Join(path, "Cargo.toml")):
module.Runtime = "rust"
module.Build = []string{"cargo build"}
module.Test = []string{"cargo test"}
require(&module, caps, "cargo")

case exists(filepath.Join(path, "pyproject.toml")):
module.Runtime = "python"
module.Build = []string{"python -m build"}
module.Test = []string{"python -m pytest"}
require(&module, caps, "python")

case exists(filepath.Join(path, "requirements.txt")):
module.Runtime = "python"
module.Build = []string{"python -m compileall ."}
module.Test = []string{"python -m pytest"}
require(&module, caps, "python")

case exists(filepath.Join(path, "Makefile")):
module.Runtime = "make"
module.Build = []string{"make"}
module.Test = []string{"make test"}
require(&module, caps, "make")

default:
module.Runtime = "unknown"
module.Status = "unsupported"
module.Waiting = append(module.Waiting, "no supported build provider detected")
}

if !exists(filepath.Join(path, "aift.repo.json")) && !exists(filepath.Join(path, ".aift", "module.json")) {
module.Waiting = append(module.Waiting, "missing AIFT manifest")
}

return module
}

func require(module *Module, caps capability.Report, name string) {
if !capability.Has(caps, name) {
module.Waiting = append(module.Waiting, "missing capability: "+name)
}
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

if len(module.Waiting) > 0 {
md += "  - waiting: " + strings.Join(module.Waiting, ", ") + "\n"
}

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

func countPlanned(modules []Module) int {
total := 0
for _, module := range modules {
if len(module.Waiting) > 0 || module.Status == "planned" || module.Status == "unsupported" {
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

if f'"{module}/internal/capability"' not in s:
    pos = s.find("import (")
    line = s.find("\\n", pos) + 1
    s = s[:line] + f'\\t"{module}/internal/capability"\\n' + s[line:]

if 'case "capability":' not in s:
    target = 'case "federation-build":'
    i = s.find(target)
    if i == -1:
        target = 'case "build":'
        i = s.find(target)
    if i == -1:
        raise SystemExit('build command case not found')

    block = '''\tcase "capability":
\t\tif err := capability.Run(cfg); err != nil {
\t\t\tpanic(err)
\t\t}
\t'''

    s = s[:i] + block + s[i:]

if 'fmt.Println("  capability")' not in s:
    s = s.replace(
        'fmt.Println("  build")',
        'fmt.Println("  build")\\n\\tfmt.Println("  capability")'
    )

p.write_text(s)
PY

gofmt -w internal/capability/capability.go internal/fedbuild/fedbuild.go cmd/aift/main.go

go test ./...
go build -o "$HOME/.local/bin/aift" ./cmd/aift
hash -r

aift capability
aift federation-build --async || true
aift build || true
aift verify

git add internal/capability internal/fedbuild cmd/aift/main.go docs/architecture/AIFT-CAPABILITY-AWARE-BUILD.md registry/capabilities registry/fedbuild reports/capabilities.md reports/federation-build.md 2>/dev/null || true
git add registry/builds reports/build-report.md registry/lifecycle reports/federation-lifecycle.md registry/compiler reports/repository-compiler-report.md var/events/events.jsonl 2>/dev/null || true

git commit -m "feat: add capability-aware federation build planning" || true
git push origin main

echo "DONE: capability-aware federation build planning added"
