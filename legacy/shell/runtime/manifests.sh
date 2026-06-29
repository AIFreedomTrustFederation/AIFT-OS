#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/common.sh"

aift_create_manifest_if_missing(){
  repo="$1"
  name="$(aift_repo_name "$repo")"
  mkdir -p "$repo/.aift/commands"
  file="$repo/.aift/repo.json"

  [ -f "$file" ] && return 0

  role="sovereign-repository"
  [ "$name" = "AIFT-OS" ] && role="federation-control-plane"

  cat > "$file" <<JSON
{
  "name": "$name",
  "role": "$role",
  "sovereign": true,
  "managedBy": "AIFT-OS",
  "dependencies": [],
  "capabilities": [],
  "commandsPath": ".aift/commands"
}
JSON
}

aift_validate_manifest(){
  repo="$1"
  file="$(aift_manifest_path "$repo")"
  [ -f "$file" ] || return 1
  grep -q '"name"' "$file" || return 1
  grep -q '"role"' "$file" || return 1
  return 0
}
