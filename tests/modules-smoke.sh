#!/usr/bin/env sh
set -eu
go test ./...
go build -o bin/aiftd ./cmd/aift
bin/aiftd modules init-all >/dev/null
bin/aiftd modules scan >/dev/null
bin/aiftd modules list >/dev/null
bin/aiftd modules repo AIFT-OS >/dev/null
bin/aiftd modules report >/dev/null
test -f registry/modules.json
test -f reports/modules.md
test -f .aift/module.json
echo "OK: modules smoke passed"
