#!/usr/bin/env bash
# no-harness: phase bootstrap script; intentionally standalone runtime check utility
set -euo pipefail

cd "$HOME/AIFT/AIFT-OS"

go test ./...
go build -o "$HOME/.local/bin/aift" ./cmd/aift
hash -r
aift services
aift start
aift runtime status
aift runtime ready
aift verify
git status --short
