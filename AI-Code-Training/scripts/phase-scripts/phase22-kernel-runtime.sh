#!/usr/bin/env bash
set -Eeuo pipefail

cd ~/AIFT/AIFT-OS

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
BIN="bin/aiftd"
REPORT="reports/kernel-runtime-$STAMP.md"

mkdir -p internal/kernelruntime docs tests schemas reports AI-Code-Training/scripts/phase-scripts bin

echo "== AIFT Kernel Runtime =="

TMP_GO="$(mktemp)"

cat > "$TMP_GO" <<'GO'
package kernelruntime

import (
"encoding/json"
"fmt"
"os"
"path/filepath"
"strings"
"time"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/discoveryengine"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/eventbus"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernelregistry"
)

type BootStep struct {
Name        string `json:"name"`
Status      string `json:"status"`
Description string `json:"description"`
StartedAt   string `json:"startedAt"`
FinishedAt  string `json:"finishedAt"`
}

type BootReport struct {
SchemaVersion string     `json:"schemaVersion"`
GeneratedAt   string     `json:"generatedAt"`
Status        string     `json:"status"`
OSHome        string     `json:"osHome"`
Root          string     `json:"root"`
Steps         []BootStep `json:"steps"`
Summary       Summary    `json:"summary"`
}

type Summary struct {
DiscoveryObjects int `json:"discoveryObjects"`
RegistryObjects  int `json:"registryObjects"`
EventCount        int `json:"eventCount"`
}

func Boot(cfg config.Config) error {
report := BootReport{
SchemaVersion: "aift.kernel.boot.v1",
GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
Status:        "booting",
OSHome:        cfg.OSHome,
Root:          cfg.Root,
Steps:         []BootStep{},
}

fmt.Println("AIFT-OS Kernel v0.1")
fmt.Println()

if err := runStep(&report, "configuration", "Configuration loaded from runtime config.", func() error {
return nil
}); err != nil {
return finish(cfg, report, err)
}

if err := runStep(&report, "discovery", "Discovering federation repositories and runtime evidence.", func() error {
return discoveryengine.Scan(cfg)
}); err != nil {
return finish(cfg, report, err)
}

if err := runStep(&report, "kernel-registry", "Building kernel registry from discovered evidence.", func() error {
return kernelregistry.Scan(cfg)
}); err != nil {
return finish(cfg, report, err)
}

if err := runStep(&report, "event-bus", "Publishing kernel boot event.", func() error {
return eventbus.Publish(cfg, "kernel.started", "kernel", "kernelruntime", "AIFT kernel boot sequence completed", map[string]string{
"osHome": cfg.OSHome,
"root":   cfg.Root,
})
}); err != nil {
return finish(cfg, report, err)
}

discovery, _ := discoveryengine.LoadOrBuild(cfg)
registry, _ := kernelregistry.LoadOrBuild(cfg)
events, _ := eventbus.Load(cfg)

report.Summary = Summary{
DiscoveryObjects: len(discovery.Objects),
RegistryObjects:  len(registry.Objects),
EventCount:        len(events),
}
report.Status = "ready"

if err := WriteReport(cfg, report); err != nil {
return err
}
if err := WriteJSON(cfg, report); err != nil {
return err
}

fmt.Println()
fmt.Println("Kernel Ready.")
fmt.Printf("Discovery objects: %d\n", report.Summary.DiscoveryObjects)
fmt.Printf("Registry objects:  %d\n", report.Summary.RegistryObjects)
fmt.Printf("Events:            %d\n", report.Summary.EventCount)

return nil
}

func Status(cfg config.Config) error {
path := filepath.Join(cfg.OSHome, "registry", "kernel-boot.json")
data, err := os.ReadFile(path)
if err != nil {
return fmt.Errorf("kernel has not booted yet: %w", err)
}
fmt.Print(string(data))
return nil
}

func Report(cfg config.Config) error {
path := filepath.Join(cfg.OSHome, "reports", "kernel-boot.md")
data, err := os.ReadFile(path)
if err != nil {
return fmt.Errorf("kernel boot report missing: %w", err)
}
fmt.Print(string(data))
return nil
}

func runStep(report *BootReport, name string, description string, fn func() error) error {
start := time.Now().UTC().Format(time.RFC3339)
step := BootStep{
Name:        name,
Status:      "running",
Description: description,
StartedAt:   start,
}

fmt.Printf("→ %s...\n", name)

if err := fn(); err != nil {
step.Status = "failed"
step.FinishedAt = time.Now().UTC().Format(time.RFC3339)
report.Steps = append(report.Steps, step)
fmt.Printf("✗ %s failed\n", name)
return err
}

step.Status = "ready"
step.FinishedAt = time.Now().UTC().Format(time.RFC3339)
report.Steps = append(report.Steps, step)
fmt.Printf("✓ %s ready\n", name)
return nil
}

func finish(cfg config.Config, report BootReport, err error) error {
report.Status = "failed"
_ = WriteReport(cfg, report)
_ = WriteJSON(cfg, report)
return err
}

func WriteJSON(cfg config.Config, report BootReport) error {
out := filepath.Join(cfg.OSHome, "registry", "kernel-boot.json")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}
data, err := json.MarshalIndent(report, "", "  ")
if err != nil {
return err
}
return os.WriteFile(out, append(data, '\n'), 0644)
}

func WriteReport(cfg config.Config, report BootReport) error {
out := filepath.Join(cfg.OSHome, "reports", "kernel-boot.md")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}

