#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

export AIFT_ROOT="$ROOT"
export AIFT_OS_HOME="$OS"

cd "$OS" || exit 1

echo "== AIFT-OS Phase 6: core services =="

mkdir -p \
  bin \
  docs \
  install \
  scripts \
  tests \
  registry \
  reports \
  logs \
  var/events \
  internal/version \
  internal/events \
  internal/providers \
  internal/scheduler \
  internal/api \
  internal/daemon \
  internal/runtime

echo "== Keep compiled binary local-only =="
rm -f bin/aift bin/aift.sh bin/aift.sh.bak
git rm -f --cached bin/aiftd 2>/dev/null || true
rm -f bin/aiftd
printf '%s\n' 'bin/aiftd' >> .gitignore
sort -u .gitignore -o .gitignore

echo "== Launcher =="
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

echo "== Version package =="
cat > internal/version/version.go <<'GO'
package version

const (
Name    = "AIFT-OS"
Version = "0.3.0-services"
Role    = "Federation Control Plane"
)
GO

echo "== Events package =="
cat > internal/events/events.go <<'GO'
package events

import (
"encoding/json"
"fmt"
"os"
"path/filepath"
"strings"
"time"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type Event struct {
Time    string            `json:"time"`
Type    string            `json:"type"`
Source  string            `json:"source"`
Message string            `json:"message"`
Data    map[string]string `json:"data,omitempty"`
}

func Emit(cfg config.Config, eventType, source, message string, data map[string]string) error {
event := Event{
Time:    time.Now().Format(time.RFC3339),
Type:    eventType,
Source:  source,
Message: message,
Data:    data,
}

dir := filepath.Join(cfg.OSHome, "var", "events")
if err := os.MkdirAll(dir, 0755); err != nil {
return err
}

path := filepath.Join(dir, "events.jsonl")
f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
if err != nil {
return err
}
defer f.Close()

b, err := json.Marshal(event)
if err != nil {
return err
}

_, err = fmt.Fprintln(f, string(b))
return err
}

func Tail(cfg config.Config, limit int) error {
path := filepath.Join(cfg.OSHome, "var", "events", "events.jsonl")
data, err := os.ReadFile(path)
if err != nil {
fmt.Println("No events yet.")
return nil
}

lines := strings.Split(strings.TrimSpace(string(data)), "\n")
if len(lines) == 1 && lines[0] == "" {
fmt.Println("No events yet.")
return nil
}

if limit > 0 && len(lines) > limit {
lines = lines[len(lines)-limit:]
}

for _, line := range lines {
fmt.Println(line)
}

return nil
}
GO

echo "== Providers package =="
cat > internal/providers/providers.go <<'GO'
package providers

import (
"encoding/json"
"fmt"
"os"
"path/filepath"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type Provider struct {
Name        string `json:"name"`
Type        string `json:"type"`
Status      string `json:"status"`
Description string `json:"description"`
}

func Defaults() []Provider {
return []Provider{
{Name: "local-git", Type: "git", Status: "enabled", Description: "Local Git repository provider"},
{Name: "github", Type: "git-host", Status: "configured-by-repo-remotes", Description: "GitHub remotes discovered from sovereign repositories"},
{Name: "aift-forge", Type: "forge", Status: "planned", Description: "Future local-first federation forge provider"},
{Name: "ollama", Type: "ai", Status: "planned", Description: "Local Ollama AI runtime provider"},
{Name: "llamacpp", Type: "ai", Status: "planned", Description: "Local llama.cpp provider"},
{Name: "vllm", Type: "ai", Status: "planned", Description: "Local/network vLLM provider"},
{Name: "openai-compatible", Type: "ai", Status: "disabled-by-default", Description: "OpenAI-compatible endpoint provider"},
}
}

func WriteRegistry(cfg config.Config) error {
out := filepath.Join(cfg.OSHome, "registry", "providers.json")
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

fmt.Printf("%-22s %-14s %-24s %s\n", "PROVIDER", "TYPE", "STATUS", "DESCRIPTION")
for _, p := range Defaults() {
fmt.Printf("%-22s %-14s %-24s %s\n", p.Name, p.Type, p.Status, p.Description)
}

return nil
}
GO

echo "== Scheduler package =="
cat > internal/scheduler/scheduler.go <<'GO'
package scheduler

import (
"time"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports"
)

type Scheduler struct {
Config config.Config
}

func New(cfg config.Config) Scheduler {
return Scheduler{Config: cfg}
}

func (s Scheduler) RunOnce() error {
if err := events.Emit(s.Config, "scheduler.tick", "scheduler", "scheduler tick started", nil); err != nil {
return err
}

if err := registry.Generate(s.Config); err != nil {
return err
}

if err := reports.Dashboard(s.Config); err != nil {
return err
}

if err := reports.Deps(s.Config); err != nil {
return err
}

return events.Emit(s.Config, "scheduler.tick.complete", "scheduler", "scheduler tick completed", map[string]string{
"interval": "manual",
})
}

func (s Scheduler) Loop(interval time.Duration) error {
ticker := time.NewTicker(interval)
defer ticker.Stop()

if err := s.RunOnce(); err != nil {
return err
}

for range ticker.C {
if err := s.RunOnce(); err != nil {
_ = events.Emit(s.Config, "scheduler.error", "scheduler", err.Error(), nil)
}
}

return nil
}
GO

echo "== API package =="
cat > internal/api/api.go <<'GO'
package api

import (
"encoding/json"
"fmt"
"net/http"
"path/filepath"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/version"
)

type Server struct {
Config config.Config
Addr   string
}

func New(cfg config.Config, addr string) Server {
return Server{Config: cfg, Addr: addr}
}

func (s Server) Serve() error {
mux := http.NewServeMux()

mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
_ = json.NewEncoder(w).Encode(map[string]string{
"status":  "ok",
"name":    version.Name,
"version": version.Version,
})
})

mux.HandleFunc("/registry/repos", func(w http.ResponseWriter, r *http.Request) {
http.ServeFile(w, r, filepath.Join(s.Config.OSHome, "registry", "repos.json"))
})

mux.HandleFunc("/registry/providers", func(w http.ResponseWriter, r *http.Request) {
http.ServeFile(w, r, filepath.Join(s.Config.OSHome, "registry", "providers.json"))
})

mux.HandleFunc("/reports/dashboard", func(w http.ResponseWriter, r *http.Request) {
http.ServeFile(w, r, filepath.Join(s.Config.OSHome, "reports", "dashboard.md"))
})

fmt.Println("AIFT-OS API listening on", s.Addr)
return http.ListenAndServe(s.Addr, mux)
}
GO

