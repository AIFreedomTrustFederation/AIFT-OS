#!/usr/bin/env bash
set -Eeuo pipefail

cd ~/AIFT/AIFT-OS

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
BIN="bin/aiftd"
REPORT="reports/discovery-engine-$STAMP.md"

mkdir -p internal/discoveryengine docs tests schemas registry reports AI-Code-Training/scripts/phase-scripts bin

echo "== AIFT Discovery Engine =="

echo "== Write discovery engine =="
TMP_GO="$(mktemp)"

cat > "$TMP_GO" <<'GO'
package discoveryengine

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

type Evidence struct {
Kind        string `json:"kind"`
Path        string `json:"path"`
Description string `json:"description"`
ObservedAt  string `json:"observedAt"`
}

type Runtime struct {
Name     string     `json:"name"`
Kind     string     `json:"kind"`
Version  string     `json:"version,omitempty"`
Evidence []Evidence `json:"evidence"`
}

type DiscoveryObject struct {
ID           string            `json:"id"`
Kind         string            `json:"kind"`
Name         string            `json:"name"`
Status       string            `json:"status"`
Location     string            `json:"location"`
Description  string            `json:"description"`
Evidence     []Evidence        `json:"evidence"`
Runtimes     []Runtime         `json:"runtimes"`
Manifests    []string          `json:"manifests"`
Docs         []string          `json:"docs"`
Schemas      []string          `json:"schemas"`
Workflows    []string          `json:"workflows"`
Commands     map[string]string `json:"commands"`
Capabilities []string          `json:"capabilities"`
Services     []string          `json:"services"`
HealthChecks []string          `json:"healthChecks"`
GeneratedAt  string            `json:"generatedAt"`
}

type Snapshot struct {
SchemaVersion string            `json:"schemaVersion"`
GeneratedAt   string            `json:"generatedAt"`
Source        string            `json:"source"`
Objects       []DiscoveryObject `json:"objects"`
}

func Scan(cfg config.Config) error {
snap, err := Build(cfg)
if err != nil {
return err
}
if err := Write(cfg, snap); err != nil {
return err
}
if err := WriteReport(cfg, snap); err != nil {
return err
}
return events.Emit(cfg, "discovery.scan.completed", "discoveryengine", "discovery scan completed", map[string]string{
"objects": fmt.Sprint(len(snap.Objects)),
})
}

func Build(cfg config.Config) (Snapshot, error) {
now := time.Now().Format(time.RFC3339)

snap := Snapshot{
SchemaVersion: "aift.discovery.v1",
GeneratedAt:   now,
Source:        "filesystem.git.manifests",
Objects:       []DiscoveryObject{},
}

repos, err := workspace.FindRepos(cfg)
if err != nil {
return snap, err
}

for _, repo := range repos {
snap.Objects = append(snap.Objects, DiscoverRepository(now, repo.Name, repo.Path))
}

sort.Slice(snap.Objects, func(i, j int) bool {
return snap.Objects[i].ID < snap.Objects[j].ID
})

return snap, nil
}

func DiscoverRepository(now string, name string, path string) DiscoveryObject {
obj := DiscoveryObject{
ID:           "repository." + safeID(name),
Kind:         "repository",
Name:         name,
Status:       "detected",
Location:     path,
Description:  "Repository discovered from filesystem and Git evidence.",
Evidence:     []Evidence{},
Runtimes:     []Runtime{},
Manifests:    []string{},
Docs:         []string{},
Schemas:      []string{},
Workflows:    []string{},
Commands:     map[string]string{},
Capabilities: []string{},
Services:     []string{},
HealthChecks: []string{},
GeneratedAt:  now,
}

addEvidence(&obj, now, "git", filepath.Join(path, ".git"), "Git repository exists.")

discoverDocs(&obj, now, path)
discoverSchemas(&obj, now, path)
discoverWorkflows(&obj, now, path)
discoverManifests(&obj, now, path)
discoverRuntimes(&obj, now, path)
discoverAIFTContracts(&obj, now, path)

if len(obj.Commands) > 0 || len(obj.Capabilities) > 0 || len(obj.Services) > 0 {
obj.Status = "ready"
}

obj.Docs = unique(obj.Docs)
obj.Schemas = unique(obj.Schemas)
obj.Workflows = unique(obj.Workflows)
obj.Manifests = unique(obj.Manifests)
obj.Capabilities = unique(obj.Capabilities)
obj.Services = unique(obj.Services)
obj.HealthChecks = unique(obj.HealthChecks)

return obj
}

