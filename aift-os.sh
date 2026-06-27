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
