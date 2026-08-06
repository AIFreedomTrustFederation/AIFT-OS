#!/usr/bin/env bash
# no-harness: coordinates two sibling repositories and invokes their native verification gates
set -Eeuo pipefail

readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly OS_REPO="$(cd -- "$SCRIPT_DIR/.." && pwd)"
readonly AIFT_ROOT="${AIFT_ROOT:-$(cd -- "$OS_REPO/.." && pwd)}"
readonly CLIENT_REPO="${AIFT_CLIENT_REPO:-$AIFT_ROOT/c-848263}"
readonly OS_FEATURE_BRANCH="agent/world-protocol-adapter"
readonly CLIENT_FIX_BRANCH="fix/production-dependency-audit"
readonly OBSOLETE_BRANCH="agent/interactive-geometry-renderer"

log() { printf '\n==> %s\n' "$*"; }
die() { printf '\nERROR: %s\n' "$*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || die "Required command not found: $1"; }

for command in git npm go make bash; do need "$command"; done
[[ -d "$OS_REPO/.git" ]] || die "AIFT-OS repository not found at $OS_REPO"
[[ -d "$CLIENT_REPO/.git" ]] || die "Mysterion repository not found at $CLIENT_REPO"

clean_known_os_whitespace() {
  local status
  status="$(git -C "$OS_REPO" status --porcelain)"
  [[ -z "$status" ]] && return 0

  if [[ "$status" == " M internal/uxihttp/server_test.go" ]] &&
     git -C "$OS_REPO" diff -w --quiet -- internal/uxihttp/server_test.go; then
    log "Restoring the known whitespace-only AIFT-OS test change"
    git -C "$OS_REPO" restore -- internal/uxihttp/server_test.go
    return 0
  fi

  git -C "$OS_REPO" status --short >&2
  die "AIFT-OS contains work not recognized as disposable; nothing was changed"
}

verify_client_changes() {
  local unexpected audit_json high critical

  [[ "$(git -C "$CLIENT_REPO" branch --show-current)" == "$CLIENT_FIX_BRANCH" ]] ||
    die "Mysterion must be on $CLIENT_FIX_BRANCH"

  unexpected="$(git -C "$CLIENT_REPO" status --porcelain |
    awk '$2 != "package.json" && $2 != "package-lock.json" { print }')"
  [[ -z "$unexpected" ]] || {
    printf '%s\n' "$unexpected" >&2
    die "Mysterion contains changes outside package.json and package-lock.json"
  }

  log "Installing the locked Mysterion dependency graph"
  (cd "$CLIENT_REPO" && npm ci)

  log "Linting Mysterion"
  (cd "$CLIENT_REPO" && npm run lint)

  log "Building Mysterion"
  (cd "$CLIENT_REPO" && npm run build)

  log "Auditing production dependencies"
  audit_json="$(mktemp "${TMPDIR:-/tmp}/aift-npm-audit.XXXXXX.json")"
  trap 'rm -f -- "$audit_json"' RETURN
  (cd "$CLIENT_REPO" && npm audit --omit=dev --json >"$audit_json") || true

  read -r high critical < <(
    node -e '
      const fs = require("fs");
      const report = JSON.parse(fs.readFileSync(process.argv[1], "utf8"));
      const v = report.metadata?.vulnerabilities ?? {};
      process.stdout.write(String(v.high ?? 0) + " " + String(v.critical ?? 0));
    ' "$audit_json"
  )
  [[ "$high" == 0 && "$critical" == 0 ]] ||
    die "Production audit still reports $high high and $critical critical vulnerabilities"

  node -e '
    const fs = require("fs");
    const report = JSON.parse(fs.readFileSync(process.argv[1], "utf8"));
    const v = report.metadata?.vulnerabilities ?? {};
    console.log("Production audit:", JSON.stringify(v));
  ' "$audit_json"

  git -C "$CLIENT_REPO" diff --check
  git -C "$CLIENT_REPO" add -- package.json package-lock.json
  if ! git -C "$CLIENT_REPO" diff --cached --quiet; then
    git -C "$CLIENT_REPO" commit -m "chore: remediate production dependencies"
  fi
  git -C "$CLIENT_REPO" push -u origin "$CLIENT_FIX_BRANCH"
}

verify_os_feature() {
  clean_known_os_whitespace

  log "Updating AIFT-OS main"
  git -C "$OS_REPO" fetch origin
  git -C "$OS_REPO" switch main
  git -C "$OS_REPO" pull --ff-only

  if git -C "$OS_REPO" show-ref --verify --quiet "refs/heads/$OBSOLETE_BRANCH"; then
    log "Deleting obsolete local renderer branch"
    git -C "$OS_REPO" branch -D "$OBSOLETE_BRANCH"
  fi

  if git -C "$OS_REPO" show-ref --verify --quiet "refs/remotes/origin/$OBSOLETE_BRANCH"; then
    log "Deleting obsolete remote renderer branch"
    git -C "$OS_REPO" push origin --delete "$OBSOLETE_BRANCH"
    git -C "$OS_REPO" fetch --prune origin
  fi

  if git -C "$OS_REPO" show-ref --verify --quiet "refs/heads/$OS_FEATURE_BRANCH"; then
    git -C "$OS_REPO" switch "$OS_FEATURE_BRANCH"
  else
    git -C "$OS_REPO" switch --track "origin/$OS_FEATURE_BRANCH"
  fi
  git -C "$OS_REPO" pull --ff-only

  log "Running the complete AIFT-OS verification gate"
  (cd "$OS_REPO" && make verify)

  [[ -z "$(git -C "$OS_REPO" status --porcelain)" ]] ||
    die "AIFT-OS verification left the worktree dirty"
}

log "Finalizing Mysterion dependency remediation"
verify_client_changes

log "Finalizing the AIFT-OS world protocol"
verify_os_feature

log "Federation finalization passed"
printf '%s\n'   "Mysterion dependency branch pushed: $CLIENT_FIX_BRANCH"   "AIFT-OS feature verified: $OS_FEATURE_BRANCH"   "No forced dependency upgrades or automatic merges were performed."
