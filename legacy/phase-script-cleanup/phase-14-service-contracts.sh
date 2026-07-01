#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

export AIFT_ROOT="$ROOT"
export AIFT_OS_HOME="$OS"

cd "$OS" || exit 1

echo "== Phase 14: Federation Service Contracts =="

mkdir -p internal/servicecontracts docs tests registry reports schemas var/events AI-Code-Training/scripts/phase-scripts bin

cat > internal/servicecontracts/servicecontracts.go <<'GO'
package servicecontracts

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

type Service struct {
Name        string   `json:"name"`
Kind        string   `json:"kind"`
Status      string   `json:"status"`
Version     string   `json:"version"`
Provides    []string `json:"provides"`
Requires    []string `json:"requires"`
Events      []string `json:"events"`
Health      string   `json:"health,omitempty"`
Start       string   `json:"start,omitempty"`
Stop        string   `json:"stop,omitempty"`
Evidence    string   `json:"evidence"`
Description string   `json:"description"`
}

type Contract struct {
Repo        string    `json:"repo"`
Services    []Service `json:"services"`
GeneratedAt string    `json:"generatedAt"`
}

type Registry struct {
GeneratedAt string     `json:"generatedAt"`
Contracts   []Contract `json:"contracts"`
Services    []ServiceRecord `json:"services"`
}

type ServiceRecord struct {
Repo     string `json:"repo"`
Name     string `json:"name"`
Kind     string `json:"kind"`
Status   string `json:"status"`
Version  string `json:"version"`
Evidence string `json:"evidence"`
}

func InitAll(cfg config.Config) error {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return err
}

for _, r := range repos {
if err := InitRepo(r.Name, r.Path); err != nil {
return err
}
}

return Scan(cfg)
}

func InitRepo(name string, repoPath string) error {
dir := filepath.Join(repoPath, ".aift")
if err := os.MkdirAll(filepath.Join(dir, "services"), 0755); err != nil {
return err
}

path := filepath.Join(dir, "services.json")
if _, err := os.Stat(path); err == nil {
return nil
}

contract := Contract{
Repo: name,
Services: []Service{
{
Name:        name + ".service",
Kind:        inferKind(name, repoPath),
Status:      "planned",
Version:     "0.1.0",
Provides:    []string{},
Requires:    []string{},
Events:      []string{"repo.changed", "capability.changed", "manual.changed"},
Evidence:    "default service contract",
Description: "Default planned federation service contract for " + name,
},
},
GeneratedAt: time.Now().Format(time.RFC3339),
}

data, err := json.MarshalIndent(contract, "", "  ")
if err != nil {
return err
}

return os.WriteFile(path, append(data, '\n'), 0644)
}

func Scan(cfg config.Config) error {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return err
}

var contracts []Contract
var services []ServiceRecord

for _, r := range repos {
c, ok := readContract(r.Name, r.Path)
if !ok {
continue
}
contracts = append(contracts, c)

for _, svc := range c.Services {
if svc.Status == "" {
svc.Status = "planned"
}
if svc.Version == "" {
svc.Version = "0.1.0"
}
if svc.Evidence == "" {
svc.Evidence = ".aift/services.json"
}

services = append(services, ServiceRecord{
Repo:     c.Repo,
Name:     svc.Name,
Kind:     svc.Kind,
Status:   svc.Status,
Version:  svc.Version,
Evidence: svc.Evidence,
})
}
}

sort.Slice(services, func(i, j int) bool {
if services[i].Repo == services[j].Repo {
return services[i].Name < services[j].Name
}
return services[i].Repo < services[j].Repo
})

reg := Registry{
GeneratedAt: time.Now().Format(time.RFC3339),
Contracts:   contracts,
Services:    services,
}

if err := writeRegistry(cfg, reg); err != nil {
return err
}
if err := writeReport(cfg, reg); err != nil {
return err
}

return events.Emit(cfg, "services.scan", "servicecontracts", "service contracts scanned", map[string]string{
"services": fmt.Sprint(len(services)),
})
}

func List(cfg config.Config) error {
reg, err := loadOrScan(cfg)
if err != nil {
return err
}

fmt.Printf("%-30s %-34s %-16s %-12s %s\n", "REPO", "SERVICE", "KIND", "STATUS", "VERSION")
for _, s := range reg.Services {
fmt.Printf("%-30s %-34s %-16s %-12s %s\n", s.Repo, s.Name, s.Kind, s.Status, s.Version)
}
return nil
}

