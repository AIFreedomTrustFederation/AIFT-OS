#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

export AIFT_ROOT="$ROOT"
export AIFT_OS_HOME="$OS"

cd "$OS" || exit 1

echo "== AIFT-OS Phase 8: Federation Integration Layer =="

mkdir -p \
  internal/federation \
  internal/repo \
  internal/workflow \
  docs \
  tests \
  registry \
  reports \
  var/events \
  bin

cat > internal/repo/repo.go <<'GO'
package repo

import (
"fmt"
"os"
"os/exec"
"path/filepath"
"strings"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/gitx"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Info struct {
Name          string `json:"name"`
Path          string `json:"path"`
Branch        string `json:"branch"`
Remote        string `json:"remote"`
Dirty         bool   `json:"dirty"`
ManifestPath  string `json:"manifestPath"`
ManifestValid bool   `json:"manifestValid"`
CommandsPath  string `json:"commandsPath"`
WorkflowsPath string `json:"workflowsPath"`
}

func List(cfg config.Config) ([]Info, error) {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return nil, err
}

out := make([]Info, 0, len(repos))
for _, r := range repos {
out = append(out, InspectRepo(r))
}

return out, nil
}

func Inspect(cfg config.Config, name string) (Info, error) {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return Info{}, err
}

for _, r := range repos {
if r.Name == name || filepath.Base(r.Path) == name {
return InspectRepo(r), nil
}
}

return Info{}, fmt.Errorf("repository not found: %s", name)
}

func InspectRepo(r workspace.Repo) Info {
return Info{
Name:          r.Name,
Path:          r.Path,
Branch:        gitx.Branch(r.Path),
Remote:        gitx.Remote(r.Path),
Dirty:         gitx.Dirty(r.Path),
ManifestPath:  manifests.Path(r.Path),
ManifestValid: manifests.Valid(r.Path),
CommandsPath:  filepath.Join(r.Path, ".aift", "commands"),
WorkflowsPath: filepath.Join(r.Path, ".aift", "workflows.json"),
}
}

func PrintList(cfg config.Config) error {
repos, err := List(cfg)
if err != nil {
return err
}

fmt.Printf("%-32s %-12s %-8s %-8s %s\n", "REPOSITORY", "BRANCH", "STATE", "MANIFEST", "REMOTE")
for _, r := range repos {
state := "clean"
if r.Dirty {
state = "dirty"
}
manifest := "valid"
if !r.ManifestValid {
manifest = "missing"
}
fmt.Printf("%-32s %-12s %-8s %-8s %s\n", r.Name, r.Branch, state, manifest, r.Remote)
}

return nil
}

func PrintInspect(cfg config.Config, name string) error {
r, err := Inspect(cfg, name)
if err != nil {
return err
}

fmt.Println("Repository:", r.Name)
fmt.Println("Path:", r.Path)
fmt.Println("Branch:", r.Branch)
fmt.Println("Remote:", r.Remote)
fmt.Println("Dirty:", r.Dirty)
fmt.Println("Manifest:", r.ManifestPath)
fmt.Println("Manifest valid:", r.ManifestValid)
fmt.Println("Commands:", r.CommandsPath)
fmt.Println("Workflows:", r.WorkflowsPath)

return nil
}

func RunCommand(cfg config.Config, name string, commandName string, args []string) error {
r, err := Inspect(cfg, name)
if err != nil {
return err
}

script := filepath.Join(r.CommandsPath, commandName+".sh")
if _, err := os.Stat(script); err != nil {
return fmt.Errorf("repo command not found: %s", script)
}

cmd := exec.Command("sh", append([]string{script}, args...)...)
cmd.Dir = r.Path
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
cmd.Stdin = os.Stdin
cmd.Env = append(os.Environ(),
"AIFT_REPO_NAME="+r.Name,
"AIFT_REPO_PATH="+r.Path,
)

return cmd.Run()
}

func EnsureExampleCommand(cfg config.Config) error {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return err
}

