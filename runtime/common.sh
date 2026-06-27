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