func discoverDocs(obj *DiscoveryObject, now string, root string) {
candidates := []string{"README.md", "AGENTS.md", "docs", "manual", "book", "site"}
for _, rel := range candidates {
path := filepath.Join(root, rel)
if exists(path) {
obj.Docs = append(obj.Docs, rel)
addEvidence(obj, now, "documentation", path, "Documentation path exists.")
}
}
}

func discoverSchemas(obj *DiscoveryObject, now string, root string) {
candidates := []string{"schemas", "schema", ".aift/schemas"}
for _, rel := range candidates {
path := filepath.Join(root, rel)
if exists(path) {
obj.Schemas = append(obj.Schemas, rel)
addEvidence(obj, now, "schema", path, "Schema path exists.")
}
}
}

func discoverWorkflows(obj *DiscoveryObject, now string, root string) {
candidates := []string{".github/workflows", "workflows", ".aift/workflows"}
for _, rel := range candidates {
path := filepath.Join(root, rel)
if exists(path) {
obj.Workflows = append(obj.Workflows, rel)
addEvidence(obj, now, "workflow", path, "Workflow path exists.")
}
}
}

func discoverManifests(obj *DiscoveryObject, now string, root string) {
manifestFiles := []string{
"package.json",
"go.mod",
"Cargo.toml",
"pyproject.toml",
"requirements.txt",
"deno.json",
"bun.lockb",
"pnpm-lock.yaml",
"yarn.lock",
"package-lock.json",
"Dockerfile",
"docker-compose.yml",
"compose.yml",
".aift/module.json",
".aift/capabilities.json",
".aift/services.json",
".aift/events.json",
".aift/manual.json",
}
for _, rel := range manifestFiles {
path := filepath.Join(root, rel)
if exists(path) {
obj.Manifests = append(obj.Manifests, rel)
addEvidence(obj, now, "manifest", path, "Manifest file exists.")
}
}
}

func discoverRuntimes(obj *DiscoveryObject, now string, root string) {
if exists(filepath.Join(root, "package.json")) {
obj.Runtimes = append(obj.Runtimes, Runtime{Name: "node", Kind: "javascript", Evidence: []Evidence{ev(now, "manifest", filepath.Join(root, "package.json"), "package.json exists.")}})
obj.Capabilities = append(obj.Capabilities, "node.package")
readPackageCommands(root, obj.Commands)
}
if exists(filepath.Join(root, "go.mod")) {
obj.Runtimes = append(obj.Runtimes, Runtime{Name: "go", Kind: "go", Evidence: []Evidence{ev(now, "manifest", filepath.Join(root, "go.mod"), "go.mod exists.")}})
obj.Capabilities = append(obj.Capabilities, "go.module")
obj.Commands["go:test"] = "go test ./..."
obj.Commands["go:build"] = "go build ./..."
}
if exists(filepath.Join(root, "Cargo.toml")) {
obj.Runtimes = append(obj.Runtimes, Runtime{Name: "cargo", Kind: "rust", Evidence: []Evidence{ev(now, "manifest", filepath.Join(root, "Cargo.toml"), "Cargo.toml exists.")}})
obj.Capabilities = append(obj.Capabilities, "rust.crate")
obj.Commands["cargo:test"] = "cargo test"
obj.Commands["cargo:build"] = "cargo build"
}
if exists(filepath.Join(root, "pyproject.toml")) || exists(filepath.Join(root, "requirements.txt")) {
obj.Runtimes = append(obj.Runtimes, Runtime{Name: "python", Kind: "python", Evidence: []Evidence{ev(now, "manifest", root, "Python manifest evidence exists.")}})
obj.Capabilities = append(obj.Capabilities, "python.project")
}
if exists(filepath.Join(root, "Dockerfile")) {
obj.Runtimes = append(obj.Runtimes, Runtime{Name: "docker", Kind: "container", Evidence: []Evidence{ev(now, "manifest", filepath.Join(root, "Dockerfile"), "Dockerfile exists.")}})
obj.Capabilities = append(obj.Capabilities, "container.image")
}
}

func discoverAIFTContracts(obj *DiscoveryObject, now string, root string) {
if exists(filepath.Join(root, ".aift", "module.json")) {
obj.Capabilities = append(obj.Capabilities, "aift.module")
}
if exists(filepath.Join(root, ".aift", "capabilities.json")) {
obj.Capabilities = append(obj.Capabilities, readNamedList(root, "capabilities.json", "capabilities")...)
}
if exists(filepath.Join(root, ".aift", "services.json")) {
obj.Services = append(obj.Services, readNamedList(root, "services.json", "services")...)
}
if exists(filepath.Join(root, ".aift", "commands", "verify.sh")) {
obj.Commands["aift:verify"] = "sh .aift/commands/verify.sh"
obj.HealthChecks = append(obj.HealthChecks, ".aift/commands/verify.sh")
}
}

