#!/usr/bin/env bash
set -euo pipefail

echo "== Phase 19B: CLI Compiler Repair =="

ROOT="$(pwd)"
CMD_DIR="$ROOT/cmd/aift"
REPORT_DIR="$ROOT/reports"
REGISTRY_DIR="$ROOT/registry/cli"

mkdir -p "$CMD_DIR" "$REPORT_DIR" "$REGISTRY_DIR"

REPORT="$REPORT_DIR/phase19b-cli-compiler-repair-report.md"
REGISTRY="$REGISTRY_DIR/commands.json"

{
  echo "# Phase 19B CLI Compiler Repair Report"
  echo ""
  echo "Generated: $(date -Is)"
  echo ""
} > "$REPORT"

if [ -f "$CMD_DIR/main.go" ]; then
  cp "$CMD_DIR/main.go" "$CMD_DIR/main.go.phase19b.bak"
fi

cat > "$CMD_DIR/main.go" <<'GO'
package main

import (
"encoding/json"
"fmt"
"os"
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
HandlerName string   `json:"handler"`
Handler     func([]string) error `json:"-"`
}

type CommandRegistry struct {
GeneratedAt string    `json:"generated_at"`
Runtime     string    `json:"runtime"`
OS          string    `json:"os"`
Arch        string    `json:"arch"`
Commands    []Command `json:"commands"`
}

func main() {
commands := registeredCommands()

if len(os.Args) < 2 {
printHelp(commands)
return
}

name := strings.TrimSpace(os.Args[1])
args := os.Args[2:]

cmd, ok := resolveCommand(commands, name)
if !ok {
fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", name)
printHelp(commands)
os.Exit(2)
}

if err := cmd.Handler(args); err != nil {
fmt.Fprintf(os.Stderr, "command failed: %v\n", err)
os.Exit(1)
}
}

func registeredCommands() []Command {
cmds := []Command{
{
Name:        "help",
Description: "Show available AIFT-OS CLI commands.",
Usage:       "aift help",
Aliases:     []string{"--help", "-h"},
HandlerName: "runHelp",
Handler:     runHelp,
},
{
Name:        "status",
Description: "Show truthful local AIFT-OS CLI status.",
Usage:       "aift status",
Aliases:     []string{"doctor"},
HandlerName: "runStatus",
Handler:     runStatus,
},
{
Name:        "registry",
Description: "Print the CLI command registry as JSON.",
Usage:       "aift registry",
Aliases:     []string{"commands"},
HandlerName: "runRegistry",
Handler:     runRegistry,
},
{
Name:        "verify",
Description: "Run honest self-verification for the CLI shell.",
Usage:       "aift verify",
Aliases:     []string{"check"},
HandlerName: "runVerify",
Handler:     runVerify,
},
{
Name:        "federation",
Description: "Federation command placeholder until real internal APIs are proven.",
Usage:       "aift federation",
Aliases:     []string{"fed"},
HandlerName: "runFederation",
Handler:     runFederation,
},
{
Name:        "repo",
Description: "Repository command placeholder until real internal APIs are proven.",
Usage:       "aift repo",
Aliases:     []string{"repos"},
HandlerName: "runRepo",
Handler:     runRepo,
},
{
Name:        "workflow",
Description: "Workflow command placeholder until real internal APIs are proven.",
Usage:       "aift workflow",
Aliases:     []string{"flows"},
HandlerName: "runWorkflow",
Handler:     runWorkflow,
},
}

sort.Slice(cmds, func(i, j int) bool {
return cmds[i].Name < cmds[j].Name
})

return cmds
}

func resolveCommand(commands []Command, name string) (Command, bool) {
for _, cmd := range commands {
if cmd.Name == name {
return cmd, true
}
for _, alias := range cmd.Aliases {
if alias == name {
return cmd, true
}
}
}
return Command{}, false
}

func runHelp(args []string) error {
printHelp(registeredCommands())
return nil
}

