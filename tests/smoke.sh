#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."

./aift-os.sh doctor
./aift-os.sh status >/dev/null
./aift-os.sh registry >/dev/null
./aift-os.sh graph >/dev/null
./aift-os.sh deps >/dev/null

echo "OK: smoke tests passed"