func List(cfg config.Config) error {
snap, err := LoadOrBuild(cfg)
if err != nil {
return err
}
fmt.Printf("%-40s %-12s %-12s %-24s %s\n", "OBJECT", "KIND", "STATUS", "RUNTIMES", "NAME")
for _, obj := range snap.Objects {
fmt.Printf("%-40s %-12s %-12s %-24s %s\n", obj.ID, obj.Kind, obj.Status, runtimeNames(obj.Runtimes), obj.Name)
}
return nil
}

func ObjectInfo(cfg config.Config, id string) error {
snap, err := LoadOrBuild(cfg)
if err != nil {
return err
}
for _, obj := range snap.Objects {
if obj.ID == id || obj.Name == id {
data, err := json.MarshalIndent(obj, "", "  ")
if err != nil {
return err
}
fmt.Println(string(data))
return nil
}
}
return fmt.Errorf("discovery object not found: %s", id)
}

func Report(cfg config.Config) error {
snap, err := LoadOrBuild(cfg)
if err != nil {
return err
}
if err := WriteReport(cfg, snap); err != nil {
return err
}
data, err := os.ReadFile(filepath.Join(cfg.OSHome, "reports", "discovery.md"))
if err != nil {
return err
}
fmt.Print(string(data))
return nil
}

func Write(cfg config.Config, snap Snapshot) error {
out := filepath.Join(cfg.OSHome, "registry", "discovery.json")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}
data, err := json.MarshalIndent(snap, "", "  ")
if err != nil {
return err
}
return os.WriteFile(out, append(data, '\n'), 0644)
}

func WriteReport(cfg config.Config, snap Snapshot) error {
out := filepath.Join(cfg.OSHome, "reports", "discovery.md")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}

var b strings.Builder
b.WriteString("# AIFT Discovery Report\n\n")
b.WriteString("Generated from filesystem, Git, manifest, runtime, documentation, workflow, schema, and AIFT contract evidence.\n\n")
b.WriteString("| Object | Status | Runtimes | Manifests | Commands | Capabilities | Services |\n")
b.WriteString("|---|---|---|---:|---:|---:|---:|\n")

