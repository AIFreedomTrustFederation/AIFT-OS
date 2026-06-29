#!/usr/bin/env sh
set -eu
. "$AIFT_OS_HOME/runtime/common.sh"

echo "AIFT-OS Doctor"
echo "root: $AIFT_ROOT"
echo "os:   $AIFT_OS_HOME"

[ -d "$AIFT_ROOT" ] || aift_die "Missing workspace root"
[ -d "$AIFT_OS_HOME/.git" ] || aift_die "AIFT-OS is not a git repo"

for d in bin commands runtime install scripts manifests registry intelligence reports docs schemas templates examples logs var; do
  [ -d "$AIFT_OS_HOME/$d" ] || aift_die "Missing directory: $d"
done

bad="$(find "$AIFT_OS_HOME" -type d \( -path "$AIFT_OS_HOME/runtime/runtime" -o -path "$AIFT_OS_HOME/scripts/scripts" -o -path "$AIFT_OS_HOME/manifests/manifests" \) -print)"
[ -z "$bad" ] || aift_die "Doubled folders remain: $bad"

echo "OK: control plane layout healthy"
