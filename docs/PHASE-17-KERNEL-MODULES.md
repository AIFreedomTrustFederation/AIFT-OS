# Phase 17: Federation Kernel Module Manifests

AIFT-OS treats every discovered repository as a kernel module candidate.

Modules are discovered from files that actually exist on disk. Planned features remain planned until evidence, commands, manifests, services, or health checks prove readiness.

## Commands

- `aiftd modules init-all`
- `aiftd modules scan`
- `aiftd modules list`
- `aiftd modules repo <repo>`
- `aiftd modules report`

## Generated files

- `.aift/module.json`
- `registry/modules.json`
- `reports/modules.md`
