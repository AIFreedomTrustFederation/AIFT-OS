#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
STAMP="$(date +%Y%m%d-%H%M%S)"
OUT="$OS/reports/ci-truth-repair/$STAMP"
BUNDLE="$OS/reports/ci-truth-repair-$STAMP.tar.gz"

mkdir -p "$OUT"/{logs,status,metadata,workflows,packages,summaries}

repos="
booksmith-ai
capital-city-provisions
tastycutz
BookSmith-Federation-OS
Aether_Coin_biozonecurrency
"

write_file() {
  path="$1"
  shift
  mkdir -p "$(dirname "$path")"
  printf "%s\n" "$@" > "$path"
}

run_repo() {
  repo="$1"
  cmd="$2"
  repo_path="$ROOT/$repo"
  log="$OUT/logs/$repo.log"
  summary="$OUT/summaries/$repo.summary.md"

  mkdir -p "$OUT/workflows/$repo" "$OUT/packages/$repo" "$OUT/status/$repo" "$OUT/metadata/$repo"

  {
    echo "# $repo"
    echo
    echo "- Path: \`$repo_path\`"
    echo "- Command: \`$cmd\`"
    echo "- Started: $(date)"
    echo
  } > "$summary"

  if [ ! -d "$repo_path/.git" ]; then
    echo "MISSING REPO: $repo_path" | tee "$log"
    echo "- Result: MISSING REPO" >> "$summary"
    return 1
  fi

  cd "$repo_path" || return 1

  git status --short > "$OUT/status/$repo/git-status-short.txt" 2>&1 || true
  git status > "$OUT/status/$repo/git-status-full.txt" 2>&1 || true
  git log -5 --oneline > "$OUT/status/$repo/git-log-last-5.txt" 2>&1 || true
  git remote -v > "$OUT/status/$repo/git-remotes.txt" 2>&1 || true
  git branch -vv > "$OUT/status/$repo/git-branches.txt" 2>&1 || true

  find .github/workflows -maxdepth 1 -type f 2>/dev/null | sort > "$OUT/workflows/$repo/workflow-list.txt" || true
  if [ -d .github/workflows ]; then
    cp -R .github/workflows "$OUT/workflows/$repo/files" 2>/dev/null || true
  fi

  [ -f package.json ] && cp package.json "$OUT/packages/$repo/package.json" || true
  [ -f package-lock.json ] && cp package-lock.json "$OUT/packages/$repo/package-lock.json" || true
  [ -f pnpm-lock.yaml ] && cp pnpm-lock.yaml "$OUT/packages/$repo/pnpm-lock.yaml" || true
  [ -f yarn.lock ] && cp yarn.lock "$OUT/packages/$repo/yarn.lock" || true
  [ -f go.mod ] && cp go.mod "$OUT/packages/$repo/go.mod" || true
  [ -f go.sum ] && cp go.sum "$OUT/packages/$repo/go.sum" || true

  {
    echo "REPO=$repo"
    echo "PATH=$repo_path"
    echo "PWD=$(pwd)"
    echo "DATE=$(date)"
    echo "NODE=$(command -v node 2>/dev/null || true) $(node -v 2>/dev/null || true)"
    echo "NPM=$(command -v npm 2>/dev/null || true) $(npm -v 2>/dev/null || true)"
    echo "GO=$(command -v go 2>/dev/null || true) $(go version 2>/dev/null || true)"
    echo "GIT=$(git --version 2>/dev/null || true)"
    echo
    echo "== git status short =="
    git status --short || true
    echo
    echo "== recent commits =="
    git log -5 --oneline || true
    echo
    echo "== workflows =="
    find .github/workflows -maxdepth 1 -type f 2>/dev/null | sort || true
    echo
    echo "== package scripts =="
    if [ -f package.json ]; then
      node -e 'const p=require("./package.json"); console.log(JSON.stringify(p.scripts||{}, null, 2))' || true
    else
      echo "No package.json"
    fi
    echo
    echo "== command =="
    echo "$cmd"
    echo
    echo "== output =="
    sh -lc "$cmd"
  } > "$log" 2>&1

  code=$?

  {
    echo "- Finished: $(date)"
    echo "- Exit code: \`$code\`"
    if [ "$code" -eq 0 ]; then
      echo "- Result: PASS"
    else
      echo "- Result: FAIL"
    fi
    echo
    echo "## Last 80 log lines"
    echo
    echo '```text'
    tail -80 "$log" || true
    echo '```'
  } >> "$summary"

  return "$code"
}

