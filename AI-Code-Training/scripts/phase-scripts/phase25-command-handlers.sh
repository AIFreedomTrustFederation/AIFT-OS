#!/usr/bin/env bash
set -Eeuo pipefail

cd ~/AIFT/AIFT-OS

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
BIN="bin/aiftd"
REPORT="reports/command-handlers-$STAMP.md"

mkdir -p cmd/aift docs tests schemas reports AI-Code-Training/scripts/phase-scripts bin

echo "== Phase 25: CLI command handler stabilization =="

echo "== Remove direct subsystem imports from main.go when handlers own them =="
python3 - <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text()

imports_to_remove = [
    "github.com/AIFreedomTrustFederation/AIFT-OS/internal/capabilityregistry",
    "github.com/AIFreedomTrustFederation/AIFT-OS/internal/discoveryengine",
    "github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernelregistry",
    "github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernelruntime",
    "github.com/AIFreedomTrustFederation/AIFT-OS/internal/eventbus",
    "github.com/AIFreedomTrustFederation/AIFT-OS/internal/patchengine",
    "github.com/AIFreedomTrustFederation/AIFT-OS/internal/modules",
]

for imp in imports_to_remove:
    s = s.replace(f'\n\t"{imp}"', "")

p.write_text(s)
PY

echo "== Write subsystem command handlers =="

cat > cmd/aift/capabilities.go <<'GO'
package main

import (
"fmt"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/capabilityregistry"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func runCapabilities(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
return capabilityregistry.Scan(cfg)
}

switch args[0] {
case "scan":
return capabilityregistry.Scan(cfg)
case "list":
return capabilityregistry.List(cfg)
case "info":
if len(args) < 2 {
return fmt.Errorf("usage: aift capabilities info <id-or-name>")
}
return capabilityregistry.Info(cfg, args[1])
case "report":
return capabilityregistry.Report(cfg)
default:
return fmt.Errorf("usage: aift capabilities scan|list|info|report")
}
}
GO

cat > cmd/aift/discovery_command.go <<'GO'
package main

import (
"fmt"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/discoveryengine"
)

func runDiscovery(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
return discoveryengine.Scan(cfg)
}

