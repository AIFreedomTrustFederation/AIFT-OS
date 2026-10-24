#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

mkdir -p "$TMP/checkout/scripts" "$TMP/checkout/space dir" "$TMP/checkout/.git/hooks" "$TMP/checkout/node_modules" "$TMP/mock-bin"
cp "$ROOT/scripts/dev.sh" "$TMP/checkout/scripts/dev.sh"
printf '#!/usr/bin/env bash\necho valid\n' > "$TMP/checkout/space dir/valid script.sh"
printf 'if\n' > "$TMP/checkout/.git/hooks/ignored.sh"
printf 'if\n' > "$TMP/checkout/node_modules/ignored.sh"

# Valid shell files with spaces in their names pass; excluded paths stay excluded.
bash "$TMP/checkout/scripts/dev.sh" shell >/dev/null

# Invalid maintained shell syntax must cause a nonzero exit status.
printf 'if\n' > "$TMP/checkout/space dir/broken script.sh"
if bash "$TMP/checkout/scripts/dev.sh" shell >"$TMP/invalid.log" 2>&1; then
  echo 'FAIL: malformed shell file was accepted' >&2
  exit 1
fi
grep -F 'broken script.sh' "$TMP/invalid.log" >/dev/null
rm "$TMP/checkout/space dir/broken script.sh"

# A failed directory traversal must never be reported as a passing shell check.
cat > "$TMP/mock-bin/find" <<'EOF'
#!/bin/sh
echo 'injected find failure' >&2
exit 42
EOF
chmod +x "$TMP/mock-bin/find"
if PATH="$TMP/mock-bin:$PATH" bash "$TMP/checkout/scripts/dev.sh" shell >"$TMP/find.log" 2>&1; then
  echo 'FAIL: find traversal failure was accepted' >&2
  exit 1
fi
grep -F 'injected find failure' "$TMP/find.log" >/dev/null

echo 'PASS: shell syntax, exclusions, spaces, and find failures'
