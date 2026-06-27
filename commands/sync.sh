#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/common.sh"

msg="${1:-AIFT federation sync}"

for repo in $(aift_find_repos); do
  name="$(aift_repo_name "$repo")"
  echo "== $name =="

  if [ -z "$(aift_remote_url "$repo")" ]; then
    echo "skip: no origin remote"
    continue
  fi

  if aift_git_dirty "$repo"; then
    git -C "$repo" add .
    git -C "$repo" commit -m "$msg" || true
  fi

  branch="$(aift_git_branch "$repo")"
  git -C "$repo" pull --rebase origin "$branch" || true
  git -C "$repo" push origin "$branch" || true
done
