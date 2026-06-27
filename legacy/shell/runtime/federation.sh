#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/common.sh"
. "$AIFT_OS_HOME/runtime/manifests.sh"

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
    manifest=false
    valid=false

    aift_git_dirty "$repo" && dirty=true
    [ -f "$(aift_manifest_path "$repo")" ] && manifest=true
    aift_validate_manifest "$repo" && valid=true || true

    [ "$first" -eq 1 ] || printf ',\n' >> "$out"
    first=0
    printf '  {"name":"%s","path":"%s","branch":"%s","remote":"%s","dirty":%s,"manifest":%s,"manifestValid":%s}' \
      "$name" "$repo" "$branch" "$remote" "$dirty" "$manifest" "$valid" >> "$out"
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
    printf '| Repository | Branch | State | Manifest | Remote |\n'
    printf '|---|---|---|---|---|\n'

    for repo in $(aift_find_repos); do
      name="$(aift_repo_name "$repo")"
      branch="$(aift_git_branch "$repo")"
      state="clean"
      manifest="missing"
      aift_git_dirty "$repo" && state="dirty"
      aift_validate_manifest "$repo" && manifest="valid" || true
      printf '| `%s` | `%s` | `%s` | `%s` | `%s` |\n' \
        "$name" "$branch" "$state" "$manifest" "$(aift_remote_url "$repo")"
    done
  } > "$out"

  printf '%s\n' "$out"
}

aift_dashboard(){
  out="$AIFT_OS_HOME/reports/dashboard.md"
  mkdir -p "$AIFT_OS_HOME/reports"

  total=0
  dirty=0
  clean=0
  valid=0
  missing=0

  for repo in $(aift_find_repos); do
    total=$((total + 1))
    if aift_git_dirty "$repo"; then dirty=$((dirty + 1)); else clean=$((clean + 1)); fi
    if aift_validate_manifest "$repo"; then valid=$((valid + 1)); else missing=$((missing + 1)); fi
  done

  {
    printf '# AIFT-OS Federation Dashboard\n\n'
    printf '- Total repositories: %s\n' "$total"
    printf '- Clean repositories: %s\n' "$clean"
    printf '- Dirty repositories: %s\n' "$dirty"
    printf '- Valid manifests: %s\n' "$valid"
    printf '- Missing/invalid manifests: %s\n\n' "$missing"

    printf '## Repositories\n\n'
    printf '| Repository | Branch | State | Manifest |\n'
    printf '|---|---|---|---|\n'

    for repo in $(aift_find_repos); do
      name="$(aift_repo_name "$repo")"
      branch="$(aift_git_branch "$repo")"
      state="clean"
      manifest="valid"
      aift_git_dirty "$repo" && state="dirty"
      aift_validate_manifest "$repo" || manifest="missing/invalid"
      printf '| `%s` | `%s` | `%s` | `%s` |\n' "$name" "$branch" "$state" "$manifest"
    done
  } > "$out"

  printf '%s\n' "$out"
}
