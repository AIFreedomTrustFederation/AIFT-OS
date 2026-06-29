#!/usr/bin/env sh
set -eu

go test ./...
go build -o bin/aiftd ./cmd/aift

bin/aiftd kernel-registry scan >/dev/null
bin/aiftd kernel-registry list >/dev/null
bin/aiftd kernel-registry object federation.local >/dev/null
bin/aiftd kernel-registry report >/dev/null

test -f registry/kernel-registry.json
test -f reports/kernel-registry.md

echo "OK: kernel registry smoke passed"
