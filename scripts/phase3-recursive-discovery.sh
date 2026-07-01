#!/usr/bin/env bash
# no-harness: phase bootstrap script; intentionally standalone discovery utility
set -euo pipefail

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"

REPOS="$OS/registry/repos/repos.tsv"
MODULES="$OS/registry/modules/modules.tsv"
COMMANDS="$OS/registry/commands/commands.tsv"
GRAPH="$OS/registry/graphs/federation-discovery.dot"

mkdir -p "$OS/registry/repos"
mkdir -p "$OS/registry/modules"
mkdir -p "$OS/registry/commands"
mkdir -p "$OS/registry/graphs"
mkdir -p "$OS/runtime/logs"

printf "name\tpath\tbranch\tstate\tmanifest\tremote\n" > "$REPOS"
printf "repo\tmodule\tpath\tkind\tmanager\tmanifest\n" > "$MODULES"
printf "repo\tmodule\tcommand\tvalue\n" > "$COMMANDS"

echo "digraph AIFT_Federation_Discovery {" > "$GRAPH"
echo "  rankdir=LR;" >> "$GRAPH"
echo "  federation [label=\"AIFT Federation\"];" >> "$GRAPH"

find "$ROOT" -mindepth 2 -maxdepth 2 -type d -name .git | sort | while read -r gitdir; do
  repo_path="$(dirname "$gitdir")"
  repo_name="$(basename "$repo_path")"

  branch="$(git -C "$repo_path" branch --show-current 2>/dev/null || true)"
  remote="$(git -C "$repo_path" remote get-url origin 2>/dev/null || true)"
  status="$(git -C "$repo_path" status --short 2>/dev/null || true)"

  if [ -z "$branch" ]; then
    branch="unknown"
  fi

  if [ -z "$status" ]; then
    state="clean"
  else
    state="dirty"
  fi

  if [ -f "$repo_path/aift.repo.json" ] || [ -f "$repo_path/.aift/module.json" ]; then
    manifest="valid"
  else
    manifest="missing"
  fi

  printf "%s\t%s\t%s\t%s\t%s\t%s\n" "$repo_name" "$repo_path" "$branch" "$state" "$manifest" "$remote" >> "$REPOS"

  safe_repo="$(echo "$repo_name" | tr -c 'A-Za-z0-9_' '_')"
  echo "  federation -> repo_$safe_repo;" >> "$GRAPH"
  echo "  repo_$safe_repo [label=\"$repo_name\"];" >> "$GRAPH"

  find "$repo_path" \
    \( -name .git -o -name node_modules -o -name .next -o -name dist -o -name build -o -name vendor -o -name runtime -o -name registry -o -name reports \) -prune \
    -o -type f \( -name package.json -o -name go.mod -o -name Cargo.toml -o -name pyproject.toml -o -name pnpm-workspace.yaml -o -name Makefile \) -print \
    | sort | while read -r manifest_file; do

      module_path="$(dirname "$manifest_file")"
      module_name="$(basename "$module_path")"
      manifest_name="$(basename "$manifest_file")"

      kind="module"
      manager="unknown"

      case "$manifest_name" in
        package.json)
          manager="node"
          ;;
        go.mod)
          manager="go"
          ;;
        Cargo.toml)
          manager="cargo"
          ;;
        pyproject.toml)
          manager="python"
          ;;
        pnpm-workspace.yaml)
          manager="pnpm-workspace"
          kind="workspace"
          ;;
        Makefile)
          manager="make"
          ;;
      esac

      printf "%s\t%s\t%s\t%s\t%s\t%s\n" "$repo_name" "$module_name" "$module_path" "$kind" "$manager" "$manifest_name" >> "$MODULES"

      safe_module="$(echo "${repo_name}_${module_name}_${manager}" | tr -c 'A-Za-z0-9_' '_')"
      echo "  repo_$safe_repo -> module_$safe_module;" >> "$GRAPH"
      echo "  module_$safe_module [label=\"$module_name ($manager)\"];" >> "$GRAPH"

      if [ "$manifest_name" = "package.json" ]; then
        node -e '
const fs = require("fs");
const file = process.argv[1];
const repo = process.argv[2];
const mod = process.argv[3];
try {
  const pkg = JSON.parse(fs.readFileSync(file, "utf8"));
  const scripts = pkg.scripts || {};
  for (const [name, value] of Object.entries(scripts)) {
    console.log(`${repo}\t${mod}\t${name}\t${value}`);
  }
} catch (err) {}
' "$manifest_file" "$repo_name" "$module_name" >> "$COMMANDS" 2>/dev/null || true
      fi

      if [ "$manifest_name" = "Makefile" ]; then
        grep -E '^[A-Za-z0-9_.-]+:' "$manifest_file" 2>/dev/null \
          | sed 's/:.*//' \
          | while read -r target; do
              printf "%s\t%s\t%s\t%s\n" "$repo_name" "$module_name" "$target" "make $target" >> "$COMMANDS"
            done || true
      fi
    done
done

echo "}" >> "$GRAPH"

sort -u "$REPOS" -o "$REPOS"
sort -u "$MODULES" -o "$MODULES"
sort -u "$COMMANDS" -o "$COMMANDS"

echo "Wrote $REPOS"
echo "Wrote $MODULES"
echo "Wrote $COMMANDS"
echo "Wrote $GRAPH"
