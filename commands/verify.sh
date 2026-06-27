#!/usr/bin/env sh
set -eu
"$AIFT_OS_HOME/aift-os.sh" doctor
"$AIFT_OS_HOME/aift-os.sh" registry
"$AIFT_OS_HOME/aift-os.sh" graph
echo "OK: federation verified"
