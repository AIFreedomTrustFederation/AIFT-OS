#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

export AIFT_ROOT="$ROOT"
export AIFT_OS_HOME="$OS"

cd "$OS" || exit 1

echo "== Phase 11: Federation Manual Contracts =="

mkdir -p internal/manual docs tests registry reports var/events AI-Code-Training/scripts/phase-scripts bin

cat > internal/manual/manual.go <<'GO'
package manual

import (
"encoding/json"
"fmt"
"os"
"path/filepath"
"strings"
"time"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Status struct {
ManualSource string `json:"manualSource"`
PDFBuild     string `json:"pdfBuild"`
WebPublish   string `json:"webPublish"`
}

type Builder struct {
Repo       string `json:"repo"`
Capability string `json:"capability"`
Status     string `json:"status"`
}

type Contract struct {
Repo        string   `json:"repo"`
Title       string   `json:"title"`
SourcePath  string   `json:"sourcePath"`
AssetsPath  string   `json:"assetsPath"`
Builder     Builder  `json:"builder"`
Status      Status   `json:"status"`
ManualType  string   `json:"manualType"`
Format      string   `json:"format"`
Sections    []string `json:"sections"`
GeneratedAt string   `json:"generatedAt"`
}

type Registry struct {
GeneratedAt string     `json:"generatedAt"`
Manuals     []Contract `json:"manuals"`
}

func InitAll(cfg config.Config) error {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return err
}

for _, r := range repos {
if err := InitRepo(cfg, r); err != nil {
return err
}
}

return Scan(cfg)
}

func InitRepo(cfg config.Config, r workspace.Repo) error {
base := filepath.Join(r.Path, "docs", "manual")
source := filepath.Join(base, "source")
assets := filepath.Join(base, "assets")

dirs := []string{
filepath.Join(source, "man0"),
filepath.Join(source, "man1"),
filepath.Join(source, "man2"),
filepath.Join(source, "man3"),
filepath.Join(source, "man4"),
filepath.Join(source, "man5"),
filepath.Join(source, "man6"),
filepath.Join(source, "man7"),
filepath.Join(source, "man8"),
filepath.Join(source, "man9"),
assets,
filepath.Join(r.Path, ".aift"),
}

for _, d := range dirs {
if err := os.MkdirAll(d, 0755); err != nil {
return err
}
}

writeIfMissing(filepath.Join(base, "README.md"), manualReadme(r.Name))
writeIfMissing(filepath.Join(source, "index.md"), manualIndex(r.Name))
writeIfMissing(filepath.Join(source, "man0", "00-introduction.md"), introPage(r.Name))
writeIfMissing(filepath.Join(source, "man7", "modularity.md"), modularityPage(r.Name))
writeIfMissing(filepath.Join(source, "man7", "truthfulness.md"), truthfulnessPage(r.Name))
writeIfMissing(filepath.Join(source, "man7", "booksmith-pipeline.md"), booksmithPage(r.Name))

contract := BuildContract(cfg, r)
data, err := json.MarshalIndent(contract, "", "  ")
if err != nil {
return err
}
return os.WriteFile(filepath.Join(r.Path, ".aift", "manual.json"), append(data, '\n'), 0644)
}

func Scan(cfg config.Config) error {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return err
}

reg := Registry{
GeneratedAt: time.Now().Format(time.RFC3339),
Manuals:     []Contract{},
}

for _, r := range repos {
reg.Manuals = append(reg.Manuals, BuildContract(cfg, r))
}

if err := writeRegistry(cfg, reg); err != nil {
return err
}
if err := writeReport(cfg, reg); err != nil {
return err
}

return events.Emit(cfg, "manual.scan", "manual", "federation manual scan complete", map[string]string{
"manuals": fmt.Sprint(len(reg.Manuals)),
})
}

func Repo(cfg config.Config, name string) error {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return err
}
for _, r := range repos {
if r.Name == name {
c := BuildContract(cfg, r)
fmt.Println("Repository:", c.Repo)
fmt.Println("Title:", c.Title)
fmt.Println("Source:", c.SourcePath)
fmt.Println("Assets:", c.AssetsPath)
fmt.Println("Builder:", c.Builder.Repo)
fmt.Println("Builder capability:", c.Builder.Capability)
fmt.Println("Builder status:", c.Builder.Status)
fmt.Println("Manual source:", c.Status.ManualSource)
fmt.Println("PDF build:", c.Status.PDFBuild)
fmt.Println("Web publish:", c.Status.WebPublish)
return nil
}
}
return fmt.Errorf("repository not found: %s", name)
}

