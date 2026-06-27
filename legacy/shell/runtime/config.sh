#!/usr/bin/env sh
set -eu

AIFT_ROOT="${AIFT_ROOT:-$HOME/AIFT}"
AIFT_OS_HOME="${AIFT_OS_HOME:-$AIFT_ROOT/AIFT-OS}"

if [ -f "$AIFT_OS_HOME/config/aift-os.env" ]; then
  # shellcheck disable=SC1090
  . "$AIFT_OS_HOME/config/aift-os.env"
fi

export AIFT_ROOT AIFT_OS_HOME
