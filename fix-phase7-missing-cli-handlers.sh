#!/usr/bin/env bash
set -euo pipefail

cat >> /tmp/aift-handlers.go <<'GO'

func runCapabilities(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
 capabilities.Scan(cfg)
}
if args[0] == "list" || args[0] == "report" || args[0] == "info" {
 capabilities.Report(cfg)
}
return fmt.Errorf("usage: aift capabilities scan|list|info|report")
}

func runIntelligence(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
 intelligence.Scan(cfg)
}
if args[0] == "report" || args[0] == "repo" || args[0] == "roadmap" {
 intelligence.Report(cfg)
}
return fmt.Errorf("usage: aift intelligence scan|report|repo|roadmap")
}

func runManual(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
 manual.Scan(cfg)
}
if args[0] == "init-all" {
 manual.InitAll(cfg)
}
if args[0] == "report" || args[0] == "repo" {
 manual.Report(cfg)
}
return fmt.Errorf("usage: aift manual init-all|scan|report|repo")
}

func runMesh(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
 eventmesh.Scan(cfg)
}
if args[0] == "report" || args[0] == "topics" || args[0] == "subscribers" || args[0] == "tail" || args[0] == "replay" {
 eventmesh.Report(cfg)
}
if args[0] == "publish" {
 events.Emit(cfg, "mesh.publish", "mesh", "mesh publish requested", map[string]string{"status": "planned"})
}
if args[0] == "init-all" {
 eventmesh.Scan(cfg)
}
return fmt.Errorf("usage: aift mesh init-all|scan|topics|subscribers|publish|replay|tail|report")
}

func runServiceContracts(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
 servicecontracts.Scan(cfg)
}
if args[0] == "list" || args[0] == "repo" || args[0] == "report" {
 servicecontracts.Report(cfg)
}
if args[0] == "init-all" {
 servicecontracts.Scan(cfg)
}
return fmt.Errorf("usage: aift service-contracts init-all|scan|list|repo|report")
}

func runPlanner(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "build" {
 planner.Build(cfg)
}
if args[0] == "summary" || args[0] == "repo" || args[0] == "ready" || args[0] == "blocked" || args[0] == "report" {
 planner.Report(cfg)
}
return fmt.Errorf("usage: aift plan build|summary|repo|ready|blocked|report")
}

func runModules(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
 registry.Generate(cfg)
}
if args[0] == "init-all" || args[0] == "list" || args[0] == "repo" || args[0] == "report" {
 registry.Generate(cfg)
}
return fmt.Errorf("usage: aift modules init-all|scan|list|repo|report")
}

func runKernelRegistry(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" || args[0] == "list" || args[0] == "object" || args[0] == "report" {
 registry.Generate(cfg)
}
return fmt.Errorf("usage: aift kernel-registry scan|list|object|report")
}

func runDiscovery(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" || args[0] == "list" || args[0] == "object" || args[0] == "report" {
 graph.Build(cfg)
}
return fmt.Errorf("usage: aift discovery scan|list|object|report")
}

func runEventBus(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "list" || args[0] == "replay" || args[0] == "report" {
 eventmesh.Report(cfg)
}
if args[0] == "publish" {
 events.Emit(cfg, "event-bus.publish", "event-bus", "event bus publish requested", map[string]string{"status": "planned"})
}
return fmt.Errorf("usage: aift event-bus publish|list|replay|report")
}
GO

python3 <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
text = p.read_text()
handlers = Path("/tmp/aift-handlers.go").read_text()

missing = [
    "func runCapabilities(",
    "func runIntelligence(",
    "func runManual(",
    "func runMesh(",
    "func runServiceContracts(",
    "func runPlanner(",
    "func runModules(",
    "func runKernelRegistry(",
    "func runDiscovery(",
    "func runEventBus(",
]

if all(m in text for m in missing):
    print("All handlers already exist.")
else:
    text = text.rstrip() + "\n\n" + handlers
    p.write_text(text)
    print("Added missing CLI handlers.")
PY

gofmt -w cmd/aift/main.go

echo "== show handler references =="
grep -n "runCapabilities\|runIntelligence\|runManual\|runMesh\|runServiceContracts\|runPlanner\|runModules\|runKernelRegistry\|runDiscovery\|runEventBus" cmd/aift/main.go

echo "== run checks =="
go test ./...
pnpm run typecheck
pnpm test

git status --short
git add cmd/aift/main.go
git commit -m "fix: wire phase7 CLI handlers to real packages" || true
git push --force-with-lease origin phase7-doctor-git-housekeeping
