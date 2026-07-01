#!/usr/bin/env bash
# no-harness: phase bootstrap script; intentionally standalone registry utility
set -euo pipefail

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"

mkdir -p "$OS/registry/apis"
mkdir -p "$OS/registry/containers"
mkdir -p "$OS/registry/dependencies"
mkdir -p "$OS/registry/docs"
mkdir -p "$OS/registry/models"
mkdir -p "$OS/registry/providers"
mkdir -p "$OS/registry/services"
mkdir -p "$OS/registry/workflows"
mkdir -p "$OS/registry/graphs"
mkdir -p "$OS/runtime/logs"

REPOS="$OS/registry/repos/repos.tsv"
MODULES="$OS/registry/modules/modules.tsv"
COMMANDS="$OS/registry/commands/commands.tsv"
WORKFLOWS="$OS/registry/workflows/workflows.tsv"
SERVICES="$OS/registry/services/services.tsv"
APIS="$OS/registry/apis/apis.tsv"
CONTAINERS="$OS/registry/containers/containers.tsv"
DOCS="$OS/registry/docs/docs.tsv"
PROVIDERS="$OS/registry/providers/providers.tsv"
MODELS="$OS/registry/models/models.tsv"
DEPS="$OS/registry/dependencies/dependencies.tsv"
GRAPH="$OS/registry/graphs/federation-runtime.dot"

printf "name\tpath\tbranch\tstate\tmanifest\tremote\n" > "$REPOS"
printf "repo\tmodule\tpath\tkind\tmanager\tmanifest\n" > "$MODULES"
printf "repo\tmodule\tcommand\tvalue\n" > "$COMMANDS"
printf "repo\tpath\tkind\tname\n" > "$WORKFLOWS"
printf "repo\tpath\tkind\tname\n" > "$SERVICES"
printf "repo\tpath\tkind\tname\n" > "$APIS"
printf "repo\tpath\tkind\tname\n" > "$CONTAINERS"
printf "repo\tpath\tkind\tname\n" > "$DOCS"
printf "repo\tpath\tkind\tname\n" > "$PROVIDERS"
printf "repo\tpath\tkind\tname\n" > "$MODELS"
printf "repo\tmodule\tmanager\tdependency\n" > "$DEPS"

echo "digraph AIFT_Federation_Runtime {" > "$GRAPH"
echo "  rankdir=LR;" >> "$GRAPH"
echo "  federation [label=\"AIFT Federation Runtime\"];" >> "$GRAPH"

safe_node() {
  printf "%s" "$1" | tr -c 'A-Za-z0-9_' '_'
}

discover_package_scripts() {
  local repo_name="$1"
  local module_name="$2"
  local manifest_file="$3"

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
}

discover_package_dependencies() {
  local repo_name="$1"
  local module_name="$2"
  local manifest_file="$3"

  node -e '
const fs = require("fs");
const file = process.argv[1];
const repo = process.argv[2];
const mod = process.argv[3];

try {
  const pkg = JSON.parse(fs.readFileSync(file, "utf8"));
  const groups = [
    pkg.dependencies || {},
    pkg.devDependencies || {},
    pkg.peerDependencies || {},
    pkg.optionalDependencies || {}
  ];
  const seen = new Set();
  for (const group of groups) {
    for (const name of Object.keys(group)) {
      if (!seen.has(name)) {
        seen.add(name);
        console.log(`${repo}\t${mod}\tnode\t${name}`);
      }
    }
  }
} catch (err) {}
' "$manifest_file" "$repo_name" "$module_name" >> "$DEPS" 2>/dev/null || true
}

discover_make_targets() {
  local repo_name="$1"
  local module_name="$2"
  local makefile="$3"

  grep -E '^[A-Za-z0-9_.-]+:' "$makefile" 2>/dev/null \
    | sed 's/:.*//' \
    | while read -r target; do
        printf "%s\t%s\t%s\t%s\n" "$repo_name" "$module_name" "$target" "make $target" >> "$COMMANDS"
      done || true
}

discover_go_dependencies() {
  local repo_name="$1"
  local module_name="$2"
  local gomod="$3"

  grep -E '^[[:space:]]*[A-Za-z0-9_.:/-]+\.[A-Za-z0-9_.:/-]+' "$gomod" 2>/dev/null \
    | awk '{print $1}' \
    | while read -r dep; do
        printf "%s\t%s\t%s\t%s\n" "$repo_name" "$module_name" "go" "$dep" >> "$DEPS"
      done || true
}

discover_python_dependencies() {
  local repo_name="$1"
  local module_name="$2"
  local pyproject="$3"

  grep -E '^[[:space:]]*"[A-Za-z0-9_.-]+' "$pyproject" 2>/dev/null \
    | sed 's/[",].*//' \
    | tr -d ' "' \
    | while read -r dep; do
        if [ -n "$dep" ]; then
          printf "%s\t%s\t%s\t%s\n" "$repo_name" "$module_name" "python" "$dep" >> "$DEPS"
        fi
      done || true
}