func Report(cfg config.Config) error {
path := filepath.Join(cfg.OSHome, "reports", "manuals.md")
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

func BuildContract(cfg config.Config, r workspace.Repo) Contract {
sourceRel := "docs/manual/source"
assetsRel := "docs/manual/assets"

sourceStatus := "planned"
if dirExists(filepath.Join(r.Path, sourceRel)) {
sourceStatus = "ready"
}

builderStatus := "planned"
booksmithPath := filepath.Join(cfg.Root, "booksmith-ai")
if !dirExists(booksmithPath) {
booksmithPath = filepath.Join(cfg.Root, "BookSmith-Federation-OS")
}
if fileExists(filepath.Join(booksmithPath, ".aift", "commands", "manual-build.sh")) {
builderStatus = "ready"
}

return Contract{
Repo:       r.Name,
Title:      titleFor(r.Name),
SourcePath: sourceRel,
AssetsPath: assetsRel,
Builder: Builder{
Repo:       "booksmith-ai",
Capability: "manual.build.pdf",
Status:     builderStatus,
},
Status: Status{
ManualSource: sourceStatus,
PDFBuild:     builderStatus,
WebPublish:   "planned",
},
ManualType:  "unix-style-federation-manual",
Format:      "markdown-source-booksmith-built-pdf",
Sections:    []string{"man0", "man1", "man2", "man3", "man4", "man5", "man6", "man7", "man8", "man9"},
GeneratedAt: time.Now().Format(time.RFC3339),
}
}

func writeRegistry(cfg config.Config, reg Registry) error {
out := filepath.Join(cfg.OSHome, "registry", "manuals.json")
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
out := filepath.Join(cfg.OSHome, "reports", "manuals.md")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}

var b strings.Builder
b.WriteString("# Federation Manual Contracts\n\n")
b.WriteString("Every repo owns manual source. BookSmith owns PDF/build/publishing. AIFT-OS owns discovery, truth, contracts, and reports.\n\n")
b.WriteString("| Repository | Source | PDF Build | Web Publish | Builder |\n")
b.WriteString("|---|---|---|---|---|\n")
for _, m := range reg.Manuals {
b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | `%s:%s` |\n",
m.Repo, m.Status.ManualSource, m.Status.PDFBuild, m.Status.WebPublish, m.Builder.Repo, m.Builder.Capability))
}

if err := os.WriteFile(out, []byte(b.String()), 0644); err != nil {
return err
}
fmt.Println("Wrote", out)
return nil
}

func writeIfMissing(path string, content string) error {
if fileExists(path) {
return nil
}
if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
return err
}
return os.WriteFile(path, []byte(content), 0644)
}

func manualReadme(repo string) string {
return "# " + titleFor(repo) + " Manual\n\nThis directory contains the UNIX-style manual source for this repository.\n\nManual source belongs to this repo. PDF and publication builds belong to BookSmith.\n\n"
}

func manualIndex(repo string) string {
return "# " + titleFor(repo) + " Manual Index\n\n- `man0/` — Doctrine and introduction\n- `man1/` — User commands\n- `man2/` — System calls and internal operations\n- `man3/` — Libraries and APIs\n- `man4/` — Devices and providers\n- `man5/` — File formats\n- `man6/` — Federation applications\n- `man7/` — Concepts, doctrines, standards\n- `man8/` — Administration\n- `man9/` — Kernel/internal interfaces\n\n"
}

func introPage(repo string) string {
return "# " + titleFor(repo) + " Introduction\n\n## NAME\n\n" + repo + " manual introduction\n\n## DESCRIPTION\n\nThis manual page defines the role, purpose, and federation contract for this repository.\n\n## STATUS\n\nManual source: ready\n\nPDF build: planned until BookSmith exposes `manual.build.pdf` as a verified capability.\n\n"
}

func modularityPage(repo string) string {
return "# Modularity Doctrine\n\n## NAME\n\nmodularity — replaceable modules and provider-agnostic architecture\n\n## DESCRIPTION\n\nNo dependency is sacred. Only the contract is sacred.\n\nThis repository should declare modules, providers, capabilities, events, and replacement paths through `.aift/` contracts.\n\n"
}

func truthfulnessPage(repo string) string {
return "# Truthfulness Doctrine\n\n## NAME\n\ntruthfulness — evidence before orchestration\n\n## DESCRIPTION\n\nAIFT-OS must never claim this repository can perform an action until that capability is verified.\n\nCapabilities move through planned, detected, ready, v1, broken, and deprecated states.\n\n"
}

func booksmithPage(repo string) string {
return "# BookSmith Manual Pipeline\n\n## NAME\n\nbooksmith-pipeline — federation manual PDF and publication pipeline\n\n## DESCRIPTION\n\nThis repository owns its manual source. BookSmith owns compilation, PDF generation, proofing, publishing packet generation, and static web export.\n\n## SOURCE\n\n`docs/manual/source/`\n\n## CONTRACT\n\n`.aift/manual.json`\n\n## STATUS\n\nBookSmith PDF build remains planned until a verified `.aift/commands/manual-build.sh` capability exists in BookSmith.\n\n"
}

func titleFor(repo string) string {
return strings.ReplaceAll(repo, "-", " ") + " UNIX Manual"
}

func fileExists(path string) bool {
info, err := os.Stat(path)
return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
info, err := os.Stat(path)
return err == nil && info.IsDir()
}
GO