var b strings.Builder
b.WriteString("# AIFT Kernel Boot Report\n\n")
b.WriteString(fmt.Sprintf("- Status: `%s`\n", report.Status))
b.WriteString(fmt.Sprintf("- Generated: `%s`\n", report.GeneratedAt))
b.WriteString(fmt.Sprintf("- OS Home: `%s`\n", report.OSHome))
b.WriteString(fmt.Sprintf("- Root: `%s`\n\n", report.Root))

b.WriteString("## Steps\n\n")
b.WriteString("| Step | Status | Description |\n")
b.WriteString("|---|---|---|\n")
for _, step := range report.Steps {
b.WriteString(fmt.Sprintf("| `%s` | `%s` | %s |\n", step.Name, step.Status, escape(step.Description)))
}

b.WriteString("\n## Summary\n\n")
b.WriteString(fmt.Sprintf("- Discovery objects: `%d`\n", report.Summary.DiscoveryObjects))
b.WriteString(fmt.Sprintf("- Registry objects: `%d`\n", report.Summary.RegistryObjects))
b.WriteString(fmt.Sprintf("- Events: `%d`\n", report.Summary.EventCount))

return os.WriteFile(out, []byte(b.String()), 0644)
}

func escape(value string) string {
return strings.ReplaceAll(value, "|", "\\|")
}
GO

gofmt -w "$TMP_GO"
cp "$TMP_GO" internal/kernelruntime/runtime.go
rm -f "$TMP_GO"

echo "== Patch CLI =="
python3 - <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text()

imp = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernelruntime"'
if imp not in s:
    marker = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/patchengine"'
    if marker not in s:
        marker = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"'
    s = s.replace(marker, marker + "\n\t" + imp, 1)

help_line = 'fmt.Println("  kernel boot|status|report")'
if help_line not in s:
    marker = 'fmt.Println("  patch-engine inspect|plan|validate")'
    if marker in s:
        s = s.replace(marker, marker + "\n\t" + help_line, 1)

if 'case "kernel":' not in s:
    marker = 'case "verify":\n\t\terr = verify(cfg)'
    if marker not in s:
        raise SystemExit("missing verify case marker")
    s = s.replace(marker, 'case "kernel":\n\t\terr = runKernelRuntime(cfg, args)\n\t' + marker, 1)

if 'func runKernelRuntime(' not in s:
    marker = 'func verify(cfg config.Config) error {'
    if marker not in s:
        raise SystemExit("missing verify function marker")
    block = r'''
func runKernelRuntime(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "boot" {
return kernelruntime.Boot(cfg)
}

switch args[0] {
case "boot":
return kernelruntime.Boot(cfg)
case "status":
return kernelruntime.Status(cfg)
case "report":
return kernelruntime.Report(cfg)
default:
return fmt.Errorf("usage: aift kernel boot|status|report")
}
}

'''
    s = s.replace(marker, block + marker, 1)

p.write_text(s)
PY

cat > tests/kernel-runtime-smoke.sh <<'SH'
#!/usr/bin/env sh
set -eu

go test ./...
go build -o bin/aiftd ./cmd/aift

bin/aiftd kernel boot >/dev/null
bin/aiftd kernel status >/dev/null
bin/aiftd kernel report >/dev/null

test -f registry/kernel-boot.json
test -f reports/kernel-boot.md

echo "OK: kernel runtime smoke passed"
SH
chmod +x tests/kernel-runtime-smoke.sh

cat > docs/KERNEL-RUNTIME.md <<'DOC'
# AIFT Kernel Runtime

The Kernel Runtime is the boot orchestrator for AIFT-OS.

It does not replace the operating system kernel of the host machine. It coordinates the federation-level AIFT kernel subsystems.

## Boot sequence

- Load configuration
- Run Discovery Engine
- Build Kernel Registry
- Publish kernel event
- Write boot report

## Commands

- `aiftd kernel boot`
- `aiftd kernel status`
- `aiftd kernel report`

## Runtime artifacts

- `registry/kernel-boot.json`
- `reports/kernel-boot.md`

These are ignored runtime state and should be regenerated from truth.
DOC

cat > schemas/kernel-boot.schema.json <<'JSON'
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "AIFT Kernel Boot Report",
  "type": "object",
  "required": ["schemaVersion", "generatedAt", "status", "steps", "summary"],
  "properties": {
    "schemaVersion": { "type": "string" },
    "generatedAt": { "type": "string" },
    "status": { "type": "string" },
    "steps": { "type": "array" },
    "summary": { "type": "object" }
  }
}
JSON

echo "== Verify =="
gofmt -w internal/kernelruntime/runtime.go cmd/aift/main.go
go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"
sh tests/kernel-runtime-smoke.sh

cat > "$REPORT" <<EOF
# Kernel Runtime Implementation Report

Generated: $STAMP

Passed:

- gofmt
- go test ./...
- go build ./cmd/aift
- tests/kernel-runtime-smoke.sh

Generated runtime artifacts are intentionally ignored:

- registry/kernel-boot.json
- reports/kernel-boot.md
EOF

cp "$0" AI-Code-Training/scripts/phase-scripts/phase22-kernel-runtime.sh 2>/dev/null || true

echo "== Stage source files only =="
git add \
  internal/kernelruntime/runtime.go \
  cmd/aift/main.go \
  tests/kernel-runtime-smoke.sh \
  docs/KERNEL-RUNTIME.md \
  schemas/kernel-boot.schema.json \
  AI-Code-Training/scripts/phase-scripts/phase22-kernel-runtime.sh

echo "== Commit and push =="
if git diff --cached --quiet; then
  echo "Nothing staged."
else
  git commit -m "Implement kernel runtime boot sequence"
  git push origin main
fi

echo "DONE"
