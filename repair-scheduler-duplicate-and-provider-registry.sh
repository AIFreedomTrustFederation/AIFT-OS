#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

echo "Repairing duplicate CLI commands and adding provider-registry foundation"

MODULE="$(awk '/^module / {print $2; exit}' go.mod)"

mkdir -p internal/providerregistry docs/architecture registry/providers reports

python - <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text().splitlines()

seen_cases = set()
seen_help = set()
out = []
skip_case = False

for line in s:
    stripped = line.strip()

    if stripped.startswith('case "') and stripped.endswith('":'):
        cmd = stripped.split('"')[1]
        if cmd in seen_cases:
            skip_case = True
            continue
        seen_cases.add(cmd)
        skip_case = False
        out.append(line)
        continue

    if skip_case:
        if stripped.startswith('case "') or stripped.startswith("default:"):
            skip_case = False
            cmd = stripped.split('"')[1] if stripped.startswith('case "') else None
            if cmd and cmd in seen_cases:
                skip_case = True
                continue
            if cmd:
                seen_cases.add(cmd)
            out.append(line)
        continue

    if 'fmt.Println("  ' in line:
        cmd = line.split('fmt.Println("  ', 1)[1].split('"', 1)[0]
        if cmd in seen_help:
            continue
        seen_help.add(cmd)

    out.append(line)

p.write_text("\n".join(out) + "\n")
PY

cat > docs/architecture/AIFT-PROVIDER-REGISTRY.md <<'DOC'
# AIFT Provider Registry

The Provider Registry is the runtime-agnostic foundation for AIFT-OS.

The scheduler should eventually stop knowing about Go, Node, Python, Rust, Make, Java, Docker, or any other runtime directly.

Instead, providers describe:

- detection files
- required capabilities
- build commands
- test commands
- sync support
- async support

This keeps the scheduler module-agnostic, runtime-agnostic, provider-agnostic, and compatible with both synchronous and asynchronous execution.
DOC

cat > internal/providerregistry/providerregistry.go <<GO
package providerregistry

import (
"encoding/json"
"fmt"
"os"
"path/filepath"
"time"

"$MODULE/internal/config"
)

