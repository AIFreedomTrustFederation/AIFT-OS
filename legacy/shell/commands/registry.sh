#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/federation.sh"
out="$(aift_registry_json)"
echo "Wrote $out"
