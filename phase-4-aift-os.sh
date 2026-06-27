#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
export AIFT_ROOT="$ROOT"
export AIFT_OS_HOME="$OS"

cd "$OS" || exit 1

mkdir -p \
  bin commands runtime install scripts manifests registry intelligence reports docs schemas templates examples logs var \
  providers plugins config

cat > config/aift-os.env <<'CONFIG'
# AIFT-OS local configuration
AIFT_ROOT="$HOME/AIFT"
AIFT_OS_HOME="$HOME/AIFT/AIFT-OS"
AIFT_DEFAULT_BRANCH="main"
AIFT_SYNC_MODE="safe"
CONFIG

cat > runtime/config.sh <<'CONFIGSH'
#!/usr/bin/env sh
set -eu

AIFT_ROOT="${AIFT_ROOT:-$HOME/AIFT}"
AIFT_OS_HOME="${AIFT_OS_HOME:-$AIFT_ROOT/AIFT-OS}"

if [ -f "$AIFT_OS_HOME/config/aift-os.env" ]; then
  # shellcheck disable=SC1090
  . "$AIFT_OS_HOME/config/aift-os.env"
fi

export AIFT_ROOT AIFT_OS_HOME
CONFIGSH

cat > runtime/common.sh <<'COMMON'
#!/usr/bin/env sh
set -eu

. "${AIFT_OS_HOME:-$HOME/AIFT/AIFT-OS}/runtime/config.sh"

aift_log(){ printf '%s\n' "$*"; }
aift_warn(){ printf 'WARN: %s\n' "$*" >&2; }
aift_die(){ printf 'ERROR: %s\n' "$*" >&2; exit 1; }

aift_repo_name(){ basename "$1"; }

aift_find_repos(){
  find "$AIFT_ROOT" -mindepth 1 -maxdepth 2 -type d -name .git 2>/dev/null \
    | sed 's#/.git$##' \
    | sort
}

aift_git_branch(){
  git -C "$1" rev-parse --abbrev-ref HEAD 2>/dev/null || printf 'unknown'
}

aift_git_dirty(){
  [ -n "$(git -C "$1" status --porcelain 2>/dev/null || true)" ]
}

aift_remote_url(){
  git -C "$1" remote get-url origin 2>/dev/null || printf ''
}

aift_manifest_path(){
  printf '%s/.aift/repo.json' "$1"
}
COMMON

cat > runtime/plugins.sh <<'PLUGINS'
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
PLUGINS

cat > runtime/manifests.sh <<'MANIFESTS'
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
MANIFESTS

cat > runtime/federation.sh <<'FEDERATION'
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
FEDERATION

cat > providers/git.sh <<'GITPROVIDER'
#!/usr/bin/env sh
set -eu

# Git provider interface for AIFT-OS.
# Future providers can implement the same commands for GitHub, Forge, local-only, or federation relay.

git_provider_name(){
  printf 'local-git\n'
}

git_provider_status(){
  repo="$1"
  git -C "$repo" status --short
}

git_provider_pull_safe(){
  repo="$1"
  branch="$(git -C "$repo" rev-parse --abbrev-ref HEAD)"
  git -C "$repo" pull --rebase origin "$branch"
}

git_provider_push(){
  repo="$1"
  branch="$(git -C "$repo" rev-parse --abbrev-ref HEAD)"
  git -C "$repo" push origin "$branch"
}
GITPROVIDER

cat > schemas/repo-manifest.schema.json <<'SCHEMA'
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "AIFT Repository Manifest",
  "type": "object",
  "required": ["name", "role", "sovereign"],
  "properties": {
    "name": { "type": "string" },
    "role": { "type": "string" },
    "sovereign": { "type": "boolean" },
    "managedBy": { "type": "string" },
    "dependencies": {
      "type": "array",
      "items": { "type": "string" }
    },
    "capabilities": {
      "type": "array",
      "items": { "type": "string" }
    },
    "commandsPath": { "type": "string" }
  }
}
SCHEMA

cat > commands/manifest.sh <<'MANIFESTCMD'
#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/manifests.sh"

for repo in $(aift_find_repos); do
  aift_create_manifest_if_missing "$repo"
  echo "manifest: $(aift_repo_name "$repo")"
done
MANIFESTCMD

cat > commands/plugins.sh <<'PLUGINSCMD'
#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/plugins.sh"

echo "AIFT plugin commands:"
aift_list_plugins || true
PLUGINSCMD

