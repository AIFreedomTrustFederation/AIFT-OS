#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

export AIFT_ROOT="$ROOT"
export AIFT_OS_HOME="$OS"

cd "$OS" || exit 1

echo "== AIFT-OS Phase 9: Truthful Federation Capability System =="

mkdir -p \
  internal/capabilities \
  docs \
  tests \
  registry \
  reports \
  var/events \
  AI-Code-Training/scripts/phase-scripts \
  bin

cat > internal/capabilities/capabilities.go <<'GO'
package capabilities

import (
"context"
"encoding/json"
"fmt"
"os"
"os/exec"
"path/filepath"
"strings"
"time"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

const (
StatusPlanned  = "planned"
StatusDetected = "detected"
StatusReady    = "ready"
StatusV1       = "v1"
StatusBroken   = "broken"
StatusMissing  = "missing"
)

type Capability struct {
Name        string `json:"name"`
Status      string `json:"status"`
Version     int    `json:"version"`
Command     string `json:"command,omitempty"`
Evidence    string `json:"evidence,omitempty"`
Description string `json:"description,omitempty"`
LastChecked string `json:"lastChecked"`
}

type RepoCapabilities struct {
Repo         string       `json:"repo"`
Path         string       `json:"path"`
Capabilities []Capability `json:"capabilities"`
}

type FederationCapabilities struct {
GeneratedAt string             `json:"generatedAt"`
Repos       []RepoCapabilities `json:"repos"`
}

func capabilityNames() []string {
return []string{
"status",
"verify",
"test",
"build",
"start",
"stop",
"health",
"deploy",
"sync",
"docs",
}
}

func Scan(cfg config.Config) error {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return err
}

all := FederationCapabilities{
GeneratedAt: time.Now().Format(time.RFC3339),
Repos:       []RepoCapabilities{},
}

for _, r := range repos {
rc, err := ScanRepo(cfg, r)
if err != nil {
return err
}
all.Repos = append(all.Repos, rc)
}

if err := writeGlobal(cfg, all); err != nil {
return err
}

if err := writeReport(cfg, all); err != nil {
return err
}

return events.Emit(cfg, "capabilities.scan", "capabilities", "federation capability scan complete", map[string]string{
"repos": fmt.Sprint(len(all.Repos)),
})
}

func ScanRepo(cfg config.Config, r workspace.Repo) (RepoCapabilities, error) {
now := time.Now().Format(time.RFC3339)
old := readExisting(r.Path)

rc := RepoCapabilities{
Repo:         r.Name,
Path:         r.Path,
Capabilities: []Capability{},
}

for _, name := range capabilityNames() {
prev := old[name]
cap := detectCapability(r.Path, name)
cap.LastChecked = now

if prev.Status == StatusV1 {
if cap.Status == StatusReady || cap.Status == StatusDetected {
if cap.Command != "" && commandPasses(r.Path, cap.Command) {
cap.Status = StatusV1
cap.Version = 1
cap.Evidence = "previously promoted to v1 and verification still passes"
} else if cap.Command != "" {
cap.Status = StatusBroken
cap.Version = 1
cap.Evidence = "was v1, but current verification failed"
}
}
}

rc.Capabilities = append(rc.Capabilities, cap)
}

if err := writeRepo(r.Path, rc); err != nil {
return rc, err
}

return rc, nil
}

func detectCapability(repoPath, name string) Capability {
c := Capability{
Name:        name,
Status:      StatusPlanned,
Version:     0,
Description: description(name),
}

cmdPath := filepath.Join(repoPath, ".aift", "commands", name+".sh")
if fileExists(cmdPath) {
c.Command = ".aift/commands/" + name + ".sh"
if commandPasses(repoPath, cmdPath) {
c.Status = StatusReady
c.Evidence = "command exists and passes local verification"
} else {
c.Status = StatusBroken
c.Evidence = "command exists but failed local verification"
}
return c
}

switch name {
case "test":
if fileExists(filepath.Join(repoPath, "package.json")) {
c.Status = StatusDetected
c.Evidence = "package.json detected; test capability may exist but no .aift command is proven"
return c
}
if fileExists(filepath.Join(repoPath, "go.mod")) {
c.Status = StatusDetected
c.Evidence = "go.mod detected; Go tests may exist but no .aift command is proven"
return c
}
case "build":
if fileExists(filepath.Join(repoPath, "package.json")) || fileExists(filepath.Join(repoPath, "go.mod")) || fileExists(filepath.Join(repoPath, "Makefile")) {
c.Status = StatusDetected
c.Evidence = "build-related project file detected but no .aift build command is proven"
return c
}
case "docs":
if fileExists(filepath.Join(repoPath, "README.md")) || dirExists(filepath.Join(repoPath, "docs")) {
c.Status = StatusDetected
c.Evidence = "README/docs detected"
return c
}
case "status":
if dirExists(filepath.Join(repoPath, ".git")) {
c.Status = StatusReady
c.Command = "git status --short"
c.Evidence = "git repository detected; built-in status capability is ready"
return c
}
case "sync":
if dirExists(filepath.Join(repoPath, ".git")) {
c.Status = StatusReady
c.Command = "git remote/status"
c.Evidence = "git repository detected; safe sync can inspect this repo"
return c
}
}

c.Evidence = "not proven yet"
return c
}

