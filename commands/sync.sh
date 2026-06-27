#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/common.sh"

mode="${1:---safe}"
msg="${2:-AIFT federation sync}"

case "$mode" in
  --safe|safe)
    echo "AIFT safe sync: pulls clean repos only; dirty repos are skipped."
    for repo in $(aift_find_repos); do
      name="$(aift_repo_name "$repo")"
      remote="$(aift_remote_url "$repo")"
      [ -n "$remote" ] || { echo "$name: skip, no origin"; continue; }

      if aift_git_dirty "$repo"; then
        echo "$name: skip, dirty"
        continue
      fi

      branch="$(aift_git_branch "$repo")"
      echo "$name: pull --rebase origin $branch"
      git -C "$repo" pull --rebase origin "$branch" || true
    done
    ;;

  --commit|commit)
    echo "AIFT commit sync: commits dirty repos, pulls, then pushes."
    for repo in $(aift_find_repos); do
      name="$(aift_repo_name "$repo")"
      remote="$(aift_remote_url "$repo")"
      [ -n "$remote" ] || { echo "$name: skip, no origin"; continue; }

      if aift_git_dirty "$repo"; then
        git -C "$repo" add .
        git -C "$repo" commit -m "$msg" || true
      fi

      branch="$(aift_git_branch "$repo")"
      git -C "$repo" pull --rebase origin "$branch" || true
      git -C "$repo" push origin "$branch" || true
    done
    ;;

  *)
    echo "Usage:"
    echo "  aift sync --safe"
    echo "  aift sync --commit \"message\""
    exit 1
    ;;
esac