python - <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text()

if '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manual"' not in s:
    s = s.replace(
        '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests"',
        '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests"\n\t"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manual"'
    )

if 'fmt.Println("  manual init-all|scan|report|repo")' not in s:
    s = s.replace(
        'fmt.Println("  intelligence scan|report|repo|roadmap")',
        'fmt.Println("  intelligence scan|report|repo|roadmap")\n\tfmt.Println("  manual init-all|scan|report|repo")'
    )

case_block = r'''case "manual":
err = runManual(cfg, args)
'''

if 'case "manual":' not in s:
    s = s.replace('case "verify":\n\t\terr = verify(cfg)', case_block + '\tcase "verify":\n\t\terr = verify(cfg)')

func_block = r'''
func runManual(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
return manual.Scan(cfg)
}
if args[0] == "init-all" {
return manual.InitAll(cfg)
}
if args[0] == "report" {
return manual.Report(cfg)
}
if args[0] == "repo" {
if len(args) < 2 {
return fmt.Errorf("usage: aift manual repo <repo>")
}
return manual.Repo(cfg, args[1])
}
return fmt.Errorf("usage: aift manual init-all|scan|report|repo")
}
'''

if 'func runManual(' not in s:
    s = s.replace('func verify(cfg config.Config) error {', func_block + '\nfunc verify(cfg config.Config) error {')

if 'if err := manual.Scan(cfg); err != nil {' not in s:
    s = s.replace(
        'if err := intelligence.Scan(cfg); err != nil {\n\t\treturn err\n\t}',
        'if err := intelligence.Scan(cfg); err != nil {\n\t\treturn err\n\t}\n\tif err := manual.Scan(cfg); err != nil {\n\t\treturn err\n\t}'
    )

p.write_text(s)
PY

cat > tests/manual-contracts-smoke.sh <<'SH'
#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" manual init-all >/dev/null
"$ROOT/aift" manual scan >/dev/null
"$ROOT/aift" manual report >/dev/null
"$ROOT/aift" manual repo AIFT-OS >/dev/null
"$ROOT/aift" verify >/dev/null

test -f registry/manuals.json
test -f reports/manuals.md
test -f .aift/manual.json
test -d docs/manual/source/man0
test -d docs/manual/source/man7

echo "OK: manual contracts smoke passed"
SH

chmod +x tests/manual-contracts-smoke.sh

cat > docs/PHASE-11-MANUAL-CONTRACTS.md <<'DOC'
# Phase 11: Federation Manual Contracts

Every repository now receives a UNIX-style manual source structure and a `.aift/manual.json` contract.

## Principle

Every repo owns its manual source.

BookSmith owns PDF generation, publication packets, proofing, and static web export.

AIFT-OS owns discovery, truth, contracts, reporting, and orchestration readiness.

## Commands

- `aift manual init-all`
- `aift manual scan`
- `aift manual report`
- `aift manual repo <repo>`

## Generated Per Repo

- `.aift/manual.json`
- `docs/manual/README.md`
- `docs/manual/source/index.md`
- `docs/manual/source/man0/`
- `docs/manual/source/man1/`
- `docs/manual/source/man2/`
- `docs/manual/source/man3/`
- `docs/manual/source/man4/`
- `docs/manual/source/man5/`
- `docs/manual/source/man6/`
- `docs/manual/source/man7/`
- `docs/manual/source/man8/`
- `docs/manual/source/man9/`
- `docs/manual/assets/`

## Generated In AIFT-OS

- `registry/manuals.json`
- `reports/manuals.md`

## Status Truth

- `manual.source` becomes `ready` when source folders exist.
- `manual.pdfBuild` remains `planned` until BookSmith exposes a verified `manual.build.pdf` capability.
- `manual.webPublish` remains `planned` until the website/static export integration is proven.
DOC

cp phase-11-manual-contracts.sh AI-Code-Training/scripts/phase-scripts/phase-11-manual-contracts.sh 2>/dev/null || true

echo "== Build/test =="
go clean -cache
gofmt -w cmd internal
go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" manual init-all
"$ROOT/aift" manual report
sh tests/manual-contracts-smoke.sh

echo "== Commit AIFT-OS =="
git add .
if git diff --cached --quiet; then
  echo "AIFT-OS: nothing staged."
else
  git commit -m "Add federation manual contract system"
fi
git push origin main

echo "== Commit manual contracts in federation repos =="
for repo in "$ROOT"/*; do
  [ -d "$repo/.git" ] || continue
  name="$(basename "$repo")"
  cd "$repo" || continue

  git add .aift/manual.json docs/manual 2>/dev/null || true

  if git diff --cached --quiet; then
    echo "$name: no manual contract changes."
  else
    git commit -m "Add AIFT manual contract"
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
echo "  ~/AIFT/aift manual report"
echo "  ~/AIFT/aift manual repo booksmith-ai"
echo "  ~/AIFT/aift manual repo AIFT-OS"
