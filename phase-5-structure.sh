#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
export AIFT_ROOT="$ROOT"
export AIFT_OS_HOME="$OS"

cd "$OS" || exit 1

mkdir -p \
  app \
  app/cli \
  app/core \
  app/providers \
  app/plugins \
  app/reports \
  app/tests \
  app/schemas \
  app/templates \
  commands \
  runtime \
  providers \
  plugins \
  config \
  docs \
  tests \
  ci \
  tools \
  examples \
  logs \
  var \
  registry \
  reports \
  intelligence \
  manifests

touch \
  app/.gitkeep \
  app/cli/.gitkeep \
  app/core/.gitkeep \
  app/providers/.gitkeep \
  app/plugins/.gitkeep \
  app/reports/.gitkeep \
  app/tests/.gitkeep \
  app/schemas/.gitkeep \
  app/templates/.gitkeep \
  tests/.gitkeep \
  ci/.gitkeep \
  tools/.gitkeep \
  plugins/.gitkeep

cat > docs/STRUCTURE.md <<'DOC'
# AIFT-OS Structure

AIFT-OS is the Federation Control Plane.

It orchestrates sovereign repositories under `~/AIFT` without absorbing them.

## Main folders

- `app/` — future real application source
- `app/cli/` — command-line interface layer
- `app/core/` — control-plane core logic
- `app/providers/` — provider interfaces and adapters
- `app/plugins/` — plugin discovery and execution logic
- `app/reports/` — report generation logic
- `app/tests/` — application-level tests
- `commands/` — current shell command modules
- `runtime/` — current shell runtime libraries
- `providers/` — current shell provider adapters
- `plugins/` — built-in AIFT-OS plugins
- `config/` — local/default configuration
- `schemas/` — JSON schemas
- `tests/` — smoke tests and regression tests
- `ci/` — CI scripts and workflow helpers
- `tools/` — developer utilities
- `registry/` — generated federation registry
- `reports/` — generated federation reports
- `docs/` — documentation
- `logs/` — local logs
- `var/` — local runtime state
DOC

cat > Makefile <<'MAKE'
.PHONY: help doctor status verify dashboard deps plugins safe-sync

help:
@./aift-os.sh help

doctor:
@./aift-os.sh doctor

status:
@./aift-os.sh status

verify:
@./aift-os.sh verify

dashboard:
@./aift-os.sh dashboard

deps:
@./aift-os.sh deps

plugins:
@./aift-os.sh plugins

safe-sync:
@./aift-os.sh sync --safe
MAKE

cat > tests/smoke.sh <<'TEST'
#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."

./aift-os.sh doctor
./aift-os.sh status >/dev/null
./aift-os.sh registry >/dev/null
./aift-os.sh graph >/dev/null
./aift-os.sh deps >/dev/null

echo "OK: smoke tests passed"
TEST

chmod +x tests/smoke.sh

git add .
if git diff --cached --quiet; then
  echo "Nothing new to commit."
else
  git commit -m "Add maintainable AIFT-OS project structure"
fi

git push origin main

echo "DONE."
echo "Now run:"
echo "  make doctor"
echo "  make status"
echo "  sh tests/smoke.sh"
