#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
STAMP="$(date +%Y%m%d-%H%M%S)"
OUT="$OS/reports/introspection/$STAMP"
BUNDLE="$OS/reports/introspection-$STAMP.tar.gz"

mkfile() {
  file="$1"
  dir="$(dirname "$file")"
  mkdir -p "$dir"
}

mkdirp() {
  mkdir -p "$@"
}

copy_if_exists() {
  src="$1"
  dst="$2"
  [ -e "$src" ] || return 0
  mkdir -p "$(dirname "$dst")"
  cp -R "$src" "$dst" 2>/dev/null || true
}

run_logged() {
  label="$1"
  cmd="$2"
  log="$3"
  mkfile "$log"

  {
    echo "LABEL=$label"
    echo "CMD=$cmd"
    echo "DATE=$(date)"
    echo
    sh -lc "$cmd"
  } > "$log" 2>&1 || true
}

discover_repo() {
  path="$1"
  name="$(basename "$path")"
  rdir="$OUT/nodes/local/repositories/$name"

  echo "Discovering repository: $name"

  mkdirp \
    "$rdir/git" \
    "$rdir/contracts" \
    "$rdir/workflows" \
    "$rdir/dependencies" \
    "$rdir/commands" \
    "$rdir/logs" \
    "$rdir/files"

  cd "$path" || return 0

  git status --short > "$rdir/git/status-short.txt" 2>&1 || true
  git status > "$rdir/git/status-full.txt" 2>&1 || true
  git log -10 --oneline > "$rdir/git/log-last-10.txt" 2>&1 || true
  git remote -v > "$rdir/git/remotes.txt" 2>&1 || true
  git branch -vv > "$rdir/git/branches.txt" 2>&1 || true

  find . -maxdepth 4 -type f | sort > "$rdir/files/index.txt" 2>/dev/null || true

  copy_if_exists "$path/.aift" "$rdir/contracts/.aift"
  copy_if_exists "$path/.github/workflows" "$rdir/workflows"
  copy_if_exists "$path/package.json" "$rdir/dependencies/package.json"
  copy_if_exists "$path/package-lock.json" "$rdir/dependencies/package-lock.json"
  copy_if_exists "$path/go.mod" "$rdir/dependencies/go.mod"
  copy_if_exists "$path/go.sum" "$rdir/dependencies/go.sum"
  copy_if_exists "$path/Cargo.toml" "$rdir/dependencies/Cargo.toml"
  copy_if_exists "$path/pyproject.toml" "$rdir/dependencies/pyproject.toml"
  copy_if_exists "$path/requirements.txt" "$rdir/dependencies/requirements.txt"
  copy_if_exists "$path/docs/manual" "$rdir/manual"

  mkfile "$rdir/commands/detected.txt"
  : > "$rdir/commands/detected.txt"

  if [ -f package.json ]; then
    node - <<NODE >> "$rdir/commands/detected.txt" 2>/dev/null || true
const fs = require("fs");
const p = JSON.parse(fs.readFileSync("package.json", "utf8"));
for (const [k,v] of Object.entries(p.scripts || {})) {
  console.log(`npm:${k}=npm run ${k}`);
}
NODE
  fi

  [ -f go.mod ] && echo "go:test=go test ./..." >> "$rdir/commands/detected.txt"
  [ -f go.mod ] && echo "go:build=go build ./..." >> "$rdir/commands/detected.txt"
  [ -f Cargo.toml ] && echo "cargo:test=cargo test" >> "$rdir/commands/detected.txt"
  [ -f Cargo.toml ] && echo "cargo:build=cargo build" >> "$rdir/commands/detected.txt"
  [ -f pyproject.toml ] && echo "python:pytest=python -m pytest" >> "$rdir/commands/detected.txt"

  if [ -f package.json ]; then
    if node -e 'const p=require("./package.json"); process.exit(p.scripts && p.scripts.verify ? 0 : 1)' 2>/dev/null; then
      run_logged "npm-verify" "npm ci && npm run verify" "$rdir/logs/npm-verify.log"
    elif node -e 'const p=require("./package.json"); process.exit(p.scripts && p.scripts.build ? 0 : 1)' 2>/dev/null; then
      run_logged "npm-build" "npm ci && npm run build" "$rdir/logs/npm-build.log"
    fi
  fi

  [ -f go.mod ] && run_logged "go-test" "go test ./..." "$rdir/logs/go-test.log"
  [ -f go.mod ] && run_logged "go-build" "go build ./..." "$rdir/logs/go-build.log"

  mkfile "$rdir/SUMMARY.md"
  {
    echo "# $name"
    echo
    echo "## Git Status"
    echo '```text'
    cat "$rdir/git/status-short.txt" || true
    echo '```'
    echo
    echo "## Detected Commands"
    echo '```text'
    cat "$rdir/commands/detected.txt" || true
    echo '```'
    echo
    echo "## Logs"
    find "$rdir/logs" -type f 2>/dev/null | sort | while read -r log; do
      echo
      echo "### $(basename "$log")"
      echo '```text'
      tail -120 "$log" || true
      echo '```'
    done
  } > "$rdir/SUMMARY.md"
}

mkdirp \
  "$OUT/nodes" \
  "$OUT/modules" \
  "$OUT/contracts" \
  "$OUT/graph" \
  "$OUT/diagnostics" \
  "$OUT/environment"

mkfile "$OUT/environment/environment.txt"
cat > "$OUT/environment/environment.txt" <<ENV
DATE=$(date)
ROOT=$ROOT
OS=$OS
SHELL=${SHELL:-}
PATH=$PATH
NODE=$(command -v node 2>/dev/null || true) $(node -v 2>/dev/null || true)
NPM=$(command -v npm 2>/dev/null || true) $(npm -v 2>/dev/null || true)
GO=$(command -v go 2>/dev/null || true) $(go version 2>/dev/null || true)
GIT=$(git --version 2>/dev/null || true)
UNAME=$(uname -a 2>/dev/null || true)
ENV

copy_if_exists "$OS/registry" "$OUT/modules/aift-os/registry"
copy_if_exists "$OS/reports" "$OUT/modules/aift-os/reports"
copy_if_exists "$OS/schemas" "$OUT/modules/aift-os/schemas"

find "$ROOT" -maxdepth 1 -type d | sort | while read -r repo; do
  [ -d "$repo/.git" ] || continue
  discover_repo "$repo"
done

mkfile "$OUT/MASTER-REPORT.md"
{
  echo "# AIFT Federation Discovery & Introspection Kernel Report"
  echo
  echo "Generated: $(date)"
  echo
  echo "## Environment"
  echo '```text'
  cat "$OUT/environment/environment.txt"
  echo '```'
  echo
  echo "## Discovered Repositories"
  find "$OUT/nodes/local/repositories" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort | while read -r r; do
    echo "- $(basename "$r")"
  done
  echo
  echo "## Repository Summaries"
  find "$OUT/nodes/local/repositories" -name SUMMARY.md 2>/dev/null | sort | while read -r s; do
    echo
    echo "---"
    cat "$s"
  done
} > "$OUT/MASTER-REPORT.md"

cd "$OS" || exit 1
tar -czf "$BUNDLE" -C "$OS/reports/introspection" "$STAMP"

echo
echo "DONE."
echo "Upload this bundle:"
echo "$BUNDLE"
echo
echo "Master report:"
echo "$OUT/MASTER-REPORT.md"