func Repo(cfg config.Config, name string) error {
reg, err := loadOrScan(cfg)
if err != nil {
return err
}

found := false
for _, c := range reg.Contracts {
if c.Repo != name {
continue
}
found = true
fmt.Println("Repository:", c.Repo)
for _, svc := range c.Services {
fmt.Println()
fmt.Println("Service:", svc.Name)
fmt.Println("Kind:", svc.Kind)
fmt.Println("Status:", svc.Status)
fmt.Println("Version:", svc.Version)
fmt.Println("Provides:", strings.Join(svc.Provides, ", "))
fmt.Println("Requires:", strings.Join(svc.Requires, ", "))
fmt.Println("Events:", strings.Join(svc.Events, ", "))
fmt.Println("Health:", svc.Health)
fmt.Println("Start:", svc.Start)
fmt.Println("Stop:", svc.Stop)
fmt.Println("Evidence:", svc.Evidence)
}
}

if !found {
return fmt.Errorf("repository not found or no service contract: %s", name)
}
return nil
}

func Report(cfg config.Config) error {
path := filepath.Join(cfg.OSHome, "reports", "service-contracts.md")
data, err := os.ReadFile(path)
if err != nil {
if err := Scan(cfg); err != nil {
return err
}
data, err = os.ReadFile(path)
if err != nil {
return err
}
}
fmt.Print(string(data))
return nil
}

func readContract(name string, repoPath string) (Contract, bool) {
path := filepath.Join(repoPath, ".aift", "services.json")
data, err := os.ReadFile(path)
if err != nil {
return Contract{}, false
}

var c Contract
if json.Unmarshal(data, &c) != nil {
return Contract{}, false
}
if c.Repo == "" {
c.Repo = name
}
return c, true
}

func writeRegistry(cfg config.Config, reg Registry) error {
out := filepath.Join(cfg.OSHome, "registry", "service-contracts.json")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}
data, err := json.MarshalIndent(reg, "", "  ")
if err != nil {
return err
}
if err := os.WriteFile(out, append(data, '\n'), 0644); err != nil {
return err
}
fmt.Println("Wrote", out)
return nil
}

func writeReport(cfg config.Config, reg Registry) error {
out := filepath.Join(cfg.OSHome, "reports", "service-contracts.md")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}

var b strings.Builder
b.WriteString("# Federation Service Contracts\n\n")
b.WriteString("Service contracts declare what each repo provides, requires, and can eventually run.\n\n")
b.WriteString("AIFT-OS records these contracts truthfully. Planned services are not executed.\n\n")
b.WriteString("| Repository | Service | Kind | Status | Version | Evidence |\n")
b.WriteString("|---|---|---|---|---|---|\n")
for _, s := range reg.Services {
b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | `%s` | %s |\n", s.Repo, s.Name, s.Kind, s.Status, s.Version, s.Evidence))
}

return os.WriteFile(out, []byte(b.String()), 0644)
}

func loadOrScan(cfg config.Config) (Registry, error) {
path := filepath.Join(cfg.OSHome, "registry", "service-contracts.json")
data, err := os.ReadFile(path)
if err != nil {
if err := Scan(cfg); err != nil {
return Registry{}, err
}
data, err = os.ReadFile(path)
if err != nil {
return Registry{}, err
}
}

var reg Registry
if err := json.Unmarshal(data, &reg); err != nil {
return Registry{}, err
}
return reg, nil
}

func inferKind(name string, repoPath string) string {
lower := strings.ToLower(name)

switch {
case strings.Contains(lower, "aift-os"):
return "control-plane"
case strings.Contains(lower, "forge"):
return "forge"
case strings.Contains(lower, "booksmith"):
return "publishing"
case strings.Contains(lower, "vps"):
return "infrastructure"
case strings.Contains(lower, "www") || strings.Contains(lower, "github.io"):
return "website"
case exists(filepath.Join(repoPath, "package.json")):
return "web-app"
case exists(filepath.Join(repoPath, "go.mod")):
return "go-service"
default:
return "repository"
}
}

func exists(path string) bool {
_, err := os.Stat(path)
return err == nil
}
GO

python - <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text()

if '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/servicecontracts"' not in s:
    s = s.replace(
        '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/services"',
        '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/services"\n\t"github.com/AIFreedomTrustFederation/AIFT-OS/internal/servicecontracts"'
    )

if 'fmt.Println("  service-contracts init-all|scan|list|repo|report")' not in s:
    s = s.replace(
        'fmt.Println("  mesh init-all|scan|topics|subscribers|publish|replay|tail|report")',
        'fmt.Println("  mesh init-all|scan|topics|subscribers|publish|replay|tail|report")\n\tfmt.Println("  service-contracts init-all|scan|list|repo|report")'
    )

case_block = r'''case "service-contracts":
err = runServiceContracts(cfg, args)
'''

if 'case "service-contracts":' not in s:
    s = s.replace('case "verify":\n\t\terr = verify(cfg)', case_block + '\tcase "verify":\n\t\terr = verify(cfg)')