cat > "$OUT/metadata/environment.txt" <<ENV
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

cat > "$OUT/README.md" <<DOC
# CI Truth Repair Report

Generated: $(date)

Purpose: reproduce GitHub Actions failures locally without hiding or bypassing CI.

## Repositories Checked

$(for r in $repos; do echo "- $r"; done)

## How to use this report

Upload this whole report bundle into ChatGPT:

\`$(basename "$BUNDLE")\`

The most important files are:

- \`MASTER-REPORT.md\`
- \`logs/*.log\`
- \`summaries/*.summary.md\`
- \`workflows/*/files/*.yml\`
- \`packages/*/package.json\`
DOC

: > "$OUT/MASTER-REPORT.md"
{
  echo "# CI Truth Repair Master Report"
  echo
  echo "Generated: $(date)"
  echo
  echo "## Environment"
  echo
  echo '```text'
  cat "$OUT/metadata/environment.txt"
  echo '```'
  echo
  echo "## Results"
  echo
} >> "$OUT/MASTER-REPORT.md"

overall=0

for repo in $repos; do
  case "$repo" in
    booksmith-ai)
      cmd="npm ci && npm run validate:library && npm run lint && npm run build"
      ;;
    capital-city-provisions)
      cmd="npm ci && npm run verify"
      ;;
    tastycutz)
      cmd="npm ci && npm run verify"
      ;;
    BookSmith-Federation-OS)
      if [ -f "$ROOT/$repo/package.json" ]; then
        cmd="npm ci && npm run build"
      else
        cmd="find . -maxdepth 4 -type f | sort && test -d .github/workflows"
      fi
      ;;
    Aether_Coin_biozonecurrency)
      if [ -f "$ROOT/$repo/package.json" ]; then
        cmd="npm ci && npm run build"
      elif [ -f "$ROOT/$repo/go.mod" ]; then
        cmd="go test ./... && go build ./..."
      else
        cmd="find . -maxdepth 4 -type f | sort"
      fi
      ;;
    *)
      cmd="echo unknown repo"
      ;;
  esac

  if run_repo "$repo" "$cmd"; then
    echo "- ✅ $repo: PASS" >> "$OUT/MASTER-REPORT.md"
  else
    echo "- ❌ $repo: FAIL" >> "$OUT/MASTER-REPORT.md"
    overall=1
  fi
done

{
  echo
  echo "## Per-Repo Summaries"
  echo
  for f in "$OUT"/summaries/*.summary.md; do
    echo
    echo "---"
    echo
    cat "$f"
  done
} >> "$OUT/MASTER-REPORT.md"

cd "$OS" || exit 1
tar -czf "$BUNDLE" -C "$OS/reports/ci-truth-repair" "$STAMP"

echo
echo "=================================================="
echo "REPORT COMPLETE"
echo "=================================================="
echo "Folder:"
echo "$OUT"
echo
echo "Bundle to upload into ChatGPT:"
echo "$BUNDLE"
echo
echo "Master report:"
echo "$OUT/MASTER-REPORT.md"
echo

git add reports/ci-truth-repair reports/ci-truth-repair-"$STAMP".tar.gz 2>/dev/null || true
git commit -m "Add CI truth repair report $STAMP" || true
git push origin main || true

exit "$overall"
