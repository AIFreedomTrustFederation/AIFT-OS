#!/usr/bin/env sh
set -eu

"$AIFT_OS_HOME/aift-os.sh" doctor
"$AIFT_OS_HOME/aift-os.sh" manifest
"$AIFT_OS_HOME/aift-os.sh" registry
"$AIFT_OS_HOME/aift-os.sh" graph
"$AIFT_OS_HOME/aift-os.sh" deps
"$AIFT_OS_HOME/aift-os.sh" dashboard

echo "OK: federation verified"