func_block = r'''
func runServiceContracts(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
return servicecontracts.Scan(cfg)
}
switch args[0] {
case "init-all":
return servicecontracts.InitAll(cfg)
case "list":
return servicecontracts.List(cfg)
case "repo":
if len(args) < 2 {
return fmt.Errorf("usage: aift service-contracts repo <repo>")
}
return servicecontracts.Repo(cfg, args[1])
case "report":
return servicecontracts.Report(cfg)
default:
return fmt.Errorf("usage: aift service-contracts init-all|scan|list|repo|report")
}
}
'''

if 'func runServiceContracts(' not in s:
    s = s.replace('func verify(cfg config.Config) error {', func_block + '\nfunc verify(cfg config.Config) error {')

if 'if err := servicecontracts.Scan(cfg); err != nil {' not in s:
    s = s.replace(
        'if err := eventmesh.Scan(cfg); err != nil {\n\t\treturn err\n\t}',
        'if err := eventmesh.Scan(cfg); err != nil {\n\t\treturn err\n\t}\n\tif err := servicecontracts.Scan(cfg); err != nil {\n\t\treturn err\n\t}'
    )

p.write_text(s)
PY

cat > schemas/service-contracts.schema.json <<'JSON'
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "AIFT Service Contracts Registry",
  "type": "object",
  "required": ["generatedAt", "contracts", "services"],
  "properties": {
    "generatedAt": { "type": "string" },
    "contracts": { "type": "array" },
    "services": { "type": "array" }
  }
}
JSON

cat > schemas/service-contract.schema.json <<'JSON'
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "AIFT Repo Service Contract",
  "type": "object",
  "required": ["repo", "services"],
  "properties": {
    "repo": { "type": "string" },
    "services": { "type": "array" },
    "generatedAt": { "type": "string" }
  }
}
JSON

cat > tests/service-contracts-smoke.sh <<'SH'
#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" service-contracts init-all >/dev/null
"$ROOT/aift" service-contracts scan >/dev/null
"$ROOT/aift" service-contracts list >/dev/null
"$ROOT/aift" service-contracts repo AIFT-OS >/dev/null
"$ROOT/aift" service-contracts report >/dev/null
"$ROOT/aift" verify >/dev/null

test -f registry/service-contracts.json
test -f reports/service-contracts.md
test -f .aift/services.json
test -d .aift/services

echo "OK: service contracts smoke passed"
SH

chmod +x tests/service-contracts-smoke.sh

cat > docs/PHASE-14-SERVICE-CONTRACTS.md <<'DOC'
# Phase 14: Federation Service Contracts

AIFT-OS now gives every repository a service contract.

## Principle

Event contracts describe how repositories communicate.

Service contracts describe what repositories provide, require, and may eventually run.

A planned service is not executed.

## Commands

- `aift service-contracts init-all`
- `aift service-contracts scan`
- `aift service-contracts list`
- `aift service-contracts repo <repo>`
- `aift service-contracts report`

## Per Repo

- `.aift/services.json`
- `.aift/services/`

## Generated

- `registry/service-contracts.json`
- `reports/service-contracts.md`

## Truth Rule

AIFT-OS records service promises, but does not treat them as executable until matching capabilities and health checks are ready or v1.
DOC

cp phase-14-service-contracts.sh AI-Code-Training/scripts/phase-scripts/phase-14-service-contracts.sh 2>/dev/null || true

echo "== Build/test =="
go clean -cache
gofmt -w cmd internal
go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" service-contracts init-all
"$ROOT/aift" service-contracts scan
"$ROOT/aift" service-contracts list
sh tests/service-contracts-smoke.sh

echo "== Commit AIFT-OS =="
git add .
if git diff --cached --quiet; then
  echo "AIFT-OS: nothing staged."
else
  git commit -m "Add federation service contracts"
fi
git push origin main

echo "== Commit service contracts in federation repos =="
for repo in "$ROOT"/*; do
  [ -d "$repo/.git" ] || continue
  name="$(basename "$repo")"
  cd "$repo" || continue

  git add .aift/services.json .aift/services 2>/dev/null || true

  if git diff --cached --quiet; then
    echo "$name: no service contract changes."
  else
    git commit -m "Add AIFT service contract"
    branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo main)"
    if git remote get-url origin >/dev/null 2>&1; then
      git push origin "$branch" || true
    fi
  fi
done

cd "$OS" || exit 1

echo
echo "DONE."
echo "Try:"
echo "  ~/AIFT/aift service-contracts list"
echo "  ~/AIFT/aift service-contracts repo AIFT-OS"
echo "  ~/AIFT/aift service-contracts report"
