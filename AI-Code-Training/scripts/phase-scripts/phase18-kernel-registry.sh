#!/usr/bin/env bash
set -Eeuo pipefail

cd ~/AIFT/AIFT-OS

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
BIN="bin/aiftd"
BACKUP=".repair-backups/kernel-registry-$STAMP"
REPORT="reports/kernel-registry-$STAMP.md"

mkdir -p "$BACKUP" internal/kernelregistry docs tests schemas reports registry AI-Code-Training/scripts/phase-scripts bin

echo "== Preflight =="
git branch --show-current | grep -qx main

cp -a cmd "$BACKUP/cmd" 2>/dev/null || true
cp -a internal "$BACKUP/internal" 2>/dev/null || true
cp -a docs "$BACKUP/docs" 2>/dev/null || true
cp -a tests "$BACKUP/tests" 2>/dev/null || true
cp -a schemas "$BACKUP/schemas" 2>/dev/null || true

TMP_GO="$(mktemp)"

cat > "$TMP_GO" <<'GO'
package kernelregistry

import (
"encoding/json"
"fmt"
"os"
"path/filepath"
"sort"
"strings"
"time"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Status string

const (
StatusPlanned    Status = "planned"
StatusDetected   Status = "detected"
StatusReady      Status = "ready"
StatusActive     Status = "active"
StatusDeprecated Status = "deprecated"
StatusRemoved    Status = "removed"
)

type Evidence struct {
Kind        string `json:"kind"`
Path        string `json:"path"`
Description string `json:"description"`
ObservedAt  string `json:"observedAt"`
}

type Object struct {
ID            string              `json:"id"`
Kind          string              `json:"kind"`
Name          string              `json:"name"`
Status        Status              `json:"status"`
Location      string              `json:"location"`
Version       string              `json:"version"`
Description   string              `json:"description"`
Evidence      []Evidence          `json:"evidence"`
Provides      []string            `json:"provides"`
Consumes      []string            `json:"consumes"`
DependsOn     []string            `json:"dependsOn"`
Publishes     []string            `json:"publishes"`
Subscribes    []string            `json:"subscribes"`
Commands      map[string]string   `json:"commands"`
Diagnostics   map[string]string   `json:"diagnostics"`
Relationships map[string][]string `json:"relationships"`
GeneratedAt   string              `json:"generatedAt"`
VerifiedAt    string              `json:"verifiedAt,omitempty"`
}

type Registry struct {
SchemaVersion string   `json:"schemaVersion"`
GeneratedAt   string   `json:"generatedAt"`
Source        string   `json:"source"`
Objects       []Object `json:"objects"`
}

func Scan(cfg config.Config) error {
reg, err := Build(cfg)
if err != nil {
return err
}
if err := Write(cfg, reg); err != nil {
return err
}
if err := WriteReport(cfg, reg); err != nil {
return err
}
return events.Emit(cfg, "kernel.registry.scanned", "kernelregistry", "kernel registry scanned", map[string]string{
"objects": fmt.Sprint(len(reg.Objects)),
})
}

func Build(cfg config.Config) (Registry, error) {
now := time.Now().Format(time.RFC3339)

reg := Registry{
SchemaVersion: "aift.kernel.registry.v1",
GeneratedAt:   now,
Source:        "discovered",
Objects:       []Object{},
}

repos, err := workspace.FindRepos(cfg)
if err != nil {
return reg, err
}

reg.Objects = append(reg.Objects, Object{
ID:          "federation.local",
Kind:        "federation",
Name:        "local",
Status:      StatusDetected,
Location:    cfg.Root,
Version:     "0.1.0",
Description: "Local discovered AIFT federation root.",
Evidence: []Evidence{{
Kind:        "directory",
Path:        cfg.Root,
Description: "AIFT root exists on local filesystem.",
ObservedAt:  now,
}},
Commands:      map[string]string{},
Diagnostics:   map[string]string{},
Relationships: map[string][]string{"contains": {}},
GeneratedAt:   now,
})

for _, repo := range repos {
obj := discoverRepository(now, repo.Name, repo.Path)
reg.Objects = append(reg.Objects, obj)
reg.Objects = append(reg.Objects, discoverModuleObjects(now, obj)...)
}

sort.Slice(reg.Objects, func(i, j int) bool {
return reg.Objects[i].ID < reg.Objects[j].ID
})

return reg, nil
}

func discoverRepository(now string, name string, path string) Object {
evidence := []Evidence{{
Kind:        "git",
Path:        filepath.Join(path, ".git"),
Description: "Repository has a .git directory.",
ObservedAt:  now,
}}

commands := map[string]string{}
provides := []string{"repository"}
diagnostics := map[string]string{
"git_status": "git status --short",
}

if exists(filepath.Join(path, "README.md")) {
evidence = append(evidence, ev(now, "doc", filepath.Join(path, "README.md"), "README documentation exists."))
provides = append(provides, "documentation.seed")
}
if exists(filepath.Join(path, "package.json")) {
evidence = append(evidence, ev(now, "manifest", filepath.Join(path, "package.json"), "Node package manifest exists."))
provides = append(provides, "node.package")
readPackageCommands(path, commands)
}
if exists(filepath.Join(path, "go.mod")) {
evidence = append(evidence, ev(now, "manifest", filepath.Join(path, "go.mod"), "Go module manifest exists."))
provides = append(provides, "go.module")
commands["go:test"] = "go test ./..."
commands["go:build"] = "go build ./..."
}
if exists(filepath.Join(path, "Cargo.toml")) {
evidence = append(evidence, ev(now, "manifest", filepath.Join(path, "Cargo.toml"), "Rust Cargo manifest exists."))
provides = append(provides, "rust.crate")
commands["cargo:test"] = "cargo test"
commands["cargo:build"] = "cargo build"
}
if exists(filepath.Join(path, ".github", "workflows")) {
evidence = append(evidence, ev(now, "workflow", filepath.Join(path, ".github", "workflows"), "GitHub workflow directory exists."))
provides = append(provides, "github.workflows")
}
if exists(filepath.Join(path, ".aift", "module.json")) {
evidence = append(evidence, ev(now, "contract", filepath.Join(path, ".aift", "module.json"), "AIFT module manifest exists."))
provides = append(provides, "aift.module.contract")
}

status := StatusDetected
if len(commands) > 0 {
status = StatusReady
}

return Object{
ID:            "repository." + safeID(name),
Kind:          "repository",
Name:          name,
Status:        status,
Location:      path,
Version:       "0.1.0",
Description:   "Discovered repository object for " + name + ".",
Evidence:      evidence,
Provides:      unique(provides),
Consumes:      []string{},
DependsOn:     []string{},
Publishes:     []string{"repository.discovered"},
Subscribes:    []string{},
Commands:      commands,
Diagnostics:   diagnostics,
Relationships: map[string][]string{"containedBy": {"federation.local"}},
GeneratedAt:   now,
}
}

func discoverModuleObjects(now string, repo Object) []Object {
out := []Object{}

if hasProvide(repo, "aift.module.contract") {
out = append(out, Object{
ID:          "module." + safeID(repo.Name),
Kind:        "module",
Name:        repo.Name,
Status:      repo.Status,
Location:    filepath.Join(repo.Location, ".aift", "module.json"),
Version:     "0.1.0",
Description: "AIFT module object discovered from repository contract.",
Evidence: []Evidence{{
Kind:        "contract",
Path:        filepath.Join(repo.Location, ".aift", "module.json"),
Description: "Module manifest exists.",
ObservedAt:  now,
}},
Provides:      []string{"aift.module"},
Consumes:      []string{},
DependsOn:     []string{repo.ID},
Publishes:     []string{"module.discovered"},
Subscribes:    []string{},
Commands:      map[string]string{},
Diagnostics:   map[string]string{},
Relationships: map[string][]string{"implementedBy": {repo.ID}},
GeneratedAt:   now,
})
}

for _, capability := range repo.Provides {
out = append(out, Object{
ID:          "capability." + safeID(repo.Name) + "." + safeID(capability),
Kind:        "capability",
Name:        capability,
Status:      repo.Status,
Location:    repo.Location,
Version:     "0.1.0",
Description: "Capability discovered from repository evidence.",
Evidence:    repo.Evidence,
Provides:    []string{capability},
Consumes:    []string{},
DependsOn:   []string{repo.ID},
Publishes:   []string{"capability.detected"},
Subscribes:  []string{},
Commands:    map[string]string{},
Diagnostics: map[string]string{},
Relationships: map[string][]string{
"providedBy": {repo.ID},
},
GeneratedAt: now,
})
}

return out
}

func List(cfg config.Config) error {
reg, err := LoadOrBuild(cfg)
if err != nil {
return err
}
fmt.Printf("%-56s %-14s %-12s %s\n", "OBJECT", "KIND", "STATUS", "NAME")
for _, obj := range reg.Objects {
fmt.Printf("%-56s %-14s %-12s %s\n", obj.ID, obj.Kind, obj.Status, obj.Name)
}
return nil
}

func ObjectInfo(cfg config.Config, id string) error {
reg, err := LoadOrBuild(cfg)
if err != nil {
return err
}
for _, obj := range reg.Objects {
if obj.ID == id || obj.Name == id {
data, err := json.MarshalIndent(obj, "", "  ")
if err != nil {
return err
}
fmt.Println(string(data))
return nil
}
}
return fmt.Errorf("kernel registry object not found: %s", id)
}

func Report(cfg config.Config) error {
reg, err := LoadOrBuild(cfg)
if err != nil {
return err
}
return WriteReport(cfg, reg)
}

func Write(cfg config.Config, reg Registry) error {
out := filepath.Join(cfg.OSHome, "registry", "kernel-registry.json")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}
data, err := json.MarshalIndent(reg, "", "  ")
if err != nil {
return err
}
return os.WriteFile(out, append(data, '\n'), 0644)
}