func Promote(cfg config.Config, repoName, capName string) error {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return err
}

for _, r := range repos {
if r.Name != repoName {
continue
}

rc, err := ScanRepo(cfg, r)
if err != nil {
return err
}

changed := false
for i := range rc.Capabilities {
if rc.Capabilities[i].Name != capName {
continue
}

if rc.Capabilities[i].Status != StatusReady {
return fmt.Errorf("cannot promote %s/%s: status is %s, not ready", repoName, capName, rc.Capabilities[i].Status)
}

rc.Capabilities[i].Status = StatusV1
rc.Capabilities[i].Version = 1
rc.Capabilities[i].Evidence = "promoted to v1 after passing local verification"
rc.Capabilities[i].LastChecked = time.Now().Format(time.RFC3339)
changed = true
}

if !changed {
return fmt.Errorf("capability not found: %s", capName)
}

if err := writeRepo(r.Path, rc); err != nil {
return err
}

if err := Scan(cfg); err != nil {
return err
}

return events.Emit(cfg, "capability.promote", "capabilities", "capability promoted to v1", map[string]string{
"repo":       repoName,
"capability": capName,
})
}

return fmt.Errorf("repository not found: %s", repoName)
}

func Report(cfg config.Config) error {
data, err := os.ReadFile(filepath.Join(cfg.OSHome, "reports", "capabilities.md"))
if err != nil {
if err := Scan(cfg); err != nil {
return err
}
data, err = os.ReadFile(filepath.Join(cfg.OSHome, "reports", "capabilities.md"))
if err != nil {
return err
}
}
fmt.Print(string(data))
return nil
}

func PrintRepo(cfg config.Config, repoName string) error {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return err
}

for _, r := range repos {
if r.Name != repoName {
continue
}
rc, err := ScanRepo(cfg, r)
if err != nil {
return err
}
printRepo(rc)
return nil
}

return fmt.Errorf("repository not found: %s", repoName)
}

func printRepo(rc RepoCapabilities) {
fmt.Println("Repository:", rc.Repo)
fmt.Println("Path:", rc.Path)
fmt.Printf("%-14s %-10s %-8s %s\n", "CAPABILITY", "STATUS", "VERSION", "EVIDENCE")
for _, c := range rc.Capabilities {
fmt.Printf("%-14s %-10s %-8d %s\n", c.Name, c.Status, c.Version, c.Evidence)
}
}

func writeRepo(repoPath string, rc RepoCapabilities) error {
dir := filepath.Join(repoPath, ".aift")
if err := os.MkdirAll(dir, 0755); err != nil {
return err
}
data, err := json.MarshalIndent(rc, "", "  ")
if err != nil {
return err
}
return os.WriteFile(filepath.Join(dir, "capabilities.json"), append(data, '\n'), 0644)
}

func writeGlobal(cfg config.Config, all FederationCapabilities) error {
out := filepath.Join(cfg.OSHome, "registry", "capabilities.json")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}
data, err := json.MarshalIndent(all, "", "  ")
if err != nil {
return err
}
if err := os.WriteFile(out, append(data, '\n'), 0644); err != nil {
return err
}
fmt.Println("Wrote", out)
return nil
}

func writeReport(cfg config.Config, all FederationCapabilities) error {
out := filepath.Join(cfg.OSHome, "reports", "capabilities.md")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}

var b strings.Builder
b.WriteString("# Federation Capabilities\n\n")
b.WriteString("Statuses: `planned`, `detected`, `ready`, `v1`, `broken`, `missing`.\n\n")

