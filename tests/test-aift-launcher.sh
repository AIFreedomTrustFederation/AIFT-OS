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
output="$(AIFT_OS_HOME="$TMP/checkout" PATH="$TMP/mock-bin:$PATH" sh "$TMP/checkout/aift-os.sh" status)"
[[ "$output" == "launched:status" ]] || { echo "unexpected first-launch output: $output" >&2; exit 1; }
[[ -x "$TMP/checkout/bin/aiftd" ]]

# A second launch must reuse the executable without requiring Go.
output="$(AIFT_OS_HOME="$TMP/checkout" PATH="/usr/bin:/bin" sh "$TMP/checkout/aift-os.sh" verify)"
[[ "$output" == "launched:verify" ]] || { echo "unexpected reuse output: $output" >&2; exit 1; }

echo "PASS: fresh launcher creates bin/ and reuses built executable"
