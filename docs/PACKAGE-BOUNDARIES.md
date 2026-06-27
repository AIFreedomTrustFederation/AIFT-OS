# AIFT-OS Package Boundaries

AIFT-OS is the Federation Control Plane. It orchestrates sovereign repositories without absorbing them.

## CLI

- `cmd/aift` owns command parsing and user-facing commands only.
- CLI code should call internal packages, not implement business logic directly.

## Configuration

- `internal/config` resolves workspace paths and runtime settings.
- No package should hard-code `~/AIFT` directly.

## Workspace and Git

- `internal/workspace` discovers repositories.
- `internal/gitx` wraps Git commands.
- Higher packages should not shell out to Git directly.

## Manifests and Registry

- `internal/manifests` owns `.aift/repo.json` manifest creation and validation.
- `internal/registry` generates machine-readable federation registry files.

## Reports

- `internal/reports` generates human-readable reports such as dashboards and dependency graphs.

## Events

- `internal/events` owns append-only runtime event logging.
- Services should emit events for important lifecycle changes.

## Providers

- `internal/providers` owns provider registration and provider registry output.
- Future provider implementations should plug in here.

## Services and Runtime

- `internal/services` describes service state.
- `internal/jobs` owns repeatable runtime jobs.
- `internal/supervisor` coordinates jobs and services.
- `internal/runtime` owns one-shot and loop runtime behavior.
- `internal/daemon` starts long-running runtime services.
- `internal/api` exposes local HTTP endpoints.

## Rules

1. Generated files live in `registry/`, `reports/`, `logs/`, and `var/`.
2. Compiled binaries stay local and must not be tracked.
3. Shell scripts only bootstrap, install, test, inspect, or preserve AI training history.
4. New features should be Go packages first, CLI commands second.
5. Development history scripts belong in `AI-Code-Training/`.
