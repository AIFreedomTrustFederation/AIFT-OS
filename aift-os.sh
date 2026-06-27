#!/usr/bin/env bash
set -euo pipefail

source "$HOME/AIFT/runtime/common.sh"
source "$HOME/AIFT/runtime/status.sh"
source "$HOME/AIFT/runtime/registry.sh"
source "$HOME/AIFT/runtime/intel.sh"
source "$HOME/AIFT/runtime/graph.sh"
source "$HOME/AIFT/runtime/dashboard.sh"
source "$HOME/AIFT/runtime/doctor.sh"
source "$HOME/AIFT/runtime/pull.sh"
source "$HOME/AIFT/runtime/verify.sh"
source "$HOME/AIFT/runtime/push.sh"
source "$HOME/AIFT/runtime/workflow.sh"

case "${1:-help}" in
  status) aift_status ;;
  registry) aift_registry ;;
  intelligence|intel) aift_intelligence ;;
  graph) aift_graph ;;
  dashboard) aift_dashboard ;;
  doctor) aift_doctor ;;
  pull) aift_pull ;;
  verify) aift_verify ;;
  push) aift_push ;;
  update) aift_update ;;
  sync)
    aift_registry
    aift_intelligence
    aift_dashboard
    aift_graph
    ;;
  help|*)
    echo "AIFT Runtime OS"
    echo
    echo "Commands:"
    echo "  ~/AIFT/aift-os.sh status"
    echo "  ~/AIFT/aift-os.sh registry"
    echo "  ~/AIFT/aift-os.sh intelligence"
    echo "  ~/AIFT/aift-os.sh graph"
    echo "  ~/AIFT/aift-os.sh dashboard"
    echo "  ~/AIFT/aift-os.sh doctor"
    echo "  ~/AIFT/aift-os.sh pull"
    echo "  ~/AIFT/aift-os.sh verify"
    echo "  ~/AIFT/aift-os.sh push"
    echo "  ~/AIFT/aift-os.sh update"
    echo "  ~/AIFT/aift-os.sh sync"
    ;;
esac
