#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
export AIFT_ROOT="$ROOT"
export AIFT_OS_HOME="$OS"

cd "$OS" || exit 1

echo "== Patch Phase 7 script for POSIX sh =="

if [ ! -f phase-7-runtime-api.sh ]; then
  echo "ERROR: phase-7-runtime-api.sh not found"
  exit 1
fi

cp phase-7-runtime-api.sh phase-7-runtime-api.sh.bak

python - <<'PY'
from pathlib import Path

p = Path("phase-7-runtime-api.sh")
s = p.read_text()

# Replace Bash-only brace expansion with POSIX-safe explicit directories.
s = s.replace(
"""mkdir -p internal/{runtime,services,api,events,state,supervisor,jobs}""",
"""mkdir -p \\
  internal/runtime \\
  internal/services \\
  internal/api \\
  internal/events \\
  internal/state \\
  internal/supervisor \\
  internal/jobs"""
)

# Make sure all required directories are created even if the exact old line changed.
needle = 'echo "== AIFT-OS Phase 7: Runtime + Internal API =="'
insert = '''echo "== AIFT-OS Phase 7: Runtime + Internal API =="

mkdir -p \\
  internal/runtime \\
  internal/services \\
  internal/api \\
  internal/events \\
  internal/state \\
  internal/supervisor \\
  internal/jobs \\
  docs \\
  tests \\
  registry \\
  reports \\
  logs \\
  var/events \\
  bin
'''
if needle in s and "internal/state \\" not in s:
    s = s.replace(needle, insert)

p.write_text(s)
PY

chmod +x phase-7-runtime-api.sh

echo "== Verify required mkdir lines =="
grep -n "internal/state" phase-7-runtime-api.sh || {
  echo "ERROR: internal/state directory creation still missing"
  exit 1
}

echo "== Run patched Phase 7 =="
sh phase-7-runtime-api.sh

echo "== Commit patched script if needed =="
git add phase-7-runtime-api.sh phase-7-runtime-api.sh.bak docs tests internal cmd Makefile registry reports var .gitignore aift-os.sh 2>/dev/null || git add .

if git diff --cached --quiet; then
  echo "Nothing staged to commit after Phase 7."
else
  git commit -m "Patch Phase 7 runtime API installer"
fi

git push origin main

echo "DONE."
