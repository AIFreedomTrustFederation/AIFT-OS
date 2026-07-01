#!/usr/bin/env bash
set -euo pipefail

echo "== Phase 20C: Fix CLI argv dispatch =="

ROOT="$(pwd)"
CMD_DIR="$ROOT/cmd/aift"
LEGACY_DIR="$ROOT/legacy/cmd-aift-phase20c"
BIN_DIR="$ROOT/bin"
REPORT_DIR="$ROOT/reports"
REGISTRY_DIR="$ROOT/registry/cli"
BOOTSTRAP_DIR="$ROOT/registry/bootstrap"

mkdir -p "$CMD_DIR" "$LEGACY_DIR" "$BIN_DIR" "$REPORT_DIR" "$REGISTRY_DIR" "$BOOTSTRAP_DIR"

REPORT="$REPORT_DIR/phase20c-fix-cli-argv-dispatch-report.md"
REGISTRY="$REGISTRY_DIR/commands.json"
BOOTSTRAP="$BOOTSTRAP_DIR/federation-bootstrap.json"
BIN="$BIN_DIR/aift"

cp "$CMD_DIR/main.go" "$CMD_DIR/main.go.phase20c.bak" 2>/dev/null || true

echo "Quarantining extra compiled CLI files..."
find "$CMD_DIR" -maxdepth 1 -name '*.go' ! -name 'main.go' -print0 | while IFS= read -r -d '' f; do
  mv "$f" "$LEGACY_DIR/"
done

cat > "$CMD_DIR/main.go" <<'GO'
package main

import (
"encoding/json"
"fmt"
"os"
"os/exec"
"path/filepath"
"runtime"
"sort"
"strings"
"time"
)

type Command struct {
Name        string   `json:"name"`
Description string   `json:"description"`
Usage       string   `json:"usage"`
Aliases     []string `json:"aliases,omitempty"`
Status      string   `json:"status"`
Handler     func([]string) error `json:"-"`
}

type Check struct {
Name   string `json:"name"`
Status string `json:"status"`
Detail string `json:"detail"`
}

func main() {
code := run(os.Args)
if code != 0 {
os.Exit(code)
}
}

func run(rawArgs []string) int {
cmds := commands()
args := normalizeArgs(rawArgs)

if len(args) == 0 {
printHelp(cmds)
return 0
}

name := args[0]
commandArgs := args[1:]

cmd, ok := resolve(cmds, name)
if !ok {
fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", name)
printHelp(cmds)
return 2
}

if err := cmd.Handler(commandArgs); err != nil {
fmt.Fprintf(os.Stderr, "command failed: %v\n", err)
return 1
}

return 0
}

func normalizeArgs(raw []string) []string {
if len(raw) == 0 {
return nil
}

args := append([]string{}, raw...)

if len(args) > 0 {
base := filepath.Base(args[0])
if base == "aift" || strings.HasSuffix(base, ".exe") || strings.Contains(args[0], "/") {
args = args[1:]
}
}

if len(args) > 0 && args[0] == "--" {
args = args[1:]
}

return args
}

func commands() []Command {
cmds := []Command{
{"help", "Show available commands.", "aift help", []string{"--help", "-h"}, "active", runHelp},
{"status", "Inspect the real local repository status.", "aift status", []string{"doctor"}, "active", runStatus},
{"verify", "Run real local bootstrap verification checks.", "aift verify", []string{"check"}, "active", runVerify},
{"registry", "Print command registry as JSON.", "aift registry", []string{"commands"}, "active", runRegistry},
{"bootstrap", "Print federation bootstrap discovery JSON.", "aift bootstrap", nil, "active", runBootstrap},
{"federation", "Federation command group; planned until real APIs are proven.", "aift federation", []string{"fed"}, "planned", planned("federation")},
{"repo", "Repository command group; planned until real APIs are proven.", "aift repo", []string{"repos"}, "planned", planned("repo")},
{"workflow", "Workflow command group; planned until real APIs are proven.", "aift workflow", []string{"flows"}, "planned", planned("workflow")},
}

sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })
return cmds
}

func resolve(cmds []Command, name string) (Command, bool) {
for _, c := range cmds {
if c.Name == name {
return c, true
}
for _, a := range c.Aliases {
if a == name {
return c, true
}
}
}
return Command{}, false
}

func runHelp(args []string) error {
printHelp(commands())
return nil
}

func printHelp(cmds []Command) {
fmt.Println("AIFT-OS CLI")
fmt.Println()
fmt.Println("Truthful local-first federation operator CLI.")
fmt.Println()
fmt.Println("Usage:")
fmt.Println("  aift <command> [args]")
fmt.Println()
fmt.Println("Commands:")
for _, c := range cmds {
fmt.Printf("  %-12s %-8s %s\n", c.Name, c.Status, c.Description)
}
}

func runRegistry(args []string) error {
return printJSON(map[string]any{
"generated_at": time.Now().Format(time.RFC3339),
"runtime":      runtime.Version(),
"os":           runtime.GOOS,
"arch":         runtime.GOARCH,
"commands":     commands(),
})
}

