#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
STAMP="$(date +%Y%m%d-%H%M%S)"
OUT="$OS/reports/diagnostics/$STAMP"
BUNDLE="$OS/reports/diagnostics-$STAMP.tar.gz"

mkdir -p "$OUT"

mkfile() {
  mkdir -p "$(dirname "$1")"
}

write() {
  mkfile "$1"
  cat > "$1"
}

copy_if_exists() {
  src="$1"
  dst="$2"
  [ -e "$src" ] || return 0
  mkdir -p "$(dirname "$dst")"
  cp -R "$src" "$dst" 2>/dev/null || true
}

run_logged() {
  repo="$1"
  label="$2"
  cmd="$3"
  log="$OUT/repos/$repo/logs/$label.log"
  mkfile "$log"

  {
    echo "REPO=$repo"
    echo "LABEL=$label"
    echo "CMD=$cmd"
    echo "DATE=$(date)"
    echo
    sh -lc "$cmd"
  } > "$log" 2>&1
}

detect_commands() {
  repo="$1"
  path="$2"
  cmdfile="$OUT/repos/$repo/detected-commands.txt"
  mkfile "$cmdfile"
  : > "$cmdfile"

  cd "$path" || return 0

  if [ -f package.json ]; then
    node - <<'NODE' >> "$cmdfile" 2>/dev/null || true
const fs = require("fs");
const p = JSON.parse(fs.readFileSync("package.json","utf8"));
const scripts = p.scripts || {};
for (const [k,v] of Object.entries(scripts)) {
  console.log(`npm:${k}=npm run ${k}`);
}
NODE
  fi

  [ -f go.mod ] && echo "go:test=go test ./..." >> "$cmdfile"
  [ -f go.mod ] && echo "go:build=go build ./..." >> "$cmdfile"
  [ -f Cargo.toml ] && echo "cargo:test=cargo test" >> "$cmdfile"
  [ -f Cargo.toml ] && echo "cargo:build=cargo build" >> "$cmdfile"
  [ -f pyproject.toml ] && echo "python:pytest=python -m pytest" >> "$cmdfile"
  [ -f requirements.txt ] && echo "python:pytest=python -m pytest" >> "$cmdfile"

  find .aift/commands -type f -maxdepth 2 2>/dev/null | sort | while read -r f; do
    [ -f "$f" ] || continue
    name="$(basename "$f")"
    echo "aift:$name=sh $f" >> "$cmdfile"
  done
}

run_safe_ci() {
  repo="$1"
  path="$2"
  cd "$path" || return 0

  # Prefer explicit verify/test/build scripts if they exist.
  if [ -f package.json ]; then
    if node -e 'const p=require("./package.json"); process.exit(p.scripts?.verify ? 0 : 1)' 2>/dev/null; then
      run_logged "$repo" "npm-verify" "npm ci && npm run verify" || true
      return 0
    fi
    if node -e 'const p=require("./package.json"); process.exit(p.scripts?.test ? 0 : 1)' 2>/dev/null; then
      run_logged "$repo" "npm-test" "npm ci && npm test" || true
    fi
    if node -e 'const p=require("./package.json"); process.exit(p.scripts?.build ? 0 : 1)' 2>/dev/null; then
      run_logged "$repo" "npm-build" "npm ci && npm run build" || true
    fi
  fi

  [ -f go.mod ] && run_logged "$repo" "go-test" "go test ./..." || true
  [ -f go.mod ] && run_logged "$repo" "go-build" "go build ./..." || true
  [ -f Cargo.toml ] && run_logged "$repo" "cargo-test" "cargo test" || true
  [ -f Cargo.toml ] && run_logged "$repo" "cargo-build" "cargo build" || true
}

