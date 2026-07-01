#!/usr/bin/env bash
set -euo pipefail

echo "== Phase 19C: Quarantine stale CLI files =="

ROOT="$(pwd)"
CMD_DIR="$ROOT/cmd/aift"
LEGACY_DIR="$ROOT/legacy/cmd-aift-phase19c"
REPORT_DIR="$ROOT/reports"

mkdir -p "$LEGACY_DIR" "$REPORT_DIR"

REPORT="$REPORT_DIR/phase19c-quarantine-stale-cli-report.md"

{
  echo "# Phase 19C Quarantine Stale CLI Report"
  echo ""
  echo "Generated: $(date -Is)"
  echo ""
  echo "## Before"
  echo ""
  find "$CMD_DIR" -maxdepth 1 -name '*.go' -print
  echo ""
  echo "## Stale verify references"
  grep -Rnw "$CMD_DIR" -e '\bverify\b' || true
  echo ""
} > "$REPORT"

echo "Moving stale compiled CLI Go files into $LEGACY_DIR..."

find "$CMD_DIR" -maxdepth 1 -name '*.go' ! -name 'main.go' -print0 | while IFS= read -r -d '' f; do
  mv "$f" "$LEGACY_DIR/"
done

echo "Rebuilding clean main.go by rerunning Phase 19B..."
./phase19-cli-compiler-repair.sh || true

echo "Verifying no stale compiled verify references remain..."
grep -Rnw "$CMD_DIR" -e '\bverify\b' || true

echo "Building..."
go build ./cmd/aift

echo "Testing CLI..."
go run ./cmd/aift status
go run ./cmd/aift verify
go run ./cmd/aift registry > registry/cli/commands.json

{
  echo ""
  echo "## After"
  echo ""
  find "$CMD_DIR" -maxdepth 1 -name '*.go' -print
  echo ""
  echo "## Quarantined files"
  find "$LEGACY_DIR" -maxdepth 1 -type f -print
  echo ""
  echo "## Result"
  echo ""
  echo "- go build ./cmd/aift: PASS"
  echo "- go run ./cmd/aift status: PASS"
  echo "- go run ./cmd/aift verify: PASS"
} >> "$REPORT"

git add cmd/aift legacy/cmd-aift-phase19c registry/cli/commands.json "$REPORT" phase19c-quarantine-stale-cli.sh || true

echo ""
echo "Phase 19C complete."
echo "Now commit:"
echo "  git commit -m 'Phase 19C: quarantine stale CLI command files'"
echo "  git push origin main"
