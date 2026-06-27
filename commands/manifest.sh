#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/manifests.sh"

for repo in $(aift_find_repos); do
  aift_create_manifest_if_missing "$repo"
  echo "manifest: $(aift_repo_name "$repo")"
done
