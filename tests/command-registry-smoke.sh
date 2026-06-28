#!/usr/bin/env sh
set -eu

go test ./...
go build -o bin/aiftd ./cmd/aift

bin/aiftd help >/dev/null
bin/aiftd discovery list >/dev/null
bin/aiftd kernel-registry list >/dev/null
bin/aiftd event-bus list >/dev/null
bin/aiftd patch-engine inspect >/dev/null
bin/aiftd capabilities list >/dev/null
bin/aiftd modules list >/dev/null

echo "OK: command registry smoke passed"
