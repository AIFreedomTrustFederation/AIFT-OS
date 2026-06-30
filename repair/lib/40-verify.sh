#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail
cd "${AIFT_ROOT:-$HOME/AIFT}/AIFT-OS"
gofmt -w cmd/aift/main.go internal/doctor/*.go
go test ./...
go build -o "$HOME/.local/bin/aift" ./cmd/aift
hash -r
aift doctor
aift doctor repair
aift doctor git
aift verify
