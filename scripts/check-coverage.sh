#!/usr/bin/env bash
set -euo pipefail

# Check that total test coverage does not drop below the baseline.
# The baseline is stored in coverage-baseline.txt at the repo root.
# To update the baseline intentionally, edit that file.

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

BASELINE_FILE="$REPO_ROOT/coverage-baseline.txt"
if [ ! -f "$BASELINE_FILE" ]; then
  echo "ERROR: $BASELINE_FILE not found"
  exit 1
fi

BASELINE=$(cat "$BASELINE_FILE" | tr -d '[:space:]')

echo "Running tests with coverage..."
go test ./... -coverprofile=coverage.out

mkdir -p reports
go tool cover -func=coverage.out > reports/coverage.txt

# Extract total coverage percentage from the last line
TOTAL=$(tail -1 reports/coverage.txt | awk '{print $NF}' | tr -d '%')

echo ""
echo "Coverage baseline: ${BASELINE}%"
echo "Current coverage:  ${TOTAL}%"

# Compare using awk for floating-point comparison
PASS=$(awk "BEGIN { print ($TOTAL >= $BASELINE) ? 1 : 0 }")

if [ "$PASS" -eq 0 ]; then
  echo ""
  echo "FAIL: Coverage ${TOTAL}% is below baseline ${BASELINE}%"
  echo "If this drop is intentional, update coverage-baseline.txt"
  exit 1
fi

echo ""
echo "PASS: Coverage ${TOTAL}% meets baseline ${BASELINE}%"
