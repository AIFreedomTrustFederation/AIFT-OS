#!/usr/bin/env bash
# Opt-in Termux/Ubuntu AIFT operator helper. No background automation.
# no-harness: This standalone, human-invoked operator interface runs outside the federation mutating-script harness and has its own approval gates.
set -Eeuo pipefail
umask 077
REPO=/root/AIFT/AIFT-OS
WORKTREE=/root/AIFT/AIFT-OS-PULL
GITHUB_REPO=AIFreedomTrustFederation/AIFT-OS
MODEL=aetherion:latest
ENDPOINT=http://127.0.0.1:11434/v1

fail() { echo "BLOCKED: $*" >&2; exit 1; }
usage() {
  cat <<'EOF'
Usage: bash scripts/aift_termux_operator.sh COMMAND [args]
Credentialless Termux/Ubuntu: prepare | verify | status | once | inspect ISSUE
  export ISSUE DIGEST | discard ISSUE DIGEST | delivered ISSUE
Separate protected GitHub operator context: protection-status | protection-apply
  publish B64FILE DIGEST
Required before privileged commands: AIFT_OPERATOR_SEPARATE=yes.
Also for protection-apply: AIFT_REVIEWER_READY=yes (independent reviewer).
Also for publish: AIFT_OPERATOR_APPROVED=yes.
Never merge, deploy, or enroll a runner from this script.
EOF
}
cd_worktree() {
  [[ -d "$WORKTREE" ]] || fail "No worktree; run prepare"
  cd "$WORKTREE"
}
check_worktree() {
  cd_worktree
  [[ -z "$(git status --porcelain)" ]] || fail "Worktree not clean"
  local head remote branch
  head="$(git rev-parse HEAD)"
  remote="$(git ls-remote origin refs/heads/main | awk '{print $1}')"
  [[ -n "$remote" && "$remote" == "$head" ]] || fail "Stale checkout: fetch/review main first"
  branch="$(git branch --show-current)"
  [[ -z "$branch" || "$branch" == main ]] || fail "Unexpected branch"
}
need_operator() {
  [[ "$(printenv AIFT_OPERATOR_SEPARATE || true)" == yes ]] ||
    fail "Use separate protected operator environment; set AIFT_OPERATOR_SEPARATE=yes"
  command -v gh >/dev/null || fail "gh missing in operator context"
  gh auth status >/dev/null 2>&1 || fail "Connect gh in operator context"
}
(( $# >= 1 )) || { usage; exit 1; }
command_name="$1"; shift
case "$command_name" in
  help) usage ;;
  prepare)
    [[ -d "$REPO/.git" || -f "$REPO/.git" ]] || fail "Missing original repo"
    git -C "$REPO" fetch origin main
    if [[ -e "$WORKTREE" ]]; then
      echo "Existing worktree preserved: $WORKTREE"
    else
      git -C "$REPO" worktree add --detach "$WORKTREE" origin/main
    fi
    check_worktree
    echo "Clean worktree at $(git rev-parse HEAD)"
    ;;
  verify)
    check_worktree
    for suite in test_aift_secure_cloud_worker.py test_aift_publisher_artifact.py \
      test_aift_aetherion_readiness.py test_aift_pull_docs_worker.py \
      test_aift_pull_publisher.py test_aift_pull_plan_handoff.py \
      test_aift_pull_postmerge_security.py
    do
      python3 -B -m unittest discover -s tests -p "$suite" -v
    done
    echo "Offline checks passed; live publication is NOT verified"
    ;;
  status) cd_worktree; python3 -B scripts/aift_pull_docs_worker.py status ;;
  once)
    check_worktree
    python3 -B scripts/aift_pull_docs_worker.py once \
      --model "$MODEL" --model-url "$ENDPOINT"
    ;;
  inspect)
    (( $# == 1 )) || fail "inspect ISSUE"
    cd_worktree
    python3 -B scripts/aift_pull_docs_worker.py inspect --issue "$1"
    ;;
  export)
    (( $# == 2 )) || fail "export ISSUE DIGEST"
    cd_worktree
    issue="$1"; digest="$2"
    [[ "$issue" =~ ^[1-9][0-9]*$ && "$digest" =~ ^[0-9a-f]{64}$ ]] || fail "Invalid issue/digest"
    exportdir="$HOME/.local/state/aift-aetherion-export"
    mkdir -p "$exportdir"; chmod 700 "$exportdir"
    outfile="$exportdir/issue-$issue-$digest.b64"
    [[ ! -e "$outfile" ]] || fail "Export exists; not overwriting"
    python3 -B scripts/aift_pull_docs_worker.py export \
      --issue "$issue" --expected-digest "$digest" > "$outfile"
    chmod 600 "$outfile"
    echo "Transfer ONLY the reviewed public-docs artifact to protected operator environment:"
    echo "$outfile"
    ;;
  discard)
    (( $# == 2 )) || fail "discard ISSUE DIGEST"
    cd_worktree
    python3 -B scripts/aift_pull_docs_worker.py discard \
      --issue "$1" --expected-digest "$2"
    ;;
  delivered)
    (( $# == 1 )) || fail "delivered ISSUE"
    [[ "$(printenv AIFT_DRAFT_PR_VERIFIED || true)" == yes ]] ||
      fail "Verify the actual draft PR and set AIFT_DRAFT_PR_VERIFIED=yes"
    cd_worktree
    python3 -B scripts/aift_pull_docs_worker.py mark-delivered --issue "$1"
    ;;
  protection-status)
    need_operator
    gh api "repos/$GITHUB_REPO/branches/main/protection" ||
      echo "Protection absent or unavailable: inspect GitHub permissions"
    ;;
  protection-apply)
    need_operator
    [[ "$(printenv AIFT_REVIEWER_READY || true)" == yes ]] ||
      fail "Independent reviewer required; set AIFT_REVIEWER_READY=yes after confirming"
    if gh api "repos/$GITHUB_REPO/branches/main/protection" >/dev/null 2>&1; then
      fail "Existing protection found; will not overwrite"
    fi
    tmp="$(mktemp)"
    trap 'rm -f "$tmp"' EXIT
    cat > "$tmp" <<'JSON'
{
  "required_status_checks":{"strict":true,"contexts":["test"]},
  "enforce_admins":true,
  "required_pull_request_reviews":{
    "dismiss_stale_reviews":true,
    "required_approving_review_count":1
  },
  "restrictions":null,
  "required_linear_history":true,
  "allow_force_pushes":false,
  "allow_deletions":false,
  "required_conversation_resolution":true
}
JSON
    gh api --method PUT "repos/$GITHUB_REPO/branches/main/protection" --input "$tmp"
    echo "Inspect actual GitHub rules after applying; no bypass permitted"
    ;;
  publish)
    (( $# == 2 )) || fail "publish B64FILE DIGEST"
    need_operator
    [[ "$(printenv AIFT_OPERATOR_APPROVED || true)" == yes ]] ||
      fail "Explicit approval required: AIFT_OPERATOR_APPROVED=yes"
    file="$1"; digest="$2"
    [[ -f "$file" && ! -L "$file" && "$digest" =~ ^[0-9a-f]{64}$ ]] ||
      fail "Reviewed .b64 file and digest required"
    actual="$(base64 --decode "$file" | sha256sum | awk '{print $1}')"
    [[ "$digest" == "$actual" ]] || fail "Digest mismatch"
    encoded="$(tr -d '\r\n' < "$file")"
    gh workflow run aift-pull-doc-publisher.yml --repo "$GITHUB_REPO" --ref main \
      -f operator_approved=true -f plan_sha256="$digest" -f plan_b64="$encoded"
    unset encoded
    echo "Dispatch requested. Verify Actions and an actual DRAFT PR separately."
    ;;
  *) usage; fail "Unknown command: $command_name" ;;
esac
