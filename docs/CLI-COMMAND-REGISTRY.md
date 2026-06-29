# AIFT CLI Command Registry

The AIFT CLI is being refactored into small Unix-style command handlers.

`cmd/aift/main.go` is the dispatcher.

Subsystems own their command files:

- `capabilities.go`
- `discovery_command.go`
- `kernel_registry_command.go`
- `kernel_runtime_command.go`
- `event_bus_command.go`
- `patch_engine_command.go`
- `modules_command.go`
- `legacy_command_stubs.go`

## Rule

A subsystem import belongs in its subsystem command file, not in `main.go`, unless `main.go` directly uses that subsystem.

## Planned handlers

Some older command routes are preserved as truthful planned stubs so the CLI compiles without pretending functionality exists.

A planned stub must return a clear error and must not simulate functionality.