func WriteReport(cfg config.Config, reg Registry) error {
out := filepath.Join(cfg.OSHome, "reports", "kernel-registry.md")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}
var b strings.Builder
b.WriteString("# AIFT Kernel Registry\n\n")
b.WriteString("This report is generated from discovered evidence. It is not a claim that every object is active.\n\n")
b.WriteString("| Object | Kind | Status | Evidence | Provides |\n")
b.WriteString("|---|---|---|---:|---|\n")
for _, obj := range reg.Objects {
b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%d` | `%s` |\n",
obj.ID,
obj.Kind,
obj.Status,
len(obj.Evidence),
strings.Join(obj.Provides, ", "),
))
}
return os.WriteFile(out, []byte(b.String()), 0644)
}

func LoadOrBuild(cfg config.Config) (Registry, error) {
path := filepath.Join(cfg.OSHome, "registry", "kernel-registry.json")
data, err := os.ReadFile(path)
if err != nil {
return Build(cfg)
}
var reg Registry
if err := json.Unmarshal(data, &reg); err != nil {
return Registry{}, err
}
return reg, nil
}

func readPackageCommands(repoPath string, commands map[string]string) {
data, err := os.ReadFile(filepath.Join(repoPath, "package.json"))
if err != nil {
return
}
var pkg struct {
Scripts map[string]string `json:"scripts"`
}
if json.Unmarshal(data, &pkg) != nil {
return
}
for name := range pkg.Scripts {
commands["npm:"+name] = "npm run " + name
}
}

func ev(now string, kind string, path string, description string) Evidence {
return Evidence{
Kind:        kind,
Path:        path,
Description: description,
ObservedAt:  now,
}
}

func exists(path string) bool {
_, err := os.Stat(path)
return err == nil
}

func hasProvide(obj Object, capability string) bool {
for _, item := range obj.Provides {
if item == capability {
return true
}
}
return false
}

func safeID(value string) string {
value = strings.ToLower(value)
value = strings.ReplaceAll(value, "/", ".")
value = strings.ReplaceAll(value, " ", "-")
value = strings.ReplaceAll(value, "_", "-")
return value
}

func unique(items []string) []string {
seen := map[string]bool{}
out := []string{}
for _, item := range items {
if item == "" || seen[item] {
continue
}
seen[item] = true
out = append(out, item)
}
sort.Strings(out)
return out
}
GO

echo "== Validate Go before installing =="
gofmt -w "$TMP_GO"
mkdir -p internal/kernelregistry
cp "$TMP_GO" internal/kernelregistry/registry.go
rm -f "$TMP_GO"

echo "== Patch CLI =="
python3 - <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text()

imp = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernelregistry"'
if imp not in s:
    marker = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernel"'
    if marker not in s:
        marker = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"'
    s = s.replace(marker, marker + "\n\t" + imp, 1)

help_line = 'fmt.Println("  kernel-registry scan|list|object|report")'
if help_line not in s:
    marker = 'fmt.Println("  modules init-all|scan|list|repo|report")'
    if marker in s:
        s = s.replace(marker, marker + "\n\t" + help_line, 1)

if 'case "kernel-registry":' not in s:
    marker = 'case "verify":\n\t\terr = verify(cfg)'
    if marker not in s:
        raise SystemExit("missing verify case marker")
    s = s.replace(marker, 'case "kernel-registry":\n\t\terr = runKernelRegistry(cfg, args)\n\t' + marker, 1)

if 'func runKernelRegistry(' not in s:
    marker = 'func verify(cfg config.Config) error {'
    if marker not in s:
        raise SystemExit("missing verify function marker")
    block = '''
func runKernelRegistry(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
return kernelregistry.Scan(cfg)
}

switch args[0] {
case "list":
return kernelregistry.List(cfg)
case "object":
if len(args) < 2 {
return fmt.Errorf("usage: aift kernel-registry object <id-or-name>")
}
return kernelregistry.ObjectInfo(cfg, args[1])
case "report":
return kernelregistry.Report(cfg)
default:
return fmt.Errorf("usage: aift kernel-registry scan|list|object|report")
}
}

'''
    s = s.replace(marker, block + marker, 1)

p.write_text(s)
PY

cat > tests/kernel-registry-smoke.sh <<'SH'
#!/usr/bin/env sh
set -eu

go test ./...
go build -o bin/aiftd ./cmd/aift

bin/aiftd kernel-registry scan >/dev/null
bin/aiftd kernel-registry list >/dev/null
bin/aiftd kernel-registry object federation.local >/dev/null
bin/aiftd kernel-registry report >/dev/null

test -f registry/kernel-registry.json
test -f reports/kernel-registry.md

echo "OK: kernel registry smoke passed"
SH
chmod +x tests/kernel-registry-smoke.sh

cat > docs/KERNEL-REGISTRY.md <<'DOC'
# AIFT Kernel Registry

The Kernel Registry is the authoritative runtime inventory of discovered federation reality.

It records objects such as:

- federations
- repositories
- modules
- capabilities
- services
- commands
- diagnostics
- relationships

Nothing enters the registry without evidence.

No object is considered active unless validation proves it.

## Commands

- `aiftd kernel-registry scan`
- `aiftd kernel-registry list`
- `aiftd kernel-registry object <id-or-name>`
- `aiftd kernel-registry report`

## Generated files

- `registry/kernel-registry.json`
- `reports/kernel-registry.md`
DOC

cat > schemas/kernel-registry.schema.json <<'JSON'
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "AIFT Kernel Registry",
  "type": "object",
  "required": ["schemaVersion", "generatedAt", "source", "objects"],
  "properties": {
    "schemaVersion": { "type": "string" },
    "generatedAt": { "type": "string" },
    "source": { "type": "string" },
    "objects": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["id", "kind", "name", "status", "evidence", "generatedAt"],
        "properties": {
          "id": { "type": "string" },
          "kind": { "type": "string" },
          "name": { "type": "string" },
          "status": {
            "type": "string",
            "enum": ["planned", "detected", "ready", "active", "deprecated", "removed"]
          },
          "evidence": { "type": "array" },
          "provides": { "type": "array" },
          "consumes": { "type": "array" },
          "dependsOn": { "type": "array" },
          "commands": { "type": "object" }
        }
      }
    }
  }
}
JSON

echo "== Verify =="
gofmt -w internal/kernelregistry/registry.go cmd/aift/main.go
go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"
sh tests/kernel-registry-smoke.sh

cat > "$REPORT" <<EOF
# Kernel Registry Implementation Report

Generated: $STAMP

Passed:

- gofmt
- go test ./...
- go build ./cmd/aift
- tests/kernel-registry-smoke.sh

This implements the first foundation of the AIFT-OS kernel registry.

EOF

cp "$0" AI-Code-Training/scripts/phase-scripts/phase18-kernel-registry.sh 2>/dev/null || true

echo "== Verify generated runtime artifacts exist =="
test -f registry/kernel-registry.json
test -f reports/kernel-registry.md

echo "== Stage source files =="
git add \
  internal/kernelregistry/registry.go \
  cmd/aift/main.go \
  tests/kernel-registry-smoke.sh \
  docs/KERNEL-REGISTRY.md \
  schemas/kernel-registry.schema.json \
  AI-Code-Training/scripts/phase-scripts/phase18-kernel-registry.sh \

echo "== Commit and push =="
if git diff --cached --quiet; then
  echo "Nothing staged."
else
  git commit -m "Implement kernel registry foundation"
  git push origin main
fi

echo "DONE"
