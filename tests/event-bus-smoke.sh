#!/usr/bin/env sh
set -eu

go test ./...
go build -o bin/aiftd ./cmd/aift

bin/aiftd event-bus publish system.test "event bus smoke" source=smoke >/dev/null
bin/aiftd event-bus list >/dev/null
bin/aiftd event-bus replay system.test >/dev/null
bin/aiftd event-bus report >/dev/null

test -f var/events/event-bus.jsonl
test -f reports/event-bus.md

echo "OK: event bus smoke passed"