echo "== Runtime package =="
cat > internal/runtime/runtime.go <<'GO'
package runtime

import (
"fmt"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/providers"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports"
)

func StartOnce(cfg config.Config) error {
if err := events.Emit(cfg, "runtime.start", "runtime", "runtime one-shot start", nil); err != nil {
return err
}

if err := providers.WriteRegistry(cfg); err != nil {
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

fmt.Println("AIFT-OS runtime completed one-shot start")
return events.Emit(cfg, "runtime.complete", "runtime", "runtime one-shot complete", nil)
}
GO

echo "== Daemon package =="
cat > internal/daemon/daemon.go <<'GO'
package daemon

import (
"fmt"
"time"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/api"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/providers"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/scheduler"
)

func Start(cfg config.Config, addr string) error {
if err := events.Emit(cfg, "daemon.start", "daemon", "AIFT-OS daemon starting", map[string]string{"addr": addr}); err != nil {
return err
}

if err := providers.WriteRegistry(cfg); err != nil {
return err
}

s := scheduler.New(cfg)
if err := s.RunOnce(); err != nil {
return err
}

go func() {
_ = s.Loop(5 * time.Minute)
}()

fmt.Println("AIFT-OS daemon started")
fmt.Println("API:", addr)
fmt.Println("Press CTRL+C to stop.")

return api.New(cfg, addr).Serve()
}
GO

echo "== CLI main =="
cat > cmd/aift/main.go <<'GO'
package main

import (
"fmt"
"os"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/api"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/daemon"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/doctor"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/gitx"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/plugins"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/providers"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/runtime"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/sync"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/version"
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
case "version":
fmt.Printf("%s %s — %s\n", version.Name, version.Version, version.Role)
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
case "providers":
err = providers.List(cfg)
case "events":
err = events.Tail(cfg, 25)
case "start":
err = runtime.StartOnce(cfg)
case "serve":
addr := ":8787"
if len(args) > 0 {
addr = args[0]
}
err = api.New(cfg, addr).Serve()
case "daemon":
addr := ":8787"
if len(args) > 0 {
addr = args[0]
}
err = daemon.Start(cfg, addr)
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
fmt.Println("  version")
fmt.Println("  doctor")
fmt.Println("  status")
fmt.Println("  manifest")
fmt.Println("  registry")
fmt.Println("  dashboard")
fmt.Println("  deps")
fmt.Println("  plugins")
fmt.Println("  providers")
fmt.Println("  events")
fmt.Println("  start")
fmt.Println("  serve [:8787]")
fmt.Println("  daemon [:8787]")
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
if err := providers.WriteRegistry(cfg); err != nil {
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
if err := events.Emit(cfg, "verify.complete", "verify", "federation verified", nil); err != nil {
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

echo "== Tests =="
cat > tests/services-smoke.sh <<'SH'
#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" version >/dev/null
"$ROOT/aift" providers >/dev/null
"$ROOT/aift" events >/dev/null
"$ROOT/aift" start >/dev/null
"$ROOT/aift" verify >/dev/null

echo "OK: AIFT-OS services smoke passed"
SH

chmod +x tests/services-smoke.sh

echo "== Docs =="
cat > docs/PHASE-6-SERVICES.md <<'DOC'
# Phase 6: Core Services

AIFT-OS now includes the first real operating-system service layer.

Implemented:

- event log
- scheduler
- provider registry
- runtime one-shot start
- daemon entry point
- local HTTP API skeleton
- version command
- events command
- providers command

Commands:

- aift version
- aift start
- aift events
- aift providers
- aift serve :8787
- aift daemon :8787

The daemon currently runs the scheduler and local API. This is the foundation for future federation services.
DOC

echo "== Build/test =="
go clean -cache
gofmt -w cmd internal
go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" version
"$ROOT/aift" start
"$ROOT/aift" events
sh tests/services-smoke.sh

echo "== Commit/push =="
git add .
if git diff --cached --quiet; then
  echo "Nothing staged to commit."
else
  git commit -m "Add AIFT-OS core service layer"
fi

git push origin main

echo
echo "DONE."
echo "Try:"
echo "  ~/AIFT/aift version"
echo "  ~/AIFT/aift providers"
echo "  ~/AIFT/aift events"
echo "  ~/AIFT/aift serve :8787"
