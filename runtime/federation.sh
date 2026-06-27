#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/common.sh"

aift_registry_json(){
  out="$AIFT_OS_HOME/registry/repos.json"
  mkdir -p "$AIFT_OS_HOME/registry"
  printf '[\n' > "$out"
  first=1
  for repo in $(aift_find_repos); do
    name="$(aift_repo_name "$repo")"
    branch="$(aift_git_branch "$repo")"
    remote="$(aift_remote_url "$repo")"
    dirty=false
    if aift_git_dirty "$repo"; then dirty=true; fi

    [ "$first" -eq 1 ] || printf ',\n' >> "$out"
    first=0
    printf '  {"name":"%s","path":"%s","branch":"%s","remote":"%s","dirty":%s}' \
      "$name" "$repo" "$branch" "$remote" "$dirty" >> "$out"
  done
  printf '\n]\n' >> "$out"
  printf '%s\n' "$out"
}

aift_graph_markdown(){
  out="$AIFT_OS_HOME/reports/federation-graph.md"
  mkdir -p "$AIFT_OS_HOME/reports"
  {
    printf '# AIFT Federation Graph\n\n'
    printf 'Control plane: `%s`\n\n' "$AIFT_OS_HOME"
    printf '| Repository | Branch | Dirty | Remote |\n'
    printf '|---|---:|---:|---|\n'
    for repo in $(aift_find_repos); do
      name="$(aift_repo_name "$repo")"
      branch="$(aift_git_branch "$repo")"
      dirty="clean"
      if aift_git_dirty "$repo"; then dirty="dirty"; fi
      remote="$(aift_remote_url "$repo")"
      printf '| `%s` | `%s` | `%s` | `%s` |\n' "$name" "$branch" "$dirty" "$remote"
    done
  } > "$out"
  printf '%s\n' "$out"
}
