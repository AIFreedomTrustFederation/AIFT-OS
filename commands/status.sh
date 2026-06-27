#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/common.sh"

printf '%-32s %-12s %-8s %s\n' "REPOSITORY" "BRANCH" "STATE" "REMOTE"
for repo in $(aift_find_repos); do
  name="$(aift_repo_name "$repo")"
  branch="$(aift_git_branch "$repo")"
  state="clean"
  if aift_git_dirty "$repo"; then state="dirty"; fi
  remote="$(aift_remote_url "$repo")"
  printf '%-32s %-12s %-8s %s\n' "$name" "$branch" "$state" "$remote"
done