for _, r := range repos {
dir := filepath.Join(r.Path, ".aift", "commands")
if err := os.MkdirAll(dir, 0755); err != nil {
return err
}
script := filepath.Join(dir, "status.sh")
if _, err := os.Stat(script); err == nil {
continue
}
content := "#!/usr/bin/env sh\nset -eu\necho \"AIFT repo: ${AIFT_REPO_NAME:-unknown}\"\ngit status --short\n"
if err := os.WriteFile(script, []byte(content), 0755); err != nil {
return err
}
}

return nil
}

func NormalizeName(s string) string {
return strings.TrimSpace(s)
}
GO

cat > internal/workflow/workflow.go <<'GO'
package workflow

import (
"encoding/json"
"fmt"
"os"
"path/filepath"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type WorkflowStep struct {
Name    string   `json:"name"`
Command string   `json:"command"`
Args    []string `json:"args,omitempty"`
}

type Workflow struct {
Name        string         `json:"name"`
Description string         `json:"description"`
Steps       []WorkflowStep `json:"steps"`
}

func Defaults() []Workflow {
return []Workflow{
{
Name:        "verify-federation",
Description: "Generate manifests, registry, providers, reports, and dependency graph.",
Steps: []WorkflowStep{
{Name: "manifest", Command: "manifest"},
{Name: "registry", Command: "registry"},
{Name: "providers", Command: "providers"},
{Name: "dashboard", Command: "dashboard"},
{Name: "deps", Command: "deps"},
{Name: "verify", Command: "verify"},
},
},
{
Name:        "safe-sync",
Description: "Run safe federation sync without auto-committing dirty repositories.",
Steps: []WorkflowStep{
{Name: "sync-safe", Command: "sync", Args: []string{"--safe"}},
},
},
}
}

func WriteRegistry(cfg config.Config) error {
out := filepath.Join(cfg.OSHome, "registry", "workflows.json")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}

data, err := json.MarshalIndent(Defaults(), "", "  ")
if err != nil {
return err
}

if err := os.WriteFile(out, append(data, '\n'), 0644); err != nil {
return err
}

fmt.Println("Wrote", out)
return nil
}

func List(cfg config.Config) error {
if err := WriteRegistry(cfg); err != nil {
return err
}

fmt.Printf("%-24s %s\n", "WORKFLOW", "DESCRIPTION")
for _, wf := range Defaults() {
fmt.Printf("%-24s %s\n", wf.Name, wf.Description)
}

return nil
}
GO

cat > internal/federation/federation.go <<'GO'
package federation

import (
"encoding/json"
"fmt"
"os"
"path/filepath"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/providers"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/repo"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workflow"
)

type Snapshot struct {
Repos int `json:"repos"`
Dirty int `json:"dirty"`
ValidManifests int `json:"validManifests"`
}