for _, repo := range all.Repos {
b.WriteString("## " + repo.Repo + "\n\n")
b.WriteString("| Capability | Status | Version | Evidence |\n")
b.WriteString("|---|---|---:|---|\n")
for _, c := range repo.Capabilities {
b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%d` | %s |\n", c.Name, c.Status, c.Version, c.Evidence))
}
b.WriteString("\n")
}

if err := os.WriteFile(out, []byte(b.String()), 0644); err != nil {
return err
}
fmt.Println("Wrote", out)
return nil
}

func readExisting(repoPath string) map[string]Capability {
out := map[string]Capability{}
data, err := os.ReadFile(filepath.Join(repoPath, ".aift", "capabilities.json"))
if err != nil {
return out
}

var rc RepoCapabilities
if json.Unmarshal(data, &rc) != nil {
return out
}

for _, c := range rc.Capabilities {
out[c.Name] = c
}
return out
}

func commandPasses(repoPath, commandPath string) bool {
ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
defer cancel()

var cmd *exec.Cmd
if strings.Contains(commandPath, " ") {
cmd = exec.CommandContext(ctx, "sh", "-c", commandPath)
} else {
cmd = exec.CommandContext(ctx, "sh", commandPath)
}
cmd.Dir = repoPath
cmd.Env = append(os.Environ(), "AIFT_CAPABILITY_CHECK=1")
return cmd.Run() == nil
}

func description(name string) string {
switch name {
case "status":
return "Report repository state"
case "verify":
return "Validate repository health"
case "test":
return "Run test suite"
case "build":
return "Build project artifacts"
case "start":
return "Start local service"
case "stop":
return "Stop local service"
case "health":
return "Check local service health"
case "deploy":
return "Deploy project"
case "sync":
return "Synchronize safely"
case "docs":
return "Documentation present or generated"
default:
return "Capability"
}
}

func fileExists(path string) bool {
info, err := os.Stat(path)
return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
info, err := os.Stat(path)
return err == nil && info.IsDir()
}
GO

cat > cmd/aift/main.go <<'GO'
package main

import (
"fmt"
"os"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/api"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/capabilities"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/daemon"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/doctor"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/federation"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/gitx"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/plugins"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/providers"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/repo"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/runtime"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/services"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/sync"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/version"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workflow"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

func main() {
cfg := config.Load()
cmd := "help"
args := []string{}

if len(os.Args) > 1 {
cmd = os.Args[1]
args = os.Args[2:]
}

if looksLikeExecutablePath(cmd) {
if len(args) > 0 {
cmd = args[0]
args = args[1:]
} else {
cmd = "help"
}
}

var err error

switch cmd {
case "help", "-h", "--help":
help()
case "version":
fmt.Printf("%s %s — %s\n", version.Name, version.Version, version.Role)
case "doctor":
err = doctor.Run(cfg)
case "status":
err = status(cfg)
case "manifest":
err = manifests.EnsureAll(cfg)
if err == nil {
fmt.Println("OK: manifests ensured")
}
case "registry":
err = registry.Generate(cfg)
case "dashboard":
err = reports.Dashboard(cfg)
case "deps":
err = reports.Deps(cfg)
case "plugins":
err = plugins.List(cfg)
case "providers":
err = providers.List(cfg)
case "events":
err = events.Tail(cfg, 25)
case "services":
err = services.List(cfg)
case "start":
err = runtime.StartOnce(cfg)
case "tick":
err = runtime.Tick(cfg)
case "serve":
addr := ":8787"
if len(args) > 0 {
addr = args[0]
}
err = api.New(cfg, addr).Serve()
case "daemon":
addr := ":8787"
if len(args) > 0 {
addr = args[0]
}
err = daemon.Start(cfg, addr)
case "sync":
if len(args) == 0 || args[0] == "--safe" || args[0] == "safe" {
err = sync.Safe(cfg)
} else {
err = fmt.Errorf("only sync --safe is implemented in Go kernel")
}
case "federation":
err = runFederation(cfg, args)
case "repo":
err = runRepo(cfg, args)
case "workflow":
err = runWorkflow(cfg, args)
case "capabilities":
err = runCapabilities(cfg, args)
case "verify":
err = verify(cfg)
default:
err = fmt.Errorf("unknown command: %s", cmd)
}

if err != nil {
fmt.Fprintln(os.Stderr, "ERROR:", err)
os.Exit(1)
}
}

func looksLikeExecutablePath(s string) bool {
return len(s) > 0 && (s[0] == '/' || s == "aiftd" || s == "./aiftd" || s == "bin/aiftd")
}

func help() {
fmt.Println("AIFT-OS Federation Control Plane")
fmt.Println()
fmt.Println("Commands:")
fmt.Println("  help")
fmt.Println("  version")
fmt.Println("  doctor")
fmt.Println("  status")
fmt.Println("  manifest")
fmt.Println("  registry")
fmt.Println("  dashboard")
fmt.Println("  deps")
fmt.Println("  plugins")
fmt.Println("  providers")
fmt.Println("  events")
fmt.Println("  services")
fmt.Println("  start")
fmt.Println("  tick")
fmt.Println("  serve [:8787]")
fmt.Println("  daemon [:8787]")
fmt.Println("  sync --safe")
fmt.Println("  federation scan|graph|verify")
fmt.Println("  repo list|inspect|run")
fmt.Println("  workflow list")
fmt.Println("  capabilities scan|report|repo|promote")
fmt.Println("  verify")
}

func runFederation(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
return federation.Scan(cfg)
}
if args[0] == "graph" {
return federation.Graph(cfg)
}
if args[0] == "verify" {
return federation.Verify(cfg)
}
return fmt.Errorf("usage: aift federation scan|graph|verify")
}

func runRepo(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "list" {
return repo.PrintList(cfg)
}
if args[0] == "inspect" {
if len(args) < 2 {
return fmt.Errorf("usage: aift repo inspect <name>")
}
return repo.PrintInspect(cfg, repo.NormalizeName(args[1]))
}
if args[0] == "run" {
if len(args) < 3 {
return fmt.Errorf("usage: aift repo run <name> <command> [args...]")
}
return repo.RunCommand(cfg, repo.NormalizeName(args[1]), args[2], args[3:])
}
return fmt.Errorf("usage: aift repo list|inspect|run")
}

func runWorkflow(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "list" {
return workflow.List(cfg)
}
return fmt.Errorf("usage: aift workflow list")
}

func runCapabilities(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
return capabilities.Scan(cfg)
}
if args[0] == "report" {
return capabilities.Report(cfg)
}
if args[0] == "repo" {
if len(args) < 2 {
return fmt.Errorf("usage: aift capabilities repo <repo>")
}
return capabilities.PrintRepo(cfg, args[1])
}
if args[0] == "promote" {
if len(args) < 3 {
return fmt.Errorf("usage: aift capabilities promote <repo> <capability>")
}
return capabilities.Promote(cfg, args[1], args[2])
}
return fmt.Errorf("usage: aift capabilities scan|report|repo|promote")
}

func verify(cfg config.Config) error {
if err := doctor.Run(cfg); err != nil {
return err
}
if err := manifests.EnsureAll(cfg); err != nil {
return err
}
if err := providers.WriteRegistry(cfg); err != nil {
return err
}
if err := registry.Generate(cfg); err != nil {
return err
}
if err := reports.Dashboard(cfg); err != nil {
return err
}
if err := reports.Deps(cfg); err != nil {
return err
}
if err := capabilities.Scan(cfg); err != nil {
return err
}
if err := events.Emit(cfg, "verify.complete", "verify", "federation verified", nil); err != nil {
return err
}
fmt.Println("OK: federation verified")
return nil
}

func status(cfg config.Config) error {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return err
}

fmt.Printf("%-32s %-12s %-8s %s\n", "REPOSITORY", "BRANCH", "STATE", "REMOTE")
for _, repo := range repos {
state := "clean"
if gitx.Dirty(repo.Path) {
state = "dirty"
}
fmt.Printf("%-32s %-12s %-8s %s\n", repo.Name, gitx.Branch(repo.Path), state, gitx.Remote(repo.Path))
}

return nil
}
GO

cat > tests/capabilities-smoke.sh <<'SH'
#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" capabilities scan >/dev/null
"$ROOT/aift" capabilities report >/dev/null
"$ROOT/aift" capabilities repo AIFT-OS >/dev/null
"$ROOT/aift" verify >/dev/null

test -f registry/capabilities.json
test -f reports/capabilities.md
test -f .aift/capabilities.json

echo "OK: capabilities smoke passed"
SH

chmod +x tests/capabilities-smoke.sh

cat > docs/PHASE-9-CAPABILITIES.md <<'DOC'
# Phase 9: Truthful Federation Capabilities

AIFT-OS now audits what each sovereign repository can actually do.

Capability statuses:

- `planned` — intended, but not proven executable
- `detected` — inferred from repo files
- `ready` — executable exists and local verification passes
- `v1` — ready capability has been explicitly promoted/versioned
- `broken` — capability was expected or promoted but current verification fails
- `missing` — not present

Commands:

- `aift capabilities scan`
- `aift capabilities report`
- `aift capabilities repo <repo>`
- `aift capabilities promote <repo> <capability>`

Generated files:

- `registry/capabilities.json`
- `reports/capabilities.md`
- `<repo>/.aift/capabilities.json`

AIFT-OS should orchestrate only capabilities marked `ready` or `v1`.
DOC

cp phase-9-capabilities.sh AI-Code-Training/scripts/phase-scripts/phase-9-capabilities.sh 2>/dev/null || true

echo "== Build/test =="
go clean -cache
gofmt -w cmd internal
go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" capabilities scan
"$ROOT/aift" capabilities repo AIFT-OS
sh tests/capabilities-smoke.sh

echo "== Commit/push =="
git add .
if git diff --cached --quiet; then
  echo "Nothing staged to commit."
else
  git commit -m "Add truthful federation capability audit system"
fi

git push origin main

echo
echo "DONE."
echo "Try:"
echo "  ~/AIFT/aift capabilities report"
echo "  ~/AIFT/aift capabilities repo AIFT-Forge"
echo "  ~/AIFT/aift capabilities promote AIFT-OS status"