for _, obj := range snap.Objects {
b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%d` | `%d` | `%d` | `%d` |\n",
obj.ID,
obj.Status,
runtimeNames(obj.Runtimes),
len(obj.Manifests),
len(obj.Commands),
len(obj.Capabilities),
len(obj.Services),
))
}

return os.WriteFile(out, []byte(b.String()), 0644)
}

func LoadOrBuild(cfg config.Config) (Snapshot, error) {
path := filepath.Join(cfg.OSHome, "registry", "discovery.json")
data, err := os.ReadFile(path)
if err != nil {
return Build(cfg)
}
var snap Snapshot
if err := json.Unmarshal(data, &snap); err != nil {
return Snapshot{}, err
}
return snap, nil
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

func readNamedList(repoPath string, fileName string, field string) []string {
data, err := os.ReadFile(filepath.Join(repoPath, ".aift", fileName))
if err != nil {
return []string{}
}
var raw map[string][]map[string]string
if json.Unmarshal(data, &raw) != nil {
return []string{}
}
out := []string{}
for _, item := range raw[field] {
if item["name"] != "" {
out = append(out, item["name"])
}
}
return out
}

func addEvidence(obj *DiscoveryObject, now string, kind string, path string, description string) {
obj.Evidence = append(obj.Evidence, ev(now, kind, path, description))
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

func runtimeNames(runtimes []Runtime) string {
names := []string{}
for _, rt := range runtimes {
names = append(names, rt.Name)
}
return strings.Join(unique(names), ",")
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

echo "== Validate generated Go before installing =="
gofmt -w "$TMP_GO"
cp "$TMP_GO" internal/discoveryengine/discovery.go
rm -f "$TMP_GO"

echo "== Patch CLI =="
python3 - <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text()

imp = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/discoveryengine"'
if imp not in s:
    marker = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernelregistry"'
    if marker not in s:
        marker = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"'
    s = s.replace(marker, marker + "\n\t" + imp, 1)

help_line = 'fmt.Println("  discovery scan|list|object|report")'
if help_line not in s:
    marker = 'fmt.Println("  kernel-registry scan|list|object|report")'
    if marker in s:
        s = s.replace(marker, marker + "\n\t" + help_line, 1)

if 'case "discovery":' not in s:
    marker = 'case "verify":\n\t\terr = verify(cfg)'
    if marker not in s:
        raise SystemExit("missing verify case marker")
    s = s.replace(marker, 'case "discovery":\n\t\terr = runDiscovery(cfg, args)\n\t' + marker, 1)

if 'func runDiscovery(' not in s:
    marker = 'func verify(cfg config.Config) error {'
    if marker not in s:
        raise SystemExit("missing verify function marker")
    block = '''
func runDiscovery(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
return discoveryengine.Scan(cfg)
}

switch args[0] {
case "list":
return discoveryengine.List(cfg)
case "object":
if len(args) < 2 {
return fmt.Errorf("usage: aift discovery object <id-or-name>")
}
return discoveryengine.ObjectInfo(cfg, args[1])
case "report":
return discoveryengine.Report(cfg)
default:
return fmt.Errorf("usage: aift discovery scan|list|object|report")
}
}

'''
    s = s.replace(marker, block + marker, 1)

p.write_text(s)
PY

echo "== Write tests docs schemas =="
cat > tests/discovery-smoke.sh <<'SH'
#!/usr/bin/env sh
set -eu

go test ./...
go build -o bin/aiftd ./cmd/aift

bin/aiftd discovery scan >/dev/null
bin/aiftd discovery list >/dev/null
bin/aiftd discovery object AIFT-OS >/dev/null
bin/aiftd discovery report >/dev/null

test -f registry/discovery.json
test -f reports/discovery.md

echo "OK: discovery smoke passed"
SH
chmod +x tests/discovery-smoke.sh

cat > docs/DISCOVERY-ENGINE.md <<'DOC'
# AIFT Discovery Engine

The Discovery Engine discovers repository reality from evidence.

It does not hard-code repository names, services, runtimes, capabilities, package managers, documentation systems, workflows, or schemas.

## Commands

- `aiftd discovery scan`
- `aiftd discovery list`
- `aiftd discovery object <id-or-name>`
- `aiftd discovery report`

## Generated runtime artifacts

- `registry/discovery.json`
- `reports/discovery.md`

These are ignored runtime state and should be regenerated from truth.

## Discovery evidence

The engine currently detects:

- Git repositories
- Documentation
- Schemas
- Workflows
- Package manifests
- Go modules
- Node packages
- Rust crates
- Python projects
- Docker projects
- AIFT contracts
- Commands
- Capabilities
- Services
- Health checks
DOC

cat > schemas/discovery.schema.json <<'JSON'
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "AIFT Discovery Snapshot",
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
        "required": ["id", "kind", "name", "status", "location", "evidence", "generatedAt"],
        "properties": {
          "id": { "type": "string" },
          "kind": { "type": "string" },
          "name": { "type": "string" },
          "status": { "type": "string" },
          "location": { "type": "string" },
          "evidence": { "type": "array" },
          "runtimes": { "type": "array" },
          "manifests": { "type": "array" },
          "docs": { "type": "array" },
          "schemas": { "type": "array" },
          "workflows": { "type": "array" },
          "commands": { "type": "object" },
          "capabilities": { "type": "array" },
          "services": { "type": "array" },
          "healthChecks": { "type": "array" }
        }
      }
    }
  }
}
JSON

echo "== Verify =="
gofmt -w internal/discoveryengine/discovery.go cmd/aift/main.go
go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"
sh tests/discovery-smoke.sh

cat > "$REPORT" <<EOF
# Discovery Engine Implementation Report

Generated: $STAMP

Passed:

- gofmt
- go test ./...
- go build ./cmd/aift
- tests/discovery-smoke.sh

Generated runtime artifacts are intentionally ignored:

- registry/discovery.json
- reports/discovery.md
EOF

cp "$0" AI-Code-Training/scripts/phase-scripts/phase19-discovery-engine.sh 2>/dev/null || true

echo "== Stage source files only =="
git add \
  internal/discoveryengine/discovery.go \
  cmd/aift/main.go \
  tests/discovery-smoke.sh \
  docs/DISCOVERY-ENGINE.md \
  schemas/discovery.schema.json \
  AI-Code-Training/scripts/phase-scripts/phase19-discovery-engine.sh

echo "== Commit and push =="
if git diff --cached --quiet; then
  echo "Nothing staged."
else
  git commit -m "Implement discovery engine foundation"
  git push origin main
fi

echo "DONE"
