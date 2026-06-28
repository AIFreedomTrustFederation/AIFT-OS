#!/usr/bin/env bash
set -Eeuo pipefail

cd ~/AIFT/AIFT-OS

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
BIN="bin/aiftd"
REPORT="reports/patch-engine-$STAMP.md"

mkdir -p internal/patchengine docs tests schemas reports AI-Code-Training/scripts/phase-scripts bin

echo "== AIFT Patch Engine =="

TMP_GO="$(mktemp)"

cat > "$TMP_GO" <<'GO'
package patchengine

import (
"encoding/json"
"fmt"
"os"
"os/exec"
"path/filepath"
"sort"
"strings"
"time"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type Operation struct {
ID          string            `json:"id"`
Kind        string            `json:"kind"`
Path        string            `json:"path"`
Status      string            `json:"status"`
Description string            `json:"description"`
Metadata    map[string]string `json:"metadata"`
}

type Plan struct {
SchemaVersion string      `json:"schemaVersion"`
GeneratedAt   string      `json:"generatedAt"`
Root          string      `json:"root"`
Operations    []Operation `json:"operations"`
}

type Result struct {
SchemaVersion string      `json:"schemaVersion"`
GeneratedAt   string      `json:"generatedAt"`
Status        string      `json:"status"`
Root          string      `json:"root"`
Operations    []Operation `json:"operations"`
Checks        []Operation `json:"checks"`
Message       string      `json:"message"`
}

func Inspect(cfg config.Config) error {
plan, err := BuildPlan(cfg)
if err != nil {
return err
}
data, err := json.MarshalIndent(plan, "", "  ")
if err != nil {
return err
}
fmt.Println(string(data))
return nil
}

func PlanCommand(cfg config.Config) error {
plan, err := BuildPlan(cfg)
if err != nil {
return err
}
if err := WritePlan(cfg, plan); err != nil {
return err
}
return WritePlanReport(cfg, plan)
}

func Validate(cfg config.Config) error {
result, err := ValidateTree(cfg)
if err != nil {
_ = WriteResult(cfg, result)
return err
}
if err := WriteResult(cfg, result); err != nil {
return err
}
data, err := json.MarshalIndent(result, "", "  ")
if err != nil {
return err
}
fmt.Println(string(data))
return nil
}

func BuildPlan(cfg config.Config) (Plan, error) {
now := time.Now().UTC().Format(time.RFC3339)
root := cfg.OSHome

ops := []Operation{}

for _, path := range discoverFiles(root, []string{".go", ".sh", ".json", ".md"}) {
kind := classify(path)
ops = append(ops, Operation{
ID:          safeID(kind + "." + rel(root, path)),
Kind:        kind,
Path:        rel(root, path),
Status:      "detected",
Description: "Patchable source artifact discovered.",
Metadata: map[string]string{
"absolutePath": path,
},
})
}

sort.Slice(ops, func(i, j int) bool {
return ops[i].Path < ops[j].Path
})

return Plan{
SchemaVersion: "aift.patch.plan.v1",
GeneratedAt:   now,
Root:          root,
Operations:    ops,
}, nil
}

func ValidateTree(cfg config.Config) (Result, error) {
now := time.Now().UTC().Format(time.RFC3339)
result := Result{
SchemaVersion: "aift.patch.result.v1",
GeneratedAt:   now,
Status:        "ready",
Root:          cfg.OSHome,
Operations:    []Operation{},
Checks:        []Operation{},
Message:       "Patch engine validation passed.",
}

checks := []struct {
id      string
command []string
desc    string
}{
{"gofmt", []string{"gofmt", "-w", "cmd", "internal"}, "Format Go source."},
{"go-test", []string{"go", "test", "./..."}, "Run Go tests."},
{"go-build", []string{"go", "build", "-o", "bin/aiftd", "./cmd/aift"}, "Build AIFT CLI."},
}

for _, check := range checks {
op := Operation{
ID:          check.id,
Kind:        "validation",
Path:        cfg.OSHome,
Status:      "ready",
Description: check.desc,
Metadata: map[string]string{
"command": strings.Join(check.command, " "),
},
}

if output, err := run(cfg.OSHome, check.command...); err != nil {
op.Status = "failed"
op.Metadata["output"] = output
result.Status = "failed"
result.Message = fmt.Sprintf("validation failed at %s", check.id)
result.Checks = append(result.Checks, op)
return result, err
} else {
op.Metadata["output"] = output
}

result.Checks = append(result.Checks, op)
}

return result, nil
}

func WritePlan(cfg config.Config, plan Plan) error {
out := filepath.Join(cfg.OSHome, "registry", "patch-plan.json")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}
data, err := json.MarshalIndent(plan, "", "  ")
if err != nil {
return err
}
return os.WriteFile(out, append(data, '\n'), 0644)
}

func WriteResult(cfg config.Config, result Result) error {
out := filepath.Join(cfg.OSHome, "registry", "patch-result.json")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}
data, err := json.MarshalIndent(result, "", "  ")
if err != nil {
return err
}
return os.WriteFile(out, append(data, '\n'), 0644)
}

