#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/federation.sh"

out="$(aift_dashboard)"
echo "Wrote $out"
