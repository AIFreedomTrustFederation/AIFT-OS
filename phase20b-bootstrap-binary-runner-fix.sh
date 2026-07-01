#!/usr/bin/env bash
set -euo pipefail

echo "== Phase 20B: Bootstrap Binary Runner Fix =="

ROOT="$(pwd)"
BIN_DIR="$ROOT/bin"
CMD_DIR="$ROOT/cmd/aift"
REPORT_DIR="$ROOT/reports"
REGISTRY_DIR="$ROOT/registry/cli"
BOOTSTRAP_DIR="$ROOT/registry/bootstrap"

mkdir -p "$BIN_DIR" "$REPORT_DIR" "$REGISTRY_DIR" "$BOOTSTRAP_DIR"

REPORT="$REPORT_DIR/phase20b-bootstrap-binary-runner-fix-report.md"
REGISTRY="$REGISTRY_DIR/commands.json"
BOOTSTRAP="$BOOTSTRAP_DIR/federation-bootstrap.json"
BIN="$BIN_DIR/aift"

{
  echo "# Phase 20B Bootstrap Binary Runner Fix Report"
  echo ""
  echo "Generated: $(date -Is)"
  echo ""
} > "$REPORT"

echo "Building actual CLI binary..."
go build -o "$BIN" ./cmd/aift

echo "Testing binary help..."
"$BIN" help

echo "Writing registry from binary..."
"$BIN" registry > "$REGISTRY"

echo "Writing bootstrap discovery from binary..."
"$BIN" bootstrap > "$BOOTSTRAP"

echo "Testing binary status..."
"$BIN" status

echo "Testing binary verify..."
VERIFY_STATUS="PASS"
"$BIN" verify || VERIFY_STATUS="FAIL"

echo "Running full Go test suite..."
TEST_STATUS="PASS"
go test ./... || TEST_STATUS="FAIL"

{
  echo ""
  echo "## Results"
  echo ""
  echo "- go build -o bin/aift ./cmd/aift: PASS"
  echo "- bin/aift help: PASS"
  echo "- bin/aift registry: PASS"
  echo "- bin/aift bootstrap: PASS"
  echo "- bin/aift status: PASS"
  echo "- bin/aift verify: $VERIFY_STATUS"
  echo "- go test ./...: $TEST_STATUS"
  echo ""
  echo "## Bootstrap Discovery"
  echo ""
  cat "$BOOTSTRAP"
} >> "$REPORT"

git add bin/aift "$REGISTRY" "$BOOTSTRAP" "$REPORT" phase20b-bootstrap-binary-runner-fix.sh || true

echo ""
echo "Phase 20B complete."
echo "Report: $REPORT"
echo "Registry: $REGISTRY"
echo "Bootstrap: $BOOTSTRAP"
echo "Binary: $BIN"
echo ""
echo "Manual checks:"
echo "  ./bin/aift help"
echo "  ./bin/aift status"
echo "  ./bin/aift verify"
echo "  ./bin/aift bootstrap"
echo ""
echo "Commit:"
echo "  git commit -m 'Phase 20B: use built binary for bootstrap validation'"
echo "  git push origin main"
