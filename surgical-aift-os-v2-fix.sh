#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

export AIFT_ROOT="$ROOT"
export AIFT_OS_HOME="$OS"

cd "$OS" || exit 1

echo "== Surgical AIFT-OS v2 fix =="

mkdir -p bin docs install scripts tests registry reports logs var

echo "== Fix Go CLI parser =="
cat > cmd/aift/main.go <<'GO'
package main

import (
"fmt"
"os"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/doctor"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/gitx"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/plugins"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/sync"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

func main() {
cfg := config.Load()

cmd := "help"
args := []string{}

if len(os.Args) > 1 {
cmd = os.Args[1]
args = os.Args[2:]
}

if looksLikeExecutablePath(cmd) {
if len(args) > 0 {
cmd = args[0]
args = args[1:]
} else {
cmd = "help"
}
}

var err error

switch cmd {
case "help", "-h", "--help":
help()
case "doctor":
err = doctor.Run(cfg)
case "status":
err = status(cfg)
case "manifest":
err = manifests.EnsureAll(cfg)
if err == nil {
fmt.Println("OK: manifests ensured")
}
case "registry":
err = registry.Generate(cfg)
case "dashboard":
err = reports.Dashboard(cfg)
case "deps":
err = reports.Deps(cfg)
case "plugins":
err = plugins.List(cfg)
case "sync":
if len(args) == 0 || args[0] == "--safe" || args[0] == "safe" {
err = sync.Safe(cfg)
} else {
err = fmt.Errorf("only sync --safe is implemented in Go kernel")
}
case "verify":
err = verify(cfg)
default:
err = fmt.Errorf("unknown command: %s", cmd)
}

if err != nil {
fmt.Fprintln(os.Stderr, "ERROR:", err)
os.Exit(1)
}
}

func looksLikeExecutablePath(s string) bool {
return len(s) > 0 && (s[0] == '/' || s == "aiftd" || s == "./aiftd" || s == "bin/aiftd")
}

func help() {
fmt.Println("AIFT-OS Federation Control Plane")
fmt.Println()
fmt.Println("Commands:")
fmt.Println("  help")
fmt.Println("  doctor")
fmt.Println("  status")
fmt.Println("  manifest")
fmt.Println("  registry")
fmt.Println("  dashboard")
fmt.Println("  deps")
fmt.Println("  plugins")
fmt.Println("  sync --safe")
fmt.Println("  verify")
}

func verify(cfg config.Config) error {
if err := doctor.Run(cfg); err != nil {
return err
}
if err := manifests.EnsureAll(cfg); err != nil {
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
fmt.Println("OK: federation verified")
return nil
}

func status(cfg config.Config) error {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return err
}

fmt.Printf("%-32s %-12s %-8s %s\n", "REPOSITORY", "BRANCH", "STATE", "REMOTE")
for _, repo := range repos {
state := "clean"
if gitx.Dirty(repo.Path) {
state = "dirty"
}
fmt.Printf("%-32s %-12s %-8s %s\n", repo.Name, gitx.Branch(repo.Path), state, gitx.Remote(repo.Path))
}

return nil
}
GO

echo "== Fix launchers =="
cat > aift-os.sh <<'SH'
#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

if [ ! -x "$BIN" ]; then
  go build -o "$BIN" ./cmd/aift
  chmod +x "$BIN"
fi

exec "$BIN" "$@"
SH

cat > "$ROOT/aift" <<'SH'
#!/usr/bin/env sh
set -eu

exec "$HOME/AIFT/AIFT-OS/aift-os.sh" "$@"
SH

chmod +x aift-os.sh "$ROOT/aift"

echo "== Remove obsolete wrapper and local binary from git =="
rm -f bin/aift bin/aift.sh bin/aift.sh.bak
git rm -f --cached bin/aiftd 2>/dev/null || true

printf '%s\n' 'bin/aiftd' >> .gitignore
sort -u .gitignore -o .gitignore

echo "== Remove temporary repair scripts =="
rm -f \
  bootstrap-go-kernel.sh \
  fix-aiftd-launcher.sh \
  fix-aift-os-kernel.sh \
  fix-binary-tracking.sh \
  fix-launcher-bug.sh \
  inspect-aift-os.sh \
  phase-4-aift-os.sh \
  phase-5-structure.sh \
  reset-aift-os-v2-launcher.sh

echo "== Build and verify =="
go clean -cache
gofmt -w cmd internal
go test ./...
rm -f "$BIN"
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$BIN" help >/dev/null
"$BIN" doctor >/dev/null
"$BIN" status >/dev/null
"$ROOT/aift" help >/dev/null
"$ROOT/aift" doctor >/dev/null
"$ROOT/aift" status >/dev/null
"$ROOT/aift" manifest >/dev/null
"$ROOT/aift" registry >/dev/null
"$ROOT/aift" dashboard >/dev/null
"$ROOT/aift" deps >/dev/null
"$ROOT/aift" plugins >/dev/null
"$ROOT/aift" sync --safe >/dev/null
"$ROOT/aift" verify >/dev/null

echo "== Write smoke test =="
cat > tests/v2-smoke.sh <<'SH'
#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$BIN" help >/dev/null
"$BIN" doctor >/dev/null
"$ROOT/aift" help >/dev/null
"$ROOT/aift" doctor >/dev/null
"$ROOT/aift" verify >/dev/null

echo "OK: AIFT-OS v2 smoke passed"
SH
chmod +x tests/v2-smoke.sh
sh tests/v2-smoke.sh

echo "== Commit and push =="
git add .
if git diff --cached --quiet; then
  echo "Nothing staged to commit."
else
  git commit -m "Stabilize AIFT-OS v2 CLI and launcher"
fi

git push origin main

echo
echo "DONE."
echo "Try:"
echo "  ~/AIFT/aift help"
echo "  ~/AIFT/aift doctor"
echo "  ~/AIFT/aift verify"
echo "  git status"
