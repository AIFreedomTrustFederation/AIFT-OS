#!/usr/bin/env bash
# Fail closed when the source under test is not the workflow's exact checkout.
# no-harness: read-only Git HEAD/status inspection; no AIFT mutation or orchestration.
set -euo pipefail
if [[ -z "${GITHUB_SHA:-}" || ! "${GITHUB_SHA}" =~ ^[0-9a-f]{40}$ ]]; then
  echo 'Missing or invalid GITHUB_SHA' >&2
  exit 1
fi
if [[ -z "${GITHUB_WORKSPACE:-}" || ! -d "$GITHUB_WORKSPACE" ]]; then
  echo 'Missing or invalid GITHUB_WORKSPACE' >&2
  exit 1
fi
actual=$(git -C "$GITHUB_WORKSPACE" rev-parse --verify HEAD)
if [[ "$actual" != "$GITHUB_SHA" ]]; then
  echo 'Collector checkout HEAD does not match GITHUB_SHA' >&2
  exit 1
fi
if [[ -n "$(git -C "$GITHUB_WORKSPACE" status --porcelain=v1 --untracked-files=all)" ]]; then
  echo 'Collector checkout contains tracked or untracked changes' >&2
  exit 1
fi
