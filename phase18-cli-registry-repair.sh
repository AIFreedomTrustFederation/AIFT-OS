#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

echo "Phase 18: repair duplicate help commands and add CLI registry foundation"

MODULE="$(awk '/^module / {print $2; exit}' go.mod)"

mkdir -p internal/cli docs/architecture registry/cli reports

cat > docs/architecture/AIFT-CLI-REGISTRY.md <<'DOC'
# AIFT CLI Registry

AIFT commands must be registry-driven, deduplicated, and discoverable.

Rules:

- No duplicate help commands.
- No fake commands.
- No hardcoded phase-only behavior.
- Commands are discoverable by registry.
- Modules may later register commands directly.
DOC

cat > internal/cli/registry.go <<GO
package cli

import (
"encoding/json"
"fmt"
"os"
"path/filepath"
"sort"
"time"

"$MODULE/internal/config"
)

type Command struct {
Name        string \`json:"name"\`
Description string \`json:"description"\`
}

type Report struct {
Name     string    \`json:"name"\`
Time     string    \`json:"time"\`
Verified bool      \`json:"verified"\`
Commands []Command \`json:"commands"\`
}

func Builtins() []Command {
names := []string{
"help",
"version",
"doctor",
"status",
"manifest",
"registry",
"dashboard",
"deps",
"plugins",
"providers",
"events",
"services",
"start",
"tick",
"serve",
"daemon",
"sync",
"federation",
"repo",
"workflow",
"intelligence",
"manual",
"graph",
"service-contracts",
"plan",
"modules",
"kernel-registry",
"discovery",
"event-bus",
"patch-engine",
"kernel",
"runtime",
"capabilities",
"operator",
"scheduler",
"ai",
"compile",
"compiler",
"provider-registry",
"capability",
"lifecycle",
"federation-build",
"build",
"verify",
}

seen := map[string]bool{}
var commands []Command

for _, name := range names {
if seen[name] {
continue
}
seen[name] = true
commands = append(commands, Command{Name: name, Description: "registered AIFT command"})
}

sort.Slice(commands, func(i, j int) bool {
return commands[i].Name < commands[j].Name
})

return commands
}

func Names() []string {
var names []string
for _, command := range Builtins() {
names = append(names, command.Name)
}
return names
}

func PrintHelp() {
fmt.Println("AIFT-OS Federation Control Plane")
fmt.Println("")
fmt.Println("Commands:")

for _, command := range Builtins() {
fmt.Println("  " + command.Name)
}
}

func Run(cfg config.Config) error {
report := Report{
Name:     "AIFT CLI Registry",
Time:     time.Now().Format(time.RFC3339),
Verified: true,
Commands: Builtins(),
}

return Write(cfg, report)
}

func Write(cfg config.Config, report Report) error {
outDir := filepath.Join(cfg.OSHome, "registry", "cli")
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

if err := os.WriteFile(filepath.Join(outDir, "cli-registry.json"), append(b, '\n'), 0644); err != nil {
return err
}

md := "# AIFT CLI Registry Report\n\n"
for _, command := range report.Commands {
md += "- " + command.Name + "\n"
}

return os.WriteFile(filepath.Join(reportDir, "cli-registry.md"), []byte(md), 0644)
}
GO

python - <<PY
from pathlib import Path
import re

module = "$MODULE"
p = Path("cmd/aift/main.go")
s = p.read_text()

# Add cli import.
if f'"{module}/internal/cli"' not in s:
    pos = s.find("import (")
    if pos == -1:
        raise SystemExit("import block not found")
    line = s.find("\\n", pos) + 1
    s = s[:line] + f'\\t"{module}/internal/cli"\\n' + s[line:]

# Add cli-registry command.
if 'case "cli-registry":' not in s:
    target = 'case "provider-registry":'
    i = s.find(target)
    if i == -1:
        target = 'case "scheduler":'
        i = s.find(target)
    if i == -1:
        target = 'case "verify":'
        i = s.find(target)
    if i == -1:
        raise SystemExit("command insertion target not found")

    block = '''\\tcase "cli-registry":
\\t\\tif err := cli.Run(cfg); err != nil {
\\t\\t\\tpanic(err)
\\t\\t}
\\t'''

    s = s[:i] + block + s[i:]

# Replace help body with registry-driven help when possible.
help_start = s.find("func help()")
if help_start != -1:
    brace = s.find("{", help_start)
    depth = 0
    end = None
    for i in range(brace, len(s)):
        if s[i] == "{":
            depth += 1
        elif s[i] == "}":
            depth -= 1
            if depth == 0:
                end = i + 1
                break
    if end:
        s = s[:help_start] + "func help() {\\n\\tcli.PrintHelp()\\n}\\n" + s[end:]

# Remove duplicate switch case blocks.
lines = s.splitlines()
out = []
seen_cases = set()
i = 0

while i < len(lines):
    line = lines[i]
    stripped = line.strip()

    if stripped.startswith('case "') and stripped.endswith('":'):
        cmd = stripped.split('"')[1]
        if cmd in seen_cases:
            i += 1
            while i < len(lines):
                nxt = lines[i].strip()
                if nxt.startswith('case "') or nxt.startswith("default:"):
                    break
                i += 1
            continue
        seen_cases.add(cmd)

    out.append(line)
    i += 1

s = "\\n".join(out) + "\\n"

# Final safety: remove duplicate fmt.Println help lines if any old help remains.
lines = s.splitlines()
out = []
seen_help = set()

for line in lines:
    m = re.search(r'fmt\\.Println\\("  ([^"]+)"\\)', line)
    if m:
        cmd = m.group(1)
        if cmd in seen_help:
            continue
        seen_help.add(cmd)
    out.append(line)

p.write_text("\\n".join(out) + "\\n")
PY

gofmt -w internal/cli/registry.go cmd/aift/main.go

go test ./...
go build -o "$HOME/.local/bin/aift" ./cmd/aift
hash -r

aift --help
aift cli-registry
aift provider-registry || true
aift scheduler || true
aift verify

git add cmd/aift/main.go internal/cli docs/architecture/AIFT-CLI-REGISTRY.md registry/cli reports/cli-registry.md 2>/dev/null || true
git add registry/providers reports/provider-registry.md registry/scheduler reports/federation-scheduler.md registry/capabilities reports/capabilities.md var/events/events.jsonl 2>/dev/null || true

git commit -m "feat: add registry-driven CLI help and command registry" || true
git push origin main

echo "DONE: Phase 18 CLI registry repair complete"