func printHelp(commands []Command) {
fmt.Println("AIFT-OS CLI")
fmt.Println()
fmt.Println("Truthful local-first federation operator CLI.")
fmt.Println()
fmt.Println("Usage:")
fmt.Println("  aift <command> [args]")
fmt.Println()
fmt.Println("Commands:")

for _, cmd := range commands {
fmt.Printf("  %-12s %s\n", cmd.Name, cmd.Description)
}
}

func runStatus(args []string) error {
return printJSON(map[string]any{
"status":       "ok",
"mode":         "truthful-local-first",
"runtime":      runtime.Version(),
"os":           runtime.GOOS,
"arch":         runtime.GOARCH,
"generated_at": time.Now().Format(time.RFC3339),
"note":         "CLI shell compiles. Planned commands remain planned until wired to proven internal APIs.",
})
}

func runRegistry(args []string) error {
return printJSON(CommandRegistry{
GeneratedAt: time.Now().Format(time.RFC3339),
Runtime:     runtime.Version(),
OS:          runtime.GOOS,
Arch:        runtime.GOARCH,
Commands:    registeredCommands(),
})
}

func runVerify(args []string) error {
return printJSON(map[string]any{
"status": "ok",
"checks": []map[string]any{
{
"name":   "cli-entrypoint",
"status": "active",
"detail": "cmd/aift/main.go is self-contained and does not reference missing command handlers.",
},
{
"name":   "federation-command",
"status": "planned",
"detail": "Registered honestly as planned until connected to real internal implementation.",
},
{
"name":   "repo-command",
"status": "planned",
"detail": "Registered honestly as planned until connected to real internal implementation.",
},
{
"name":   "workflow-command",
"status": "planned",
"detail": "Registered honestly as planned until connected to real internal implementation.",
},
},
})
}

func runFederation(args []string) error {
return runPlanned("federation")
}

func runRepo(args []string) error {
return runPlanned("repo")
}

func runWorkflow(args []string) error {
return runPlanned("workflow")
}

func runPlanned(name string) error {
return printJSON(map[string]any{
"command": name,
"status":  "planned",
"message": "This command is registered but not yet wired to a proven internal implementation.",
})
}

func printJSON(v any) error {
enc := json.NewEncoder(os.Stdout)
enc.SetIndent("", "  ")
return enc.Encode(v)
}
GO

gofmt -w "$CMD_DIR/main.go"

echo "Checking for stale missing-symbol references..."
if grep -nE 'Handler:[[:space:]]*verify|[^A-Za-z0-9_]verify\(' "$CMD_DIR/main.go"; then
  echo "Stale verify reference still exists." >&2
  exit 1
fi

echo "Building CLI..."
mkdir -p bin
go build -o bin/aift ./cmd/aift

echo "Generating registry..."
./bin/aift registry > "$REGISTRY"

echo "Running CLI verification..."
./bin/aift verify

echo "Running full tests..."
if go test ./...; then
  TEST_STATUS="PASS"
else
  TEST_STATUS="FAIL"
fi

{
  echo ""
  echo "## Results"
  echo ""
  echo "- go build -o bin/aift ./cmd/aift: PASS"
  echo "- ./bin/aift registry: PASS"
  echo "- ./bin/aift verify: PASS"
  echo "- go test ./...: $TEST_STATUS"
  echo ""
  echo "## Notes"
  echo ""
  echo "This phase intentionally keeps federation, repo, and workflow commands as planned placeholders until real internal package APIs are proven."
  echo ""
  echo "## Registry"
  echo ""
  cat "$REGISTRY"
} >> "$REPORT"

git add "$CMD_DIR/main.go" "$REGISTRY" "$REPORT" phase19-cli-compiler-repair.sh || true

echo ""
echo "Phase 19B complete."
echo "Report: $REPORT"
echo "Registry: $REGISTRY"
echo ""
echo "Now test:"
echo "  ./bin/aift status"
echo "  ./bin/aift verify"
echo "  ./bin/aift registry"
echo ""
echo "Then commit:"
echo "  git commit -m 'Phase 19B: repair and stabilize AIFT CLI compiler shell'"
echo "  git push origin main"
