#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/common.sh"

aift_workspace_manifest(){
  mkdir -p "$AIFT_OS_HOME/manifests"
  cat > "$AIFT_OS_HOME/manifests/workspace.json" <<JSON
{
  "name": "AI Freedom Trust Federation Workspace",
  "root": "$AIFT_ROOT",
  "controlPlane": "$AIFT_OS_HOME",
  "principle": "Other repositories remain sovereign; AIFT-OS orchestrates but does not absorb them."
}
JSON
}
