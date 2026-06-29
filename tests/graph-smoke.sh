#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" graph >/dev/null
"$ROOT/aift" graph summary >/dev/null
"$ROOT/aift" graph repo AIFT-OS >/dev/null
"$ROOT/aift" graph type Repository >/dev/null
"$ROOT/aift" graph status planned >/dev/null
"$ROOT/aift" verify >/dev/null

test -f registry/graph.json
test -f registry/graph.mermaid
test -f registry/graph.dot
test -f registry/graph.graphml
test -f registry/graph.cypher
test -f registry/graph.rdf

test -f reports/graph.md
test -f reports/graph-summary.md
test -f reports/dependency-tree.md
test -f reports/orphaned-capabilities.md
test -f reports/planned-vs-running.md
test -f reports/service-map.md

echo "OK: graph smoke passed"
