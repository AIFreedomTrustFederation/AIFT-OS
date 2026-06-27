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