diagnose_repo() {
  path="$1"
  repo="$(basename "$path")"

  echo "Diagnosing $repo"

  mkdir -p "$OUT/repos/$repo"

  cd "$path" || return 0

  git status --short > "$OUT/repos/$repo/git-status-short.txt" 2>&1 || true
  git status > "$OUT/repos/$repo/git-status-full.txt" 2>&1 || true
  git log -10 --oneline > "$OUT/repos/$repo/git-log.txt" 2>&1 || true
  git remote -v > "$OUT/repos/$repo/git-remotes.txt" 2>&1 || true
  git branch -vv > "$OUT/repos/$repo/git-branches.txt" 2>&1 || true

  find . -maxdepth 3 -type f | sort > "$OUT/repos/$repo/file-index.txt" 2>/dev/null || true

  copy_if_exists "$path/package.json" "$OUT/repos/$repo/package.json"
  copy_if_exists "$path/package-lock.json" "$OUT/repos/$repo/package-lock.json"
  copy_if_exists "$path/go.mod" "$OUT/repos/$repo/go.mod"
  copy_if_exists "$path/Cargo.toml" "$OUT/repos/$repo/Cargo.toml"
  copy_if_exists "$path/pyproject.toml" "$OUT/repos/$repo/pyproject.toml"
  copy_if_exists "$path/requirements.txt" "$OUT/repos/$repo/requirements.txt"
  copy_if_exists "$path/.github/workflows" "$OUT/repos/$repo/workflows"
  copy_if_exists "$path/.aift" "$OUT/repos/$repo/aift-contracts"

  detect_commands "$repo" "$path"
  run_safe_ci "$repo" "$path"

  {
    echo "# $repo"
    echo
    echo "## Git"
    echo '```text'
    cat "$OUT/repos/$repo/git-status-short.txt"
    echo '```'
    echo
    echo "## Detected Commands"
    echo '```text'
    cat "$OUT/repos/$repo/detected-commands.txt" 2>/dev/null || true
    echo '```'
    echo
    echo "## Logs"
    find "$OUT/repos/$repo/logs" -type f 2>/dev/null | sort | while read -r log; do
      echo
      echo "### $(basename "$log")"
      echo '```text'
      tail -120 "$log"
      echo '```'
    done
  } > "$OUT/repos/$repo/SUMMARY.md"
}

write "$OUT/environment.txt" <<ENV
DATE=$(date)
ROOT=$ROOT
OS=$OS
SHELL=$SHELL
PATH=$PATH
NODE=$(command -v node 2>/dev/null || true) $(node -v 2>/dev/null || true)
NPM=$(command -v npm 2>/dev/null || true) $(npm -v 2>/dev/null || true)
GO=$(command -v go 2>/dev/null || true) $(go version 2>/dev/null || true)
GIT=$(git --version 2>/dev/null || true)
UNAME=$(uname -a 2>/dev/null || true)
ENV

# Include generated AIFT knowledge if it exists.
copy_if_exists "$OS/registry" "$OUT/aift-os/registry"
copy_if_exists "$OS/reports" "$OUT/aift-os/reports"
copy_if_exists "$OS/schemas" "$OUT/aift-os/schemas"

find "$ROOT" -maxdepth 1 -type d | sort | while read -r repo; do
  [ -d "$repo/.git" ] || continue
  diagnose_repo "$repo"
done

{
  echo "# AIFT Federation Diagnostic Bundle"
  echo
  echo "Generated: $(date)"
  echo
  echo "## Environment"
  echo '```text'
  cat "$OUT/environment.txt"
  echo '```'
  echo
  echo "## Repositories"
  find "$OUT/repos" -mindepth 1 -maxdepth 1 -type d | sort | while read -r r; do
    echo "- $(basename "$r")"
  done
  echo
  echo "## Repo Summaries"
  find "$OUT/repos" -name SUMMARY.md | sort | while read -r s; do
    echo
    echo "---"
    cat "$s"
  done
} > "$OUT/MASTER-REPORT.md"

cd "$OS" || exit 1
tar -czf "$BUNDLE" -C "$OS/reports/diagnostics" "$STAMP"

echo
echo "DONE."
echo "Upload this bundle:"
echo "$BUNDLE"
echo
echo "Master report:"
echo "$OUT/MASTER-REPORT.md"