func WritePlanReport(cfg config.Config, plan Plan) error {
out := filepath.Join(cfg.OSHome, "reports", "patch-plan.md")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}
var b strings.Builder
b.WriteString("# AIFT Patch Engine Plan\n\n")
b.WriteString("This plan lists patchable source artifacts discovered from the repository. It does not modify files.\n\n")
b.WriteString("| Path | Kind | Status |\n")
b.WriteString("|---|---|---|\n")
for _, op := range plan.Operations {
b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` |\n", op.Path, op.Kind, op.Status))
}
return os.WriteFile(out, []byte(b.String()), 0644)
}

func discoverFiles(root string, suffixes []string) []string {
out := []string{}
_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
if err != nil {
return nil
}
name := d.Name()
if d.IsDir() {
switch name {
case ".git", "node_modules", ".repair-backups", "registry", "reports", "bin":
return filepath.SkipDir
}
return nil
}
for _, suffix := range suffixes {
if strings.HasSuffix(path, suffix) {
out = append(out, path)
break
}
}
return nil
})
return out
}

func classify(path string) string {
switch {
case strings.HasSuffix(path, ".go"):
return "go-source"
case strings.HasSuffix(path, ".sh"):
return "shell-script"
case strings.HasSuffix(path, ".json"):
return "json-document"
case strings.HasSuffix(path, ".md"):
return "markdown-document"
default:
return "file"
}
}

func run(dir string, args ...string) (string, error) {
cmd := exec.Command(args[0], args[1:]...)
cmd.Dir = dir
output, err := cmd.CombinedOutput()
return string(output), err
}

func rel(root, path string) string {
value, err := filepath.Rel(root, path)
if err != nil {
return path
}
return value
}

func safeID(value string) string {
value = strings.ToLower(value)
replacer := strings.NewReplacer("/", ".", " ", "-", "_", "-", ":", "-", "\\", ".")
value = replacer.Replace(value)
return strings.Trim(value, ".-")
}
GO

gofmt -w "$TMP_GO"
cp "$TMP_GO" internal/patchengine/patchengine.go
rm -f "$TMP_GO"

echo "== Patch CLI =="
python3 - <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text()

imp = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/patchengine"'
if imp not in s:
    marker = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/eventbus"'
    if marker not in s:
        marker = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"'
    s = s.replace(marker, marker + "\n\t" + imp, 1)

help_line = 'fmt.Println("  patch-engine inspect|plan|validate")'
if help_line not in s:
    marker = 'fmt.Println("  event-bus publish|list|replay|report")'
    if marker in s:
        s = s.replace(marker, marker + "\n\t" + help_line, 1)

if 'case "patch-engine":' not in s:
    marker = 'case "verify":\n\t\terr = verify(cfg)'
    if marker not in s:
        raise SystemExit("missing verify case marker")
    s = s.replace(marker, 'case "patch-engine":\n\t\terr = runPatchEngine(cfg, args)\n\t' + marker, 1)

if 'func runPatchEngine(' not in s:
    marker = 'func verify(cfg config.Config) error {'
    if marker not in s:
        raise SystemExit("missing verify function marker")
    block = r'''
func runPatchEngine(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "inspect" {
return patchengine.Inspect(cfg)
}

switch args[0] {
case "inspect":
return patchengine.Inspect(cfg)
case "plan":
return patchengine.PlanCommand(cfg)
case "validate":
return patchengine.Validate(cfg)
default:
return fmt.Errorf("usage: aift patch-engine inspect|plan|validate")
}
}

'''
    s = s.replace(marker, block + marker, 1)

p.write_text(s)
PY

cat > tests/patch-engine-smoke.sh <<'SH'
#!/usr/bin/env sh
set -eu

go test ./...
go build -o bin/aiftd ./cmd/aift

bin/aiftd patch-engine inspect >/dev/null
bin/aiftd patch-engine plan >/dev/null
bin/aiftd patch-engine validate >/dev/null

test -f registry/patch-plan.json
test -f registry/patch-result.json
test -f reports/patch-plan.md

echo "OK: patch engine smoke passed"
SH
chmod +x tests/patch-engine-smoke.sh

cat > docs/PATCH-ENGINE.md <<'DOC'
# AIFT Patch Engine

The Patch Engine is the safe mutation layer of AIFT-OS.

It exists because blindly modifying repositories with string replacement does not scale.

## Current foundation

The initial Patch Engine can:

- inspect patchable source artifacts
- generate a patch plan
- run validation commands
- write machine-readable validation results
- write a human-readable patch plan report

## Commands

- `aiftd patch-engine inspect`
- `aiftd patch-engine plan`
- `aiftd patch-engine validate`

## Future direction

The Patch Engine should evolve toward syntax-aware mutation using Go AST, Tree-sitter, LSP, or other parsers where practical.
DOC

cat > schemas/patch-engine.schema.json <<'JSON'
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "AIFT Patch Engine Plan",
  "type": "object",
  "required": ["schemaVersion", "generatedAt", "root", "operations"],
  "properties": {
    "schemaVersion": { "type": "string" },
    "generatedAt": { "type": "string" },
    "root": { "type": "string" },
    "operations": { "type": "array" }
  }
}
JSON

echo "== Verify =="
gofmt -w internal/patchengine/patchengine.go cmd/aift/main.go
go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"
sh tests/patch-engine-smoke.sh

cat > "$REPORT" <<EOF
# Patch Engine Implementation Report

Generated: $STAMP

Passed:

- gofmt
- go test ./...
- go build ./cmd/aift
- tests/patch-engine-smoke.sh

Generated runtime artifacts are intentionally ignored:

- registry/patch-plan.json
- registry/patch-result.json
- reports/patch-plan.md
EOF

cp "$0" AI-Code-Training/scripts/phase-scripts/phase21-patch-engine.sh 2>/dev/null || true

echo "== Stage source files only =="
git add \
  internal/patchengine/patchengine.go \
  cmd/aift/main.go \
  tests/patch-engine-smoke.sh \
  docs/PATCH-ENGINE.md \
  schemas/patch-engine.schema.json \
  AI-Code-Training/scripts/phase-scripts/phase21-patch-engine.sh

echo "== Commit and push =="
if git diff --cached --quiet; then
  echo "Nothing staged."
else
  git commit -m "Implement patch engine foundation"
  git push origin main
fi

echo "DONE"