func Scan(cfg config.Config) error {
if err := manifests.EnsureAll(cfg); err != nil {
return err
}
if err := repo.EnsureExampleCommand(cfg); err != nil {
return err
}
if err := providers.WriteRegistry(cfg); err != nil {
return err
}
if err := workflow.WriteRegistry(cfg); err != nil {
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

snap, err := SnapshotState(cfg)
if err != nil {
return err
}

out := filepath.Join(cfg.OSHome, "registry", "federation-snapshot.json")
data, err := json.MarshalIndent(snap, "", "  ")
if err != nil {
return err
}
if err := os.WriteFile(out, append(data, '\n'), 0644); err != nil {
return err
}

if err := events.Emit(cfg, "federation.scan", "federation", "federation scan complete", map[string]string{
"repos": fmt.Sprint(snap.Repos),
"dirty": fmt.Sprint(snap.Dirty),
}); err != nil {
return err
}

fmt.Println("Wrote", out)
return nil
}

func SnapshotState(cfg config.Config) (Snapshot, error) {
repos, err := repo.List(cfg)
if err != nil {
return Snapshot{}, err
}

var snap Snapshot
snap.Repos = len(repos)
for _, r := range repos {
if r.Dirty {
snap.Dirty++
}
if r.ManifestValid {
snap.ValidManifests++
}
}

return snap, nil
}

func Verify(cfg config.Config) error {
if err := Scan(cfg); err != nil {
return err
}
snap, err := SnapshotState(cfg)
if err != nil {
return err
}
fmt.Printf("Federation verified: repos=%d dirty=%d validManifests=%d\n", snap.Repos, snap.Dirty, snap.ValidManifests)
return nil
}

func Graph(cfg config.Config) error {
if err := reports.Deps(cfg); err != nil {
return err
}
fmt.Println("Wrote", filepath.Join(cfg.OSHome, "reports", "dependency-graph.md"))
return nil
}
GO

python - <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text()

imports = {
    '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/federation"':
        '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"',
    '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/repo"':
        '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports"',
    '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workflow"':
        '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"',
}

for newimp, after in imports.items():
    if newimp not in s:
        s = s.replace(after, after + "\n\t" + newimp)

help_add = [
    'fmt.Println("  federation scan|graph|verify")',
    'fmt.Println("  repo list|inspect|run")',
    'fmt.Println("  workflow list")',
]
if help_add[0] not in s:
    s = s.replace(
        'fmt.Println("  verify")',
        'fmt.Println("  verify")\n\t' + "\n\t".join(help_add),
    )

case_block = r'''case "federation":
if len(args) == 0 || args[0] == "scan" {
err = federation.Scan(cfg)
} else if args[0] == "graph" {
err = federation.Graph(cfg)
} else if args[0] == "verify" {
err = federation.Verify(cfg)
} else {
err = fmt.Errorf("usage: aift federation scan|graph|verify")
}
case "repo":
if len(args) == 0 || args[0] == "list" {
err = repo.PrintList(cfg)
} else if args[0] == "inspect" {
if len(args) < 2 {
err = fmt.Errorf("usage: aift repo inspect <name>")
} else {
err = repo.PrintInspect(cfg, repo.NormalizeName(args[1]))
}
} else if args[0] == "run" {
if len(args) < 3 {
err = fmt.Errorf("usage: aift repo run <name> <command> [args...]")
} else {
err = repo.RunCommand(cfg, repo.NormalizeName(args[1]), args[2], args[3:])
}
} else {
err = fmt.Errorf("usage: aift repo list|inspect|run")
}
case "workflow":
if len(args) == 0 || args[0] == "list" {
err = workflow.List(cfg)
} else {
err = fmt.Errorf("usage: aift workflow list")
}
'''

if 'case "federation":' not in s:
    s = s.replace('case "verify":\n\t\terr = verify(cfg)', case_block + '\tcase "verify":\n\t\terr = verify(cfg)')

p.write_text(s)
PY

cat > tests/federation-integration-smoke.sh <<'SH'
#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" repo list >/dev/null
"$ROOT/aift" repo inspect AIFT-OS >/dev/null
"$ROOT/aift" workflow list >/dev/null
"$ROOT/aift" federation scan >/dev/null
"$ROOT/aift" federation graph >/dev/null
"$ROOT/aift" federation verify >/dev/null

test -f registry/federation-snapshot.json
test -f registry/workflows.json

echo "OK: federation integration smoke passed"
SH
chmod +x tests/federation-integration-smoke.sh

cat > docs/PHASE-8-FEDERATION-INTEGRATION.md <<'DOC'
# Phase 8: Federation Integration Layer

AIFT-OS now integrates sovereign repositories through manifest contracts, repository inspection, federation scanning, workflows, and repo command execution.

Commands:

- `aift repo list`
- `aift repo inspect <name>`
- `aift repo run <name> <command> [args...]`
- `aift workflow list`
- `aift federation scan`
- `aift federation graph`
- `aift federation verify`

Generated files:

- `registry/federation-snapshot.json`
- `registry/workflows.json`

Each repository remains sovereign. AIFT-OS discovers and coordinates through `.aift/` contracts.
DOC

echo "== Build/test =="
go clean -cache
gofmt -w cmd internal
go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" federation scan
"$ROOT/aift" repo list
"$ROOT/aift" workflow list
sh tests/federation-integration-smoke.sh

echo "== Commit/push =="
git add .
if git diff --cached --quiet; then
  echo "Nothing staged to commit."
else
  git commit -m "Add AIFT-OS federation integration layer"
fi

git push origin main

echo
echo "DONE."
echo "Try:"
echo "  ~/AIFT/aift federation verify"
echo "  ~/AIFT/aift repo inspect AIFT-Forge"
echo "  ~/AIFT/aift workflow list"
