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

for command in git npm node go make bash; do need "$command"; done
[[ -d "$OS_REPO/.git" ]] || die "AIFT-OS repository not found at $OS_REPO"
[[ -d "$CLIENT_REPO/.git" ]] || die "Mysterion repository not found at $CLIENT_REPO"

require_clean_os_worktree() {
  if [[ -n "$(git -C "$OS_REPO" status --porcelain)" ]]; then
    git -C "$OS_REPO" status --short >&2
    die "AIFT-OS worktree must be cleaned explicitly by its owner"
  fi
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

  log "Typechecking Mysterion when declared"
  (cd "$CLIENT_REPO" && npm run typecheck --if-present)

  log "Testing Mysterion when declared"
  (cd "$CLIENT_REPO" && npm test --if-present)

  log "Building Mysterion"
  (cd "$CLIENT_REPO" && npm run build)

  log "Auditing production dependencies"
  audit_json="$(mktemp "${TMPDIR:-/tmp}/aift-npm-audit.XXXXXX.json")"
  (cd "$CLIENT_REPO" && npm audit --omit=dev --json >"$audit_json") || true

  read -r high critical < <(
    node -e '
      const fs = require("fs");
      const report = JSON.parse(fs.readFileSync(process.argv[1], "utf8"));
      const v = report.metadata?.vulnerabilities ?? {};
      console.log(String(v.high ?? 0) + " " + String(v.critical ?? 0));
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
  rm -f -- "$audit_json"

  git -C "$CLIENT_REPO" diff --check
  git -C "$CLIENT_REPO" add -- package.json package-lock.json
  if ! git -C "$CLIENT_REPO" diff --cached --quiet; then
    git -C "$CLIENT_REPO" commit -m "chore: remediate production dependencies"
  fi
  log "Mysterion is verified locally; publication is deferred"
}

verify_os_feature() {
  require_clean_os_worktree

  log "Updating AIFT-OS main"
  git -C "$OS_REPO" fetch origin
  git -C "$OS_REPO" switch main
  git -C "$OS_REPO" pull --ff-only

  if git -C "$OS_REPO" show-ref --verify --quiet "refs/heads/$OBSOLETE_BRANCH"; then
    log "Deleting obsolete local renderer branch"
    git -C "$OS_REPO" branch -D "$OBSOLETE_BRANCH"
  fi

  if git -C "$OS_REPO" show-ref --verify --quiet "refs/heads/$OS_FEATURE_BRANCH"; then
    git -C "$OS_REPO" switch "$OS_FEATURE_BRANCH"
  else
    git -C "$OS_REPO" switch --track "origin/$OS_FEATURE_BRANCH"
  fi
  git -C "$OS_REPO" pull --ff-only

  log "Formatting the world-protocol Go sources"
  (
    cd "$OS_REPO"
    gofmt -w \
      internal/uxi/mobox_runtime.go \
      internal/uxi/world_protocol.go \
      internal/uxi/world_protocol_test.go \
      internal/uxihttp/server_test.go
    git diff --check
    git add -- \
      internal/uxi/mobox_runtime.go \
      internal/uxi/world_protocol.go \
      internal/uxi/world_protocol_test.go \
      internal/uxihttp/server_test.go
    if ! git diff --cached --quiet; then
      git commit -m "style: format world protocol sources"
      log "Formatting commit is verified locally; publication is deferred"
    fi
  )

  log "Running the complete AIFT-OS verification gate"
  (cd "$OS_REPO" && make verify)

  [[ -z "$(git -C "$OS_REPO" status --porcelain)" ]] ||
    die "AIFT-OS verification left the worktree dirty"
}

publish_verified_changes() {
  log "Publishing verified Mysterion changes"
  git -C "$CLIENT_REPO" push -u origin "$CLIENT_FIX_BRANCH"

  log "Publishing verified AIFT-OS changes"
  git -C "$OS_REPO" push origin "$OS_FEATURE_BRANCH"

  if git -C "$OS_REPO" show-ref --verify --quiet "refs/remotes/origin/$OBSOLETE_BRANCH"; then
    log "Deleting obsolete remote renderer branch after verification"
    git -C "$OS_REPO" push origin --delete "$OBSOLETE_BRANCH"
    git -C "$OS_REPO" fetch --prune origin
  fi
}

log "Finalizing Mysterion dependency remediation"
verify_client_changes

log "Finalizing the AIFT-OS world protocol"
verify_os_feature

publish_verified_changes

log "Federation finalization passed"
printf '%s\n'   "Mysterion dependency branch pushed: $CLIENT_FIX_BRANCH"   "AIFT-OS feature verified: $OS_FEATURE_BRANCH"   "No forced dependency upgrades or automatic merges were performed."
