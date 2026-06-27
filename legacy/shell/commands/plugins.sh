#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/plugins.sh"

echo "AIFT plugin commands:"
aift_list_plugins || true