discover_rust_dependencies() {
  local repo_name="$1"
  local module_name="$2"
  local cargo="$3"

  awk '
    /^\[dependencies\]/ { in_deps=1; next }
    /^\[/ { in_deps=0 }
    in_deps && /^[A-Za-z0-9_-]+/ { print $1 }
  ' "$cargo" 2>/dev/null \
    | sed 's/=.*//' \
    | while read -r dep; do
        if [ -n "$dep" ]; then
          printf "%s\t%s\t%s\t%s\n" "$repo_name" "$module_name" "cargo" "$dep" >> "$DEPS"
        fi
      done || true
}

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

  repo_node="repo_$(safe_node "$repo_name")"
  echo "  federation -> $repo_node;" >> "$GRAPH"
  echo "  $repo_node [label=\"$repo_name\"];" >> "$GRAPH"

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

      module_node="module_$(safe_node "${repo_name}_${module_name}_${manager}")"
      echo "  $repo_node -> $module_node;" >> "$GRAPH"
      echo "  $module_node [label=\"$module_name ($manager)\"];" >> "$GRAPH"

      if [ "$manifest_name" = "package.json" ]; then
        discover_package_scripts "$repo_name" "$module_name" "$manifest_file"
        discover_package_dependencies "$repo_name" "$module_name" "$manifest_file"
      fi

      if [ "$manifest_name" = "Makefile" ]; then
        discover_make_targets "$repo_name" "$module_name" "$manifest_file"
      fi

      if [ "$manifest_name" = "go.mod" ]; then
        discover_go_dependencies "$repo_name" "$module_name" "$manifest_file"
      fi

      if [ "$manifest_name" = "pyproject.toml" ]; then
        discover_python_dependencies "$repo_name" "$module_name" "$manifest_file"
      fi

      if [ "$manifest_name" = "Cargo.toml" ]; then
        discover_rust_dependencies "$repo_name" "$module_name" "$manifest_file"
      fi
    done

  find "$repo_path/.github/workflows" -type f \( -name "*.yml" -o -name "*.yaml" \) 2>/dev/null | sort | while read -r f; do
    printf "%s\t%s\t%s\t%s\n" "$repo_name" "$f" "github-actions" "$(basename "$f")" >> "$WORKFLOWS"
  done

  find "$repo_path" \
    \( -name .git -o -name node_modules -o -name .next -o -name dist -o -name build -o -name vendor \) -prune \
    -o -type f \( -name Dockerfile -o -name docker-compose.yml -o -name compose.yml -o -name compose.yaml \) -print \
    | sort | while read -r f; do
      printf "%s\t%s\t%s\t%s\n" "$repo_name" "$f" "container" "$(basename "$f")" >> "$CONTAINERS"
    done

  find "$repo_path" \
    \( -name .git -o -name node_modules -o -name .next -o -name dist -o -name build -o -name vendor \) -prune \
    -o -type f \( -name "*openapi*.json" -o -name "*openapi*.yaml" -o -name "*openapi*.yml" -o -name "*.graphql" -o -name "*.gql" \) -print \
    | sort | while read -r f; do
      case "$f" in
        *.graphql|*.gql)
          kind="graphql"
          ;;
        *)
          kind="openapi"
          ;;
      esac
      printf "%s\t%s\t%s\t%s\n" "$repo_name" "$f" "$kind" "$(basename "$f")" >> "$APIS"
    done

  find "$repo_path" \
    \( -name .git -o -name node_modules -o -name .next -o -name dist -o -name build -o -name vendor \) -prune \
    -o -type f \( -name "*.md" -o -name "*.mdx" \) -print \
    | sort | while read -r f; do
      printf "%s\t%s\t%s\t%s\n" "$repo_name" "$f" "documentation" "$(basename "$f")" >> "$DOCS"
    done

  grep -RIl "ollama\|llama.cpp\|vllm\|openai\|provider" "$repo_path" \
    --exclude-dir=.git \
    --exclude-dir=node_modules \
    --exclude-dir=.next \
    --exclude-dir=dist \
    --exclude-dir=build \
    --exclude-dir=vendor \
    2>/dev/null | sort | while read -r f; do
      printf "%s\t%s\t%s\t%s\n" "$repo_name" "$f" "ai-provider-indicator" "$(basename "$f")" >> "$PROVIDERS"
    done

  grep -RIl "model\|embedding\|chat\|completion\|generation" "$repo_path" \
    --exclude-dir=.git \
    --exclude-dir=node_modules \
    --exclude-dir=.next \
    --exclude-dir=dist \
    --exclude-dir=build \
    --exclude-dir=vendor \
    2>/dev/null | sort | while read -r f; do
      printf "%s\t%s\t%s\t%s\n" "$repo_name" "$f" "model-indicator" "$(basename "$f")" >> "$MODELS"
    done

  grep -RIl "server\|listen\|port\|route\|handler\|endpoint" "$repo_path" \
    --exclude-dir=.git \
    --exclude-dir=node_modules \
    --exclude-dir=.next \
    --exclude-dir=dist \
    --exclude-dir=build \
    --exclude-dir=vendor \
    2>/dev/null | sort | while read -r f; do
      printf "%s\t%s\t%s\t%s\n" "$repo_name" "$f" "service-indicator" "$(basename "$f")" >> "$SERVICES"
    done
done

echo "}" >> "$GRAPH"

sort -u "$REPOS" -o "$REPOS"
sort -u "$MODULES" -o "$MODULES"
sort -u "$COMMANDS" -o "$COMMANDS"
sort -u "$WORKFLOWS" -o "$WORKFLOWS"
sort -u "$SERVICES" -o "$SERVICES"
sort -u "$APIS" -o "$APIS"
sort -u "$CONTAINERS" -o "$CONTAINERS"
sort -u "$DOCS" -o "$DOCS"
sort -u "$PROVIDERS" -o "$PROVIDERS"
sort -u "$MODELS" -o "$MODELS"
sort -u "$DEPS" -o "$DEPS"

echo "Wrote federation runtime graph registry."
