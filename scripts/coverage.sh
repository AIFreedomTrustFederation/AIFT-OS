#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

mkdir -p reports

echo "Running tests with coverage..."
go test ./... -coverprofile=coverage.out

echo ""
echo "Generating coverage report..."
go tool cover -func=coverage.out > reports/coverage.txt

echo ""
echo "--- Per-function coverage ---"
cat reports/coverage.txt

echo ""
echo "--- Summary ---"
tail -1 reports/coverage.txt

echo ""
echo "Coverage profile: coverage.out"
echo "Coverage report:  reports/coverage.txt"
echo ""
echo "To view HTML coverage: go tool cover -html=coverage.out"