switch args[0] {
case "scan":
return discoveryengine.Scan(cfg)
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
GO

cat > cmd/aift/kernel_registry_command.go <<'GO'
package main

import (
"fmt"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernelregistry"
)

func runKernelRegistry(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
return kernelregistry.Scan(cfg)
}

switch args[0] {
case "scan":
return kernelregistry.Scan(cfg)
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
GO

cat > cmd/aift/kernel_runtime_command.go <<'GO'
package main

import (
"fmt"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernelruntime"
)

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
GO

cat > cmd/aift/event_bus_command.go <<'GO'
package main

import (
"fmt"
"strings"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/eventbus"
)

func runEventBus(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "list" {
return eventbus.List(cfg)
}

switch args[0] {
case "publish":
if len(args) < 3 {
return fmt.Errorf("usage: aift event-bus publish <topic> <message> [key=value...]")
}
payload := map[string]string{}
for _, item := range args[3:] {
parts := strings.SplitN(item, "=", 2)
if len(parts) == 2 {
payload[parts[0]] = parts[1]
}
}
return eventbus.Publish(cfg, args[1], "manual", "aiftd", args[2], payload)
case "list":
return eventbus.List(cfg)
case "replay":
topic := ""
if len(args) > 1 {
topic = args[1]
}
return eventbus.Replay(cfg, topic)
case "report":
return eventbus.Report(cfg)
default:
return fmt.Errorf("usage: aift event-bus publish|list|replay|report")
}
}
GO

cat > cmd/aift/patch_engine_command.go <<'GO'
package main

import (
"fmt"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/patchengine"
)

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
GO

cat > cmd/aift/modules_command.go <<'GO'
package main

import (
"fmt"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/modules"
)

func runModules(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
return modules.Scan(cfg)
}

switch args[0] {
case "init-all":
return modules.InitAll(cfg)
case "scan":
return modules.Scan(cfg)
case "list":
return modules.List(cfg)
case "repo":
if len(args) < 2 {
return fmt.Errorf("usage: aift modules repo <repo>")
}
return modules.Repo(cfg, args[1])
case "report":
return modules.Report(cfg)
default:
return fmt.Errorf("usage: aift modules init-all|scan|list|repo|report")
}
}
GO

cat > cmd/aift/legacy_command_stubs.go <<'GO'
package main

import (
"fmt"
"strings"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func runIntelligence(cfg config.Config, args []string) error {
return plannedCommand("intelligence", args)
}

func runManual(cfg config.Config, args []string) error {
return plannedCommand("manual", args)
}

func runMesh(cfg config.Config, args []string) error {
return plannedCommand("mesh", args)
}

func runServiceContracts(cfg config.Config, args []string) error {
return plannedCommand("service-contracts", args)
}

func runPlanner(cfg config.Config, args []string) error {
return plannedCommand("planner", args)
}

func plannedCommand(name string, args []string) error {
detail := strings.Join(args, " ")
if detail == "" {
detail = "no subcommand"
}
return fmt.Errorf("%s command is planned but not active yet: %s", name, detail)
}
GO

echo "== Write docs =="
cat > docs/CLI-COMMAND-REGISTRY.md <<'DOC'
# AIFT CLI Command Registry

The AIFT CLI is being refactored into small Unix-style command handlers.

`cmd/aift/main.go` is the dispatcher.

Subsystems own their command files:

- `capabilities.go`
- `discovery_command.go`
- `kernel_registry_command.go`
- `kernel_runtime_command.go`
- `event_bus_command.go`
- `patch_engine_command.go`
- `modules_command.go`
- `legacy_command_stubs.go`

## Rule

A subsystem import belongs in its subsystem command file, not in `main.go`, unless `main.go` directly uses that subsystem.

## Planned handlers

Some older command routes are preserved as truthful planned stubs so the CLI compiles without pretending functionality exists.

A planned stub must return a clear error and must not simulate functionality.
DOC

cat > tests/command-registry-smoke.sh <<'SH'
#!/usr/bin/env sh
set -eu

go test ./...
go build -o bin/aiftd ./cmd/aift

bin/aiftd help >/dev/null
bin/aiftd discovery list >/dev/null
bin/aiftd kernel-registry list >/dev/null
bin/aiftd event-bus list >/dev/null
bin/aiftd patch-engine inspect >/dev/null
bin/aiftd capabilities list >/dev/null
bin/aiftd modules list >/dev/null

echo "OK: command registry smoke passed"
SH
chmod +x tests/command-registry-smoke.sh

echo "== Verify no duplicate handlers =="
for fn in runCapabilities runDiscovery runKernelRegistry runKernelRuntime runEventBus runPatchEngine runModules runIntelligence runManual runMesh runServiceContracts runPlanner; do
  count="$(grep -R "func $fn" -n cmd/aift | wc -l | tr -d ' ')"
  if [ "$count" != "1" ]; then
    echo "ERROR: expected exactly one $fn, found $count"
    grep -R "func $fn" -n cmd/aift || true
    exit 1
  fi
done

echo "== Verify =="
gofmt -w cmd/aift
go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"
sh tests/command-registry-smoke.sh

cat > "$REPORT" <<EOF
# CLI Command Registry Refactor Report

Generated: $STAMP

Passed:

- exactly one handler per dispatcher route
- gofmt
- go test ./...
- go build ./cmd/aift
- tests/command-registry-smoke.sh

Truth rule:

Legacy routes that are not implemented return planned errors instead of pretending to work.
EOF

cp "$0" AI-Code-Training/scripts/phase-scripts/phase25-command-handlers.sh 2>/dev/null || true

echo "== Stage source files only =="
git add \
  cmd/aift/main.go \
  cmd/aift/capabilities.go \
  cmd/aift/discovery_command.go \
  cmd/aift/kernel_registry_command.go \
  cmd/aift/kernel_runtime_command.go \
  cmd/aift/event_bus_command.go \
  cmd/aift/patch_engine_command.go \
  cmd/aift/modules_command.go \
  cmd/aift/legacy_command_stubs.go \
  tests/command-registry-smoke.sh \
  docs/CLI-COMMAND-REGISTRY.md \
  AI-Code-Training/scripts/phase-scripts/phase25-command-handlers.sh

echo "== Commit and push =="
if git diff --cached --quiet; then
  echo "Nothing staged."
else
  git commit -m "Refactor CLI command handlers"
  git push origin main
fi

echo "DONE"
