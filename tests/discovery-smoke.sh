#!/usr/bin/env sh
set -eu

go test ./...
go build -o bin/aiftd ./cmd/aift

bin/aiftd discovery scan >/dev/null
bin/aiftd discovery list >/dev/null
bin/aiftd discovery object AIFT-OS >/dev/null
bin/aiftd discovery report >/dev/null

test -f registry/discovery.json
test -f reports/discovery.md

echo "OK: discovery smoke passed"
