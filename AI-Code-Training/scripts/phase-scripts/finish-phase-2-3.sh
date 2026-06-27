#!/usr/bin/env sh
set -eu

ROOT="$HOME/AIFT"
OS="$ROOT/AIFT-OS"
export AIFT_ROOT="$ROOT"
export AIFT_OS_HOME="$OS"

cd "$OS" || exit 1

mkdir -p bin commands runtime install scripts manifests registry intelligence reports docs schemas templates examples logs var

# Flatten accidental doubled folders
for d in runtime scripts manifests; do
  if [ -d "$d/$d" ]; then
    find "$d/$d" -mindepth 1 -maxdepth 1 -exec mv {} "$d/" \;
    rmdir "$d/$d" 2>/dev/null || true
  fi
done

cat > runtime/common.sh <<'COMMON'
#!/usr/bin/env sh
set -eu

AIFT_ROOT="${AIFT_ROOT:-$HOME/AIFT}"
AIFT_OS_HOME="${AIFT_OS_HOME:-$AIFT_ROOT/AIFT-OS}"

aift_log(){ printf '%s\n' "$*"; }
aift_warn(){ printf 'WARN: %s\n' "$*" >&2; }
aift_die(){ printf 'ERROR: %s\n' "$*" >&2; exit 1; }

aift_is_repo(){
  [ -d "$1/.git" ]
}

aift_repo_name(){
  basename "$1"
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

aift_find_repos(){
  find "$AIFT_ROOT" -mindepth 1 -maxdepth 2 -type d -name .git 2>/dev/null \
    | sed 's#/.git$##' \
    | sort
}
COMMON

cat > runtime/workspace.sh <<'WORKSPACE'
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
WORKSPACE

cat > runtime/federation.sh <<'FEDERATION'
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
FEDERATION

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
  aift-os.sh <command>

Commands:
  help       Show this help
  doctor     Inspect local control-plane health
  status     Show federation repository status
  registry   Generate registry/repos.json
  graph      Generate reports/federation-graph.md
  verify     Run doctor + registry + graph
  sync       Commit/pull/push sovereign repos safely
  install    Install top-level launchers
HELP
    ;;
  *)
    file="$AIFT_OS_HOME/commands/$cmd.sh"
    [ -f "$file" ] || {
      echo "Unknown command: $cmd" >&2
      echo "Run: aift help" >&2
      exit 1
    }
    sh "$file" "$@"
    ;;
esac
MAIN

cat > bin/aift <<'BIN'
#!/usr/bin/env sh
AIFT_ROOT="${AIFT_ROOT:-$HOME/AIFT}"
AIFT_OS_HOME="${AIFT_OS_HOME:-$AIFT_ROOT/AIFT-OS}"
exec "$AIFT_OS_HOME/aift-os.sh" "$@"
BIN

cat > commands/doctor.sh <<'DOCTOR'
#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/common.sh"

echo "AIFT-OS Doctor"
echo "root: $AIFT_ROOT"
echo "os:   $AIFT_OS_HOME"

[ -d "$AIFT_ROOT" ] || aift_die "Missing workspace root"
[ -d "$AIFT_OS_HOME/.git" ] || aift_die "AIFT-OS is not a git repo"

for d in bin commands runtime install scripts manifests registry intelligence reports docs schemas templates examples logs var; do
  [ -d "$AIFT_OS_HOME/$d" ] || aift_die "Missing directory: $d"
done

bad="$(find "$AIFT_OS_HOME" -type d \( -path "$AIFT_OS_HOME/runtime/runtime" -o -path "$AIFT_OS_HOME/scripts/scripts" -o -path "$AIFT_OS_HOME/manifests/manifests" \) -print)"
[ -z "$bad" ] || aift_die "Doubled folders remain: $bad"

echo "OK: control plane layout healthy"
DOCTOR

cat > commands/status.sh <<'STATUS'
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
STATUS

cat > commands/registry.sh <<'REGISTRY'
#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/federation.sh"
out="$(aift_registry_json)"
echo "Wrote $out"
REGISTRY

cat > commands/graph.sh <<'GRAPH'
#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/federation.sh"
out="$(aift_graph_markdown)"
echo "Wrote $out"
GRAPH

cat > commands/verify.sh <<'VERIFY'
#!/usr/bin/env sh
set -eu
"$AIFT_OS_HOME/aift-os.sh" doctor
"$AIFT_OS_HOME/aift-os.sh" registry
"$AIFT_OS_HOME/aift-os.sh" graph
echo "OK: federation verified"
VERIFY

cat > commands/install.sh <<'INSTALL'
#!/usr/bin/env sh
set -eu
mkdir -p "$AIFT_ROOT"
cat > "$AIFT_ROOT/aift-os.sh" <<LAUNCH
#!/usr/bin/env sh
exec "\$HOME/AIFT/AIFT-OS/aift-os.sh" "\$@"
LAUNCH
cat > "$AIFT_ROOT/aift" <<LAUNCH
#!/usr/bin/env sh
exec "\$HOME/AIFT/AIFT-OS/bin/aift" "\$@"
LAUNCH
chmod +x "$AIFT_ROOT/aift-os.sh" "$AIFT_ROOT/aift" "$AIFT_OS_HOME/aift-os.sh" "$AIFT_OS_HOME/bin/aift"
echo "Installed launchers:"
echo "  $AIFT_ROOT/aift"
echo "  $AIFT_ROOT/aift-os.sh"
INSTALL

cat > commands/sync.sh <<'SYNC'
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
SYNC

chmod +x aift-os.sh bin/aift commands/*.sh runtime/*.sh

# Install top-level launchers
sh commands/install.sh

# Generate phase 3 outputs
sh aift-os.sh verify

# Commit and push
git add .
if git diff --cached --quiet; then
  echo "Nothing new to commit."
else
  git commit -m "Build AIFT-OS federation control plane"
fi
git push origin main

echo
echo "DONE."
echo "Try:"
echo "  ~/AIFT/aift doctor"
echo "  ~/AIFT/aift status"
echo "  ~/AIFT/aift registry"
echo "  ~/AIFT/aift graph"
