#!/usr/bin/env bash
# scripts/lib/discovery.sh
# Discover AIFT repo and workspace paths.
# Source this file; it exports AIFT_ROOT, AIFT_OS_HOME, and REPO_ROOT.

set -euo pipefail

# REPO_ROOT: the root of the repository that sourced this file.
# Walk up from the script's directory until we find .git or go.mod.
_discover_repo_root() {
  local dir
  dir="$(cd "$(dirname "${BASH_SOURCE[1]:-${BASH_SOURCE[0]}}")" && pwd)"
  while [ "$dir" != "/" ]; do
    if [ -d "$dir/.git" ] || [ -f "$dir/go.mod" ]; then
      echo "$dir"
      return
    fi
    dir="$(dirname "$dir")"
  done
  echo "$(pwd)"
}

export REPO_ROOT="${REPO_ROOT:-$(_discover_repo_root)}"
export AIFT_ROOT="${AIFT_ROOT:-${HOME}/AIFT}"
export AIFT_OS_HOME="${AIFT_OS_HOME:-${AIFT_ROOT}/AIFT-OS}"
