#!/data/data/com.termux/files/usr/bin/bash
set -u

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="$ROOT/AIFT-OS"

cd "$OS" || exit 1

find internal tests -name "*_test.go" -type f | while read -r file; do
  if grep -Eq 'config\.Config\{.*(Root|OS|Stdout|Stderr)' "$file" 2>/dev/null; then
    echo "Removing bad guessed config struct test: $file"
    rm -f "$file"
    continue
  fi

  if grep -Eq 'Root:|OS:|Stdout:|Stderr:' "$file" 2>/dev/null && grep -q 'config.Config' "$file" 2>/dev/null; then
    echo "Removing bad guessed config struct test: $file"
    rm -f "$file"
    continue
  fi
done
