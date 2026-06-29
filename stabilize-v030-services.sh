#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"
TAG="v0.3.0-services"
TRAINING_DIR="AI-Code-Training"

export AIFT_ROOT="$ROOT"
export AIFT_OS_HOME="$OS"

cd "$OS" || exit 1

echo "== Stabilize AIFT-OS $TAG with AI Code Training archive =="

mkdir -p \
  "$TRAINING_DIR/scripts/migration" \
  "$TRAINING_DIR/scripts/patches" \
  "$TRAINING_DIR/scripts/phase-scripts" \
  "$TRAINING_DIR/scripts/experiments" \
  "$TRAINING_DIR/successful-patches" \
  "$TRAINING_DIR/failures" \
  "$TRAINING_DIR/architecture-evolution" \
  "$TRAINING_DIR/lessons-learned" \
  "$TRAINING_DIR/metadata" \
  docs tests install scripts registry reports logs var/events bin

archive_script() {
  src="$1"
  dest="$2"

  if [ -f "$src" ]; then
    mkdir -p "$(dirname "$dest")"

    if git ls-files --error-unmatch "$src" >/dev/null 2>&1; then
      git mv "$src" "$dest"
    else
      mv "$src" "$dest"
      git add "$dest" 2>/dev/null || true
    fi

    echo "archived: $src -> $dest"
  fi
}

echo "== Archive migration/repair scripts for AI training =="

archive_script bootstrap-go-kernel.sh "$TRAINING_DIR/scripts/migration/bootstrap-go-kernel.sh"
archive_script finish-phase-2-3.sh "$TRAINING_DIR/scripts/phase-scripts/finish-phase-2-3.sh"
archive_script phase-4-aift-os.sh "$TRAINING_DIR/scripts/phase-scripts/phase-4-aift-os.sh"
archive_script phase-5-structure.sh "$TRAINING_DIR/scripts/phase-scripts/phase-5-structure.sh"
archive_script phase-6-services.sh "$TRAINING_DIR/scripts/phase-scripts/phase-6-services.sh"
archive_script phase-7-runtime-api.sh "$TRAINING_DIR/scripts/phase-scripts/phase-7-runtime-api.sh"
archive_script phase-7-runtime-api.sh.bak "$TRAINING_DIR/scripts/phase-scripts/phase-7-runtime-api.sh.bak"

archive_script fix-aiftd-launcher.sh "$TRAINING_DIR/scripts/patches/fix-aiftd-launcher.sh"
archive_script fix-aift-os-kernel.sh "$TRAINING_DIR/scripts/patches/fix-aift-os-kernel.sh"
archive_script fix-binary-tracking.sh "$TRAINING_DIR/scripts/patches/fix-binary-tracking.sh"
archive_script fix-launcher-bug.sh "$TRAINING_DIR/scripts/patches/fix-launcher-bug.sh"
archive_script reset-aift-os-v2-launcher.sh "$TRAINING_DIR/scripts/patches/reset-aift-os-v2-launcher.sh"
archive_script surgical-aift-os-v2-fix.sh "$TRAINING_DIR/successful-patches/surgical-aift-os-v2-fix.sh"
archive_script patch-phase7-script.sh "$TRAINING_DIR/successful-patches/patch-phase7-script.sh"

archive_script inspect-aift-os.sh "$TRAINING_DIR/scripts/experiments/inspect-aift-os.sh"

