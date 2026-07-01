# Phase 20 Cleanup Final Report

Generated: 2026-07-01T06:04:32.9396680-07:00

## CLI Checks

- `go build -o bin/aift ./cmd/aift`: checked
- `go build -o bin/aift.exe ./cmd/aift`: checked on Windows
- `./bin/aift help`: checked through `bin/aift.exe` on Windows
- `./bin/aift status`: checked through `bin/aift.exe` on Windows
- `./bin/aift verify`: checked through `bin/aift.exe` on Windows
- `./bin/aift registry`: checked through `bin/aift.exe` on Windows
- `./bin/aift bootstrap`: checked through `bin/aift.exe` on Windows

## Cleanup Performed

- Corrected CLI argv dispatch to call `run(os.Args[1:])` and ignore `os.Args[0]`.
- Removed manual executable path stripping from CLI argument normalization.
- Confirmed `cmd/aift/main.go` is the only active compiling Go source under `cmd/aift`.
- Preserved stale `cmd/aift` phase backups as `.bak` files.
- Quarantined stale legacy CLI Go files by renaming `legacy/cmd-aift-phase19c/*.go` to `*.go.bak`.
- Quarantined stale root phase-generation scripts with old `verify(cfg)` wiring into `legacy/phase-script-cleanup/`.
- Replaced active `go run ./cmd/aift` script checks with binary-first `bin/aift` checks.
- Added `bin/` to `.gitignore` and removed the tracked `bin/aift` binary from the index.
- Regenerated `registry/cli/commands.json`.
- Regenerated `registry/bootstrap/federation-bootstrap.json`.
- Updated integration tests to validate the active truthful CLI surface only.
- Made Windows path and missing-shell test behavior explicit.

## Test Results

- `go test ./...`: pass

## Notes

The CLI should not claim federation, repo, workflow, runtime, scheduler, event bus, module, or capability command groups are active until real internal implementations are wired and verified.

On Windows, `go build -o bin/aift ./cmd/aift` creates an extensionless PE executable that is not invoked normally by PowerShell. Local Windows verification used `go build -o bin/aift.exe ./cmd/aift`; both binaries remain ignored under `bin/`.
