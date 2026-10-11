#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

mkdir -p "$TMP/checkout" "$TMP/mock-bin"
cp "$ROOT/aift-os.sh" "$TMP/checkout/aift-os.sh"

cat > "$TMP/mock-bin/go" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
[[ "$1" == build && "$2" == -o && "$4" == ./cmd/aift ]]
[[ -d "$(dirname "$3")" ]] || { echo "missing output directory" >&2; exit 10; }
printf '%s\n' '#!/usr/bin/env sh' 'printf "launched:%s\\n" "$*"' > "$3"
chmod +x "$3"
SH
chmod +x "$TMP/mock-bin/go"

# No bin/ exists at first launch. The wrapper must create it.
[[ ! -e "$TMP/checkout/bin" ]]
output="$(cd "$TMP" && env -u AIFT_OS_HOME -u AIFT_ROOT PATH="$TMP/mock-bin:$PATH" sh "$TMP/checkout/aift-os.sh" status)"
[[ "$output" == "launched:status" ]] || { echo "unexpected first-launch output: $output" >&2; exit 1; }
[[ -x "$TMP/checkout/bin/aiftd" ]]

# A second launch must reuse the executable without requiring Go.
output="$(cd "$TMP" && env -u AIFT_OS_HOME -u AIFT_ROOT PATH="/usr/bin:/bin" sh "$TMP/checkout/aift-os.sh" verify)"
[[ "$output" == "launched:verify" ]] || { echo "unexpected reuse output: $output" >&2; exit 1; }

# Explicit AIFT_ROOT must resolve to <root>/AIFT-OS when AIFT_OS_HOME is unset.
mkdir -p "$TMP/custom-root/AIFT-OS"
cp "$ROOT/aift-os.sh" "$TMP/custom-root/AIFT-OS/aift-os.sh"
output="$(cd "$TMP" && env -u AIFT_OS_HOME AIFT_ROOT="$TMP/custom-root" PATH="$TMP/mock-bin:$PATH" sh "$TMP/custom-root/AIFT-OS/aift-os.sh" status)"
[[ "$output" == "launched:status" ]] || { echo "unexpected AIFT_ROOT output: $output" >&2; exit 1; }
[[ -x "$TMP/custom-root/AIFT-OS/bin/aiftd" ]]

# AIFT_OS_HOME takes precedence over AIFT_ROOT.
mkdir -p "$TMP/explicit-home"
cp "$ROOT/aift-os.sh" "$TMP/explicit-home/aift-os.sh"
output="$(cd "$TMP" && AIFT_OS_HOME="$TMP/explicit-home" AIFT_ROOT="$TMP/nonexistent-root" PATH="$TMP/mock-bin:$PATH" sh "$TMP/explicit-home/aift-os.sh" verify)"
[[ "$output" == "launched:verify" ]] || { echo "unexpected AIFT_OS_HOME output: $output" >&2; exit 1; }
[[ -x "$TMP/explicit-home/bin/aiftd" ]]

echo "PASS: default checkout, AIFT_ROOT, AIFT_OS_HOME precedence, and binary reuse"