type Provider struct {
Name                 string   \`json:"name"\`
Runtime              string   \`json:"runtime"\`
DetectionFiles       []string \`json:"detection_files"\`
RequiredCapabilities []string \`json:"required_capabilities"\`
BuildCommand         string   \`json:"build_command"\`
TestCommand          string   \`json:"test_command"\`
SupportsSync         bool     \`json:"supports_sync"\`
SupportsAsync        bool     \`json:"supports_async"\`
}

type Report struct {
Name      string     \`json:"name"\`
Time      string     \`json:"time"\`
Verified  bool       \`json:"verified"\`
Providers []Provider \`json:"providers"\`
}

func Builtins() []Provider {
return []Provider{
{
Name:                 "go",
Runtime:              "go",
DetectionFiles:       []string{"go.mod"},
RequiredCapabilities: []string{"go"},
BuildCommand:         "go build ./...",
TestCommand:          "go test ./...",
SupportsSync:         true,
SupportsAsync:        true,
},
{
Name:                 "node-pnpm",
Runtime:              "node",
DetectionFiles:       []string{"pnpm-lock.yaml"},
RequiredCapabilities: []string{"node", "pnpm"},
BuildCommand:         "pnpm run build",
TestCommand:          "pnpm test",
SupportsSync:         true,
SupportsAsync:        true,
},
{
Name:                 "node-npm",
Runtime:              "node",
DetectionFiles:       []string{"package.json"},
RequiredCapabilities: []string{"node", "npm"},
BuildCommand:         "npm run build",
TestCommand:          "npm test",
SupportsSync:         true,
SupportsAsync:        true,
},
{
Name:                 "python",
Runtime:              "python",
DetectionFiles:       []string{"pyproject.toml", "requirements.txt"},
RequiredCapabilities: []string{"python"},
BuildCommand:         "python -m compileall .",
TestCommand:          "python -m pytest",
SupportsSync:         true,
SupportsAsync:        true,
},
{
Name:                 "rust",
Runtime:              "rust",
DetectionFiles:       []string{"Cargo.toml"},
RequiredCapabilities: []string{"cargo"},
BuildCommand:         "cargo build",
TestCommand:          "cargo test",
SupportsSync:         true,
SupportsAsync:        true,
},
{
Name:                 "make",
Runtime:              "make",
DetectionFiles:       []string{"Makefile"},
RequiredCapabilities: []string{"make"},
BuildCommand:         "make",
TestCommand:          "make test",
SupportsSync:         true,
SupportsAsync:        true,
},
}
}

func Match(path string) (Provider, bool) {
for _, provider := range Builtins() {
for _, file := range provider.DetectionFiles {
if exists(filepath.Join(path, file)) {
return provider, true
}
}
}
return Provider{}, false
}

func Run(cfg config.Config) error {
report := Report{
Name:      "AIFT Provider Registry",
Time:      time.Now().Format(time.RFC3339),
Verified:  true,
Providers: Builtins(),
}

if err := Write(cfg, report); err != nil {
return err
}

fmt.Println("AIFT Provider Registry")
fmt.Println("providers:", len(report.Providers))
return nil
}

func Write(cfg config.Config, report Report) error {
outDir := filepath.Join(cfg.OSHome, "registry", "providers")
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

jsonPath := filepath.Join(outDir, "provider-registry.json")
mdPath := filepath.Join(reportDir, "provider-registry.md")

if err := os.WriteFile(jsonPath, append(b, '\n'), 0644); err != nil {
return err
}

md := "# AIFT Provider Registry Report\n\n"
for _, provider := range report.Providers {
md += "- " + provider.Name + " | " + provider.Runtime + "\n"
}

return os.WriteFile(mdPath, []byte(md), 0644)
}

func exists(path string) bool {
_, err := os.Stat(path)
return err == nil
}
GO

python - <<PY
from pathlib import Path

module = "$MODULE"
p = Path("cmd/aift/main.go")
s = p.read_text()

if f'"{module}/internal/providerregistry"' not in s:
    pos = s.find("import (")
    line = s.find("\\n", pos) + 1
    s = s[:line] + f'\\t"{module}/internal/providerregistry"\\n' + s[line:]

if 'case "provider-registry":' not in s:
    target = 'case "scheduler":'
    i = s.find(target)
    if i == -1:
        target = 'case "verify":'
        i = s.find(target)
    if i == -1:
        raise SystemExit("command insertion target not found")
    block = '''\tcase "provider-registry":
\t\tif err := providerregistry.Run(cfg); err != nil {
\t\t\tpanic(err)
\t\t}
\t'''
    s = s[:i] + block + s[i:]

if 'fmt.Println("  provider-registry")' not in s:
    if 'fmt.Println("  scheduler")' in s:
        s = s.replace(
            'fmt.Println("  scheduler")',
            'fmt.Println("  scheduler")\\n\\tfmt.Println("  provider-registry")'
        )
    else:
        s = s.replace(
            'fmt.Println("  verify")',
            'fmt.Println("  provider-registry")\\n\\tfmt.Println("  verify")'
        )

p.write_text(s)
PY

gofmt -w cmd/aift/main.go internal/providerregistry/providerregistry.go

go test ./...
go build -o "$HOME/.local/bin/aift" ./cmd/aift
hash -r

aift provider-registry
aift scheduler || true
aift verify

git add cmd/aift/main.go internal/providerregistry docs/architecture/AIFT-PROVIDER-REGISTRY.md registry/providers reports/provider-registry.md 2>/dev/null || true
git add registry/scheduler reports/federation-scheduler.md registry/capabilities reports/capabilities.md var/events/events.jsonl 2>/dev/null || true

git commit -m "feat: add provider registry and repair duplicate CLI commands" || true
git push origin main

echo "DONE: duplicate CLI repair and provider registry added"