func runStatus(args []string) error {
root, _ := os.Getwd()
checks := collectChecks()

return printJSON(map[string]any{
"status":       aggregate(checks),
"generated_at": time.Now().Format(time.RFC3339),
"root":         root,
"runtime":      runtime.Version(),
"os":           runtime.GOOS,
"arch":         runtime.GOARCH,
"checks":       checks,
})
}

func runVerify(args []string) error {
checks := collectChecks()
status := aggregate(checks)

if err := printJSON(map[string]any{
"status": status,
"checks": checks,
}); err != nil {
return err
}

if status == "fail" {
return fmt.Errorf("verification failed")
}

return nil
}

func runBootstrap(args []string) error {
root, _ := os.Getwd()

return printJSON(map[string]any{
"generated_at": time.Now().Format(time.RFC3339),
"root":         root,
"discovery": map[string]any{
"git":       exists(".git"),
"go_mod":    exists("go.mod"),
"package":   exists("package.json"),
"registry":  exists("registry"),
"internal":  exists("internal"),
"cmd_aift":  exists("cmd/aift/main.go"),
"reports":   exists("reports"),
"scripts":   exists("scripts"),
"manifests": exists("manifests"),
"workflows": exists(".github/workflows"),
},
"commands": commands(),
})
}

func collectChecks() []Check {
checks := []Check{
fileCheck("git-repository", ".git", "Local Git repository exists."),
fileCheck("go-module", "go.mod", "Go module manifest exists."),
fileCheck("cli-entrypoint", "cmd/aift/main.go", "AIFT CLI entrypoint exists."),
fileCheck("registry-directory", "registry", "Federation registry directory exists."),
fileCheck("internal-directory", "internal", "Internal package directory exists."),
fileCheck("reports-directory", "reports", "Reports directory exists."),
toolCheck("git-binary", "git"),
toolCheck("go-binary", "go"),
}

checks = append(checks, commandCheck("go-build-cmd-aift", "go", "build", "./cmd/aift"))
return checks
}

func fileCheck(name, path, detail string) Check {
if exists(path) {
return Check{name, "pass", detail}
}
return Check{name, "planned", "Missing: " + path}
}

func toolCheck(name, tool string) Check {
if _, err := exec.LookPath(tool); err == nil {
return Check{name, "pass", tool + " is available."}
}
return Check{name, "fail", tool + " is not available."}
}

func commandCheck(name string, cmd string, args ...string) Check {
c := exec.Command(cmd, args...)
out, err := c.CombinedOutput()
if err != nil {
return Check{name, "fail", strings.TrimSpace(string(out))}
}
return Check{name, "pass", strings.TrimSpace(string(out))}
}

func aggregate(checks []Check) string {
status := "pass"
for _, c := range checks {
if c.Status == "fail" {
return "fail"
}
if c.Status == "planned" {
status = "partial"
}
}
return status
}

func planned(name string) func([]string) error {
return func(args []string) error {
return printJSON(map[string]any{
"command": name,
"status":  "planned",
"message": "Registered honestly but not yet wired to a proven internal implementation.",
})
}
}

func exists(path string) bool {
_, err := os.Stat(path)
return err == nil
}

func printJSON(v any) error {
enc := json.NewEncoder(os.Stdout)
enc.SetIndent("", "  ")
return enc.Encode(v)
}
GO

gofmt -w "$CMD_DIR/main.go"

echo "Building binary..."
go build -o "$BIN" ./cmd/aift

echo "Testing binary..."
"$BIN" help
"$BIN" status
"$BIN" verify
"$BIN" registry > "$REGISTRY"
"$BIN" bootstrap > "$BOOTSTRAP"

TEST_STATUS="PASS"
go test ./... || TEST_STATUS="FAIL"

{
  echo "# Phase 20C Fix CLI argv Dispatch Report"
  echo ""
  echo "Generated: $(date -Is)"
  echo ""
  echo "## Results"
  echo ""
  echo "- go build -o bin/aift ./cmd/aift: PASS"
  echo "- bin/aift help: PASS"
  echo "- bin/aift status: PASS"
  echo "- bin/aift verify: PASS"
  echo "- bin/aift registry: PASS"
  echo "- bin/aift bootstrap: PASS"
  echo "- go test ./...: $TEST_STATUS"
  echo ""
  echo "## Quarantined Files"
  echo ""
  find "$LEGACY_DIR" -maxdepth 1 -type f -print || true
} > "$REPORT"

git add "$CMD_DIR/main.go" "$LEGACY_DIR" "$REGISTRY" "$BOOTSTRAP" "$REPORT" phase20c-fix-cli-argv-dispatch.sh || true

echo ""
echo "Phase 20C complete."
echo ""
echo "Commit:"
echo "  git commit -m 'Phase 20C: fix AIFT CLI argv dispatch'"
echo "  git push origin main"
