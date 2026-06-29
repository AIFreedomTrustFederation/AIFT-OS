#!/usr/bin/env sh
set -eu

go test ./...
go build -o bin/aiftd ./cmd/aift

bin/aiftd patch-engine inspect >/dev/null
bin/aiftd patch-engine plan >/dev/null
bin/aiftd patch-engine validate >/dev/null

test -f registry/patch-plan.json
test -f registry/patch-result.json
test -f reports/patch-plan.md

echo "OK: patch engine smoke passed"
