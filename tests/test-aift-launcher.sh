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
[[ -x "$TMP/checkout/bin/aift" ]]

# A second launch must reuse the executable without requiring Go.
output="$(cd "$TMP" && env -u AIFT_OS_HOME -u AIFT_ROOT PATH="/usr/bin:/bin" sh "$TMP/checkout/aift-os.sh" verify)"
[[ "$output" == "launched:verify" ]] || { echo "unexpected reuse output: $output" >&2; exit 1; }

# Explicit AIFT_ROOT must resolve to <root>/AIFT-OS when AIFT_OS_HOME is unset.
mkdir -p "$TMP/custom-root/AIFT-OS"
cp "$ROOT/aift-os.sh" "$TMP/custom-root/AIFT-OS/aift-os.sh"
output="$(cd "$TMP" && env -u AIFT_OS_HOME AIFT_ROOT="$TMP/custom-root" PATH="$TMP/mock-bin:$PATH" sh "$TMP/custom-root/AIFT-OS/aift-os.sh" status)"
[[ "$output" == "launched:status" ]] || { echo "unexpected AIFT_ROOT output: $output" >&2; exit 1; }
[[ -x "$TMP/custom-root/AIFT-OS/bin/aift" ]]

# AIFT_OS_HOME takes precedence over AIFT_ROOT.
mkdir -p "$TMP/explicit-home"
cp "$ROOT/aift-os.sh" "$TMP/explicit-home/aift-os.sh"
output="$(cd "$TMP" && AIFT_OS_HOME="$TMP/explicit-home" AIFT_ROOT="$TMP/nonexistent-root" PATH="$TMP/mock-bin:$PATH" sh "$TMP/explicit-home/aift-os.sh" verify)"
[[ "$output" == "launched:verify" ]] || { echo "unexpected AIFT_OS_HOME output: $output" >&2; exit 1; }
[[ -x "$TMP/explicit-home/bin/aift" ]]

# A normal `make build` creates both binaries. The launcher must use the CLI,
# not the MoBox UXI HTTP server at bin/aiftd, even when both are executable.
mkdir -p "$TMP/prebuilt/bin"
cp "$ROOT/aift-os.sh" "$TMP/prebuilt/aift-os.sh"
printf '%s\n' '#!/usr/bin/env sh' 'printf "prebuilt-cli:%s\n" "$*"' > "$TMP/prebuilt/bin/aift"
printf '%s\n' '#!/usr/bin/env sh' 'echo ERROR-server-executed >&2; exit 99' > "$TMP/prebuilt/bin/aiftd"
chmod +x "$TMP/prebuilt/bin/aift" "$TMP/prebuilt/bin/aiftd"
output="$(cd "$TMP" && AIFT_OS_HOME="$TMP/prebuilt" PATH="/usr/bin:/bin" sh "$TMP/prebuilt/aift-os.sh" status)"
[[ "$output" == "prebuilt-cli:status" ]] || { echo "launcher selected wrong binary: $output" >&2; exit 1; }

# A preexisting daemon must never be overwritten by first-run CLI bootstrapping.
mkdir -p "$TMP/daemon-only/bin"
cp "$ROOT/aift-os.sh" "$TMP/daemon-only/aift-os.sh"
printf '%s\n' '#!/usr/bin/env sh' 'echo daemon-intact' > "$TMP/daemon-only/bin/aiftd"
chmod +x "$TMP/daemon-only/bin/aiftd"
output="$(cd "$TMP" && AIFT_OS_HOME="$TMP/daemon-only" PATH="$TMP/mock-bin:$PATH" sh "$TMP/daemon-only/aift-os.sh" verify)"
[[ "$output" == "launched:verify" ]] || { echo "unexpected daemon-only output: $output" >&2; exit 1; }
[[ -x "$TMP/daemon-only/bin/aift" ]]
[[ "$(sh "$TMP/daemon-only/bin/aiftd")" == "daemon-intact" ]] || { echo "daemon overwritten" >&2; exit 1; }

echo "PASS: default checkout, overrides, CLI reuse, build coexistence, and daemon preservation"
