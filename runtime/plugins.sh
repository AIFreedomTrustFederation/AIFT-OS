#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/common.sh"

aift_plugin_command_dirs(){
  for repo in $(aift_find_repos); do
    [ -d "$repo/.aift/commands" ] && printf '%s\n' "$repo/.aift/commands"
  done
}

aift_find_plugin_command(){
  cmd="$1"
  for dir in $(aift_plugin_command_dirs); do
    [ -x "$dir/$cmd.sh" ] && { printf '%s\n' "$dir/$cmd.sh"; return 0; }
    [ -f "$dir/$cmd.sh" ] && { printf '%s\n' "$dir/$cmd.sh"; return 0; }
  done
  return 1
}

aift_list_plugins(){
  for dir in $(aift_plugin_command_dirs); do
    repo="$(dirname "$(dirname "$dir")")"
    for f in "$dir"/*.sh; do
      [ -f "$f" ] || continue
      printf '%s :: %s\n' "$(basename "$repo")" "$(basename "$f" .sh)"
    done
  done
}
