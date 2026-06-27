#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/federation.sh"
out="$(aift_graph_markdown)"
echo "Wrote $out"
