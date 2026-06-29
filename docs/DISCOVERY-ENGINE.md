# AIFT Discovery Engine

The Discovery Engine discovers repository reality from evidence.

It does not hard-code repository names, services, runtimes, capabilities, package managers, documentation systems, workflows, or schemas.

## Commands

- `aiftd discovery scan`
- `aiftd discovery list`
- `aiftd discovery object <id-or-name>`
- `aiftd discovery report`

## Generated runtime artifacts

- `registry/discovery.json`
- `reports/discovery.md`

These are ignored runtime state and should be regenerated from truth.

## Discovery evidence

The engine currently detects:

- Git repositories
- Documentation
- Schemas
- Workflows
- Package manifests
- Go modules
- Node packages
- Rust crates
- Python projects
- Docker projects
- AIFT contracts
- Commands
- Capabilities
- Services
- Health checks
