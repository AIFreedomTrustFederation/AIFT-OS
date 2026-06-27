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