# Preserve any remaining root-level generated scripts that match known migration/fix naming.
for f in ./*.sh; do
  [ -f "$f" ] || continue
  name="$(basename "$f")"

  case "$name" in
    aift-os.sh|stabilize-v030-services.sh)
      ;;
    fix-*.sh|phase-*.sh|reset-*.sh|surgical-*.sh|patch-*.sh|inspect-*.sh|bootstrap-*.sh|finish-*.sh)
      archive_script "$name" "$TRAINING_DIR/scripts/experiments/$name"
      ;;
  esac
done

echo "== Write AI Code Training metadata =="

cat > "$TRAINING_DIR/README.md" <<'DOC'
# AI Code Training

This directory preserves AIFT-OS development history for future AI agents.

It contains migration scripts, failed approaches, successful patches, repair scripts, architecture evolution notes, and lessons learned.

This is not a trash folder. It is a training corpus.

Future AIFT agents should inspect this archive before proposing large refactors so they can understand:

- what worked
- what failed
- which launcher patterns caused recursion
- why compiled binaries are not tracked
- why shell scripts should avoid Bash-only features when run with POSIX `sh`
- how AIFT-OS migrated from shell scripts into a Go control-plane kernel
DOC

cat > "$TRAINING_DIR/lessons-learned/launcher-and-shell-lessons.md" <<'DOC'
# Launcher and Shell Lessons

## Launcher lessons

- Do not track compiled binaries such as `bin/aiftd`.
- Keep the shell launcher thin.
- Use one workspace launcher: `~/AIFT/aift`.
- Use one repo launcher: `aift-os.sh`.
- Use one compiled binary: `bin/aiftd`.
- Avoid recursive wrapper chains.
- Never pass the binary path as a user command argument.

## Shell lessons

- Termux can handle long scripts, but ChatGPT-generated heredocs can become fragile.
- POSIX `sh` does not support Bash brace expansion such as `mkdir -p internal/{api,state}`.
- Use explicit directory lists in portable scripts.
- Prefer small idempotent scripts once the system grows.

## Go migration lessons

- Keep CLI parsing in `cmd/aift`.
- Keep business logic in `internal/*`.
- Generated outputs belong in `registry/`, `reports/`, `logs/`, and `var/`.
- Runtime state belongs in `var/`.
DOC

cat > "$TRAINING_DIR/architecture-evolution/aift-os-evolution.md" <<'DOC'
# AIFT-OS Architecture Evolution

AIFT-OS began as a shell-based federation control plane.

The project evolved through these stages:

1. Directory cleanup and control-plane layout.
2. Shell command dispatcher.
3. Federation registry and reports.
4. Plugin and manifest architecture.
5. Go kernel introduction.
6. Launcher stabilization.
7. Runtime service layer.
8. Internal API and supervisor foundation.

The current direction is a Go-based federation operating system with shell scripts only for bootstrap, install, tests, and archival migration utilities.
DOC

cat > "$TRAINING_DIR/metadata/index.json" <<'JSON'
{
  "name": "AI Code Training",
  "purpose": "Preserve AIFT-OS development scripts and architecture evolution for future AI agents.",
  "status": "active training corpus",
  "categories": [
    "migration scripts",
    "phase scripts",
    "patch scripts",
    "successful patches",
    "failed approaches",
    "architecture evolution",
    "lessons learned"
  ],
  "policy": {
    "delete_old_scripts": false,
    "archive_old_scripts": true,
    "compiled_binaries_tracked": false,
    "shell_scripts_are_training_data": true
  }
}
JSON

cat > "$TRAINING_DIR/metadata/timeline.json" <<'JSON'
[
  {
    "phase": "shell-control-plane",
    "summary": "Initial AIFT-OS control-plane scripts and repo structure."
  },
  {
    "phase": "go-kernel",
    "summary": "Migrated command kernel to Go."
  },
  {
    "phase": "launcher-stabilization",
    "summary": "Separated workspace launcher, repo launcher, and compiled Go binary."
  },
  {
    "phase": "services-runtime",
    "summary": "Added event log, providers, runtime, scheduler, supervisor, services, and API foundation."
  }
]
JSON

echo "== Keep binary local only =="
git rm -f --cached bin/aiftd 2>/dev/null || true
rm -f bin/aiftd
printf '%s\n' 'bin/aiftd' >> .gitignore
sort -u .gitignore -o .gitignore

echo "== Package boundary docs =="
cat > docs/PACKAGE-BOUNDARIES.md <<'DOC'
# AIFT-OS Package Boundaries

AIFT-OS is the Federation Control Plane. It orchestrates sovereign repositories without absorbing them.

## CLI

- `cmd/aift` owns command parsing and user-facing commands only.
- CLI code should call internal packages, not implement business logic directly.

## Configuration

- `internal/config` resolves workspace paths and runtime settings.
- No package should hard-code `~/AIFT` directly.

## Workspace and Git

- `internal/workspace` discovers repositories.
- `internal/gitx` wraps Git commands.
- Higher packages should not shell out to Git directly.

## Manifests and Registry

- `internal/manifests` owns `.aift/repo.json` manifest creation and validation.
- `internal/registry` generates machine-readable federation registry files.

## Reports

- `internal/reports` generates human-readable reports such as dashboards and dependency graphs.

## Events

- `internal/events` owns append-only runtime event logging.
- Services should emit events for important lifecycle changes.

## Providers

- `internal/providers` owns provider registration and provider registry output.
- Future provider implementations should plug in here.

## Services and Runtime

- `internal/services` describes service state.
- `internal/jobs` owns repeatable runtime jobs.
- `internal/supervisor` coordinates jobs and services.
- `internal/runtime` owns one-shot and loop runtime behavior.
- `internal/daemon` starts long-running runtime services.
- `internal/api` exposes local HTTP endpoints.

## Rules

1. Generated files live in `registry/`, `reports/`, `logs/`, and `var/`.
2. Compiled binaries stay local and must not be tracked.
3. Shell scripts only bootstrap, install, test, inspect, or preserve AI training history.
4. New features should be Go packages first, CLI commands second.
5. Development history scripts belong in `AI-Code-Training/`.
DOC

echo "== Roadmap docs =="
cat > docs/ROADMAP.md <<'DOC'
# AIFT-OS Roadmap

## Completed

- Repository structure
- Go CLI kernel
- Launcher stabilization
- Registry generation
- Manifest generation
- Dashboard and dependency reports
- Event log
- Provider registry
- Runtime state
- Service list
- Supervisor and job runner
- Local API skeleton
- AI Code Training archive

## v0.3.0-services

This release is the first stable runtime baseline.

It includes:

- `aift version`
- `aift doctor`
- `aift status`
- `aift manifest`
- `aift registry`
- `aift dashboard`
- `aift deps`
- `aift providers`
- `aift events`
- `aift services`
- `aift start`
- `aift tick`
- `aift verify`
- `aift daemon :8787`

## Next: Phase 8 Plugin Runtime

Planned commands:

- `aift plugin list`
- `aift plugin run <repo> <command>`
- `aift plugin enable <repo> <command>`
- `aift plugin disable <repo> <command>`

## Later

- Full provider interfaces
- Service manager
- Local UI API
- Forge integration
- Federation node discovery
- Cross-node sync
- AI orchestration
DOC

echo "== Integration test =="
cat > tests/integration.sh <<'SH'
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
"$ROOT/aift" doctor >/dev/null
"$ROOT/aift" status >/dev/null
"$ROOT/aift" manifest >/dev/null
"$ROOT/aift" registry >/dev/null
"$ROOT/aift" dashboard >/dev/null
"$ROOT/aift" deps >/dev/null
"$ROOT/aift" providers >/dev/null
"$ROOT/aift" events >/dev/null
"$ROOT/aift" services >/dev/null
"$ROOT/aift" start >/dev/null
"$ROOT/aift" tick >/dev/null
"$ROOT/aift" verify >/dev/null

test -f registry/repos.json
test -f registry/providers.json
test -f reports/dashboard.md
test -f reports/dependency-graph.md
test -f var/events/events.jsonl

echo "OK: integration test passed"
SH

chmod +x tests/integration.sh

echo "== Build and test =="
gofmt -w cmd internal
go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

sh tests/integration.sh

echo "== Commit stabilization =="
git add .
if git diff --cached --quiet; then
  echo "Nothing staged to commit."
else
  git commit -m "Stabilize AIFT-OS services runtime baseline with AI training archive"
fi

echo "== Push main =="
git push origin main

echo "== Tag baseline =="
if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "Tag $TAG already exists locally."
else
  git tag -a "$TAG" -m "AIFT-OS stable services runtime baseline"
fi

git push origin "$TAG" || true

echo
echo "DONE."
echo "Stable baseline: $TAG"
echo "Training archive: $TRAINING_DIR"
echo "Try:"
echo "  ~/AIFT/aift version"
echo "  ~/AIFT/aift services"
echo "  sh tests/integration.sh"
