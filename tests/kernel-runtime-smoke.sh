#!/usr/bin/env sh
set -eu

go test ./...
go build -o bin/aiftd ./cmd/aift

bin/aiftd kernel boot >/dev/null
bin/aiftd kernel status >/dev/null
bin/aiftd kernel report >/dev/null

test -f registry/kernel-boot.json
test -f reports/kernel-boot.md

echo "OK: kernel runtime smoke passed"