cat > commands/dashboard.sh <<'DASHBOARD'
#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/federation.sh"

out="$(aift_dashboard)"
echo "Wrote $out"
DASHBOARD

cat > commands/deps.sh <<'DEPS'
#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/common.sh"

out="$AIFT_OS_HOME/reports/dependency-graph.md"
mkdir -p "$AIFT_OS_HOME/reports"

{
  echo "# AIFT Dependency Graph"
  echo
  echo "| Repository | Dependencies |"
  echo "|---|---|"

  for repo in $(aift_find_repos); do
    name="$(aift_repo_name "$repo")"
    file="$(aift_manifest_path "$repo")"
    deps="[]"
    if [ -f "$file" ]; then
      deps="$(grep '"dependencies"' "$file" 2>/dev/null || echo '"dependencies": []')"
    fi
    deps="$(printf '%s' "$deps" | sed 's/^[[:space:]]*//; s/[",]//g')"
    echo "| \`$name\` | \`$deps\` |"
  done
} > "$out"

echo "Wrote $out"
DEPS

cat > commands/sync.sh <<'SYNC'
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
SYNC

cat > commands/verify.sh <<'VERIFY'
#!/usr/bin/env sh
set -eu

"$AIFT_OS_HOME/aift-os.sh" doctor
"$AIFT_OS_HOME/aift-os.sh" manifest
"$AIFT_OS_HOME/aift-os.sh" registry
"$AIFT_OS_HOME/aift-os.sh" graph
"$AIFT_OS_HOME/aift-os.sh" deps
"$AIFT_OS_HOME/aift-os.sh" dashboard

echo "OK: federation verified"
VERIFY

cat > aift-os.sh <<'MAIN'
#!/usr/bin/env sh
set -eu

AIFT_ROOT="${AIFT_ROOT:-$HOME/AIFT}"
AIFT_OS_HOME="${AIFT_OS_HOME:-$AIFT_ROOT/AIFT-OS}"
export AIFT_ROOT AIFT_OS_HOME

cmd="${1:-help}"
shift || true

case "$cmd" in
  help|-h|--help)
    cat <<HELP
AIFT-OS Federation Control Plane

Usage:
  aift <command>

Core:
  help        Show help
  doctor      Inspect control-plane health
  status      Show federation repo status
  verify      Run full verification

Federation:
  manifest    Create missing .aift/repo.json manifests
  registry    Generate registry/repos.json
  graph       Generate federation graph
  deps        Generate dependency graph
  dashboard   Generate dashboard report
  plugins     List plugin commands
  sync        Safe sync or explicit commit sync

Sync modes:
  aift sync --safe
  aift sync --commit "commit message"

Plugin commands:
  Any repo may expose commands at:
    .aift/commands/<command>.sh
HELP
    ;;
  *)
    builtin="$AIFT_OS_HOME/commands/$cmd.sh"
    if [ -f "$builtin" ]; then
      sh "$builtin" "$@"
      exit $?
    fi

    . "$AIFT_OS_HOME/runtime/plugins.sh"
    plugin="$(aift_find_plugin_command "$cmd" || true)"
    if [ -n "$plugin" ]; then
      sh "$plugin" "$@"
      exit $?
    fi

    echo "Unknown command: $cmd" >&2
    echo "Run: aift help" >&2
    exit 1
    ;;
esac
MAIN

cat > docs/PHASE-4.md <<'DOC'
# AIFT-OS Phase 4

Phase 4 turns AIFT-OS into an extensible federation control plane.

Implemented:

- Local configuration through `config/aift-os.env`
- Repository manifests through `.aift/repo.json`
- Plugin command discovery through `.aift/commands`
- Federation registry generation
- Federation graph report
- Dependency graph report
- Dashboard report
- Provider interface foundation
- Safe sync mode
- Explicit commit sync mode

AIFT-OS does not absorb sovereign repositories. It discovers, validates, reports, and orchestrates them.
DOC

chmod +x aift-os.sh bin/aift commands/*.sh runtime/*.sh providers/*.sh

./aift-os.sh verify

git add .
if git diff --cached --quiet; then
  echo "Nothing new to commit."
else
  git commit -m "Add AIFT-OS plugin and federation manifest architecture"
fi

git push origin main

echo
echo "DONE."
echo "Try:"
echo "  ~/AIFT/aift help"
echo "  ~/AIFT/aift plugins"
echo "  ~/AIFT/aift dashboard"
echo "  ~/AIFT/aift deps"
echo "  ~/AIFT/aift sync --safe"
