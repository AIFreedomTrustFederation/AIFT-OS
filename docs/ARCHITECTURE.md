# AIFT-OS Architecture

Generated: 2026-06-28T17:34:16Z

## Package Categories

### Foundation

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `config` | 0 | 35 | yes |
| `fsutil` | 0 | 8 | yes |
| `gitx` | 0 | 6 | no |
| `jsonfile` | 0 | 15 | yes |
| `sliceutil` | 0 | 7 | yes |
| `version` | 0 | 2 | no |
| `workspace` | 1 | 17 | no |

### Runtime

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `api` | 10 | 2 | yes |
| `daemon` | 4 | 1 | no |
| `jobs` | 5 | 3 | no |
| `runtime` | 4 | 2 | no |
| `scheduler` | 4 | 0 | no |
| `supervisor` | 5 | 1 | no |

### Events

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `eventbus` | 2 | 2 | yes |
| `eventmesh` | 4 | 1 | yes |
| `events` | 1 | 18 | yes |

### Analysis

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `capabilities` | 5 | 2 | yes |
| `discoveryengine` | 6 | 2 | yes |
| `graph` | 6 | 1 | yes |
| `intelligence` | 8 | 1 | yes |

### Kernel

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `kernel` | 1 | 0 | no |
| `kernelregistry` | 6 | 2 | yes |
| `kernelruntime` | 5 | 1 | no |
| `modules` | 6 | 1 | yes |
| `patchengine` | 2 | 1 | yes |

### Data

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `manifests` | 2 | 6 | yes |
| `registry` | 5 | 5 | no |
| `repo` | 4 | 2 | yes |
| `reports` | 4 | 5 | no |
| `state` | 1 | 3 | no |

### Federation

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `federation` | 8 | 1 | no |
| `sync` | 3 | 2 | no |
| `workflow` | 2 | 2 | no |

### Operations

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `doctor` | 1 | 1 | no |
| `manual` | 5 | 1 | no |

### Planning

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `planner` | 5 | 1 | yes |
| `servicecontracts` | 5 | 1 | yes |
| `services` | 3 | 3 | no |

### Extensions

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `plugins` | 2 | 1 | no |
| `providers` | 2 | 4 | no |

## Command Registry

| Command | Status | Handler | Help |
|---|---|---|---|
| `capabilities` | active | yes | yes |
| `daemon` | active | yes | yes |
| `dashboard` | active | yes | yes |
| `deps` | active | yes | yes |
| `discovery` | active | yes | yes |
| `doctor` | active | yes | yes |
| `event-bus` | active | yes | yes |
| `events` | active | yes | yes |
| `federation` | active | yes | yes |
| `graph` | active | yes | yes |
| `intelligence` | planned | yes | yes |
| `kernel` | active | yes | yes |
| `kernel-registry` | active | yes | yes |
| `manifest` | active | yes | yes |
| `manual` | planned | yes | yes |
| `mesh` | planned | yes | yes |
| `modules` | active | yes | yes |
| `patch-engine` | active | yes | yes |
| `plan` | active | yes | yes |
| `plugins` | active | yes | yes |
| `providers` | active | yes | yes |
| `registry` | active | yes | yes |
| `repo` | active | yes | yes |
| `serve` | active | yes | yes |
| `service-contracts` | planned | yes | yes |
| `services` | active | yes | yes |
| `start` | active | yes | yes |
| `status` | active | yes | yes |
| `sync` | active | yes | yes |
| `tick` | active | yes | yes |
| `verify` | active | yes | yes |
| `version` | active | yes | yes |
| `workflow` | active | yes | yes |

## Package Dependency Graph

```mermaid
graph TD
    api --> config
    api --> jobs
    api --> manifests
    api --> providers
    api --> registry
    api --> reports
    api --> services
    api --> state
    api --> sync
    api --> version
    capabilities --> config
    capabilities --> events
    capabilities --> fsutil
    capabilities --> jsonfile
    capabilities --> workspace
    daemon --> api
    daemon --> config
    daemon --> events
    daemon --> runtime
    discoveryengine --> config
    discoveryengine --> events
    discoveryengine --> fsutil
    discoveryengine --> jsonfile
    discoveryengine --> sliceutil
    discoveryengine --> workspace
    doctor --> config
    eventbus --> config
    eventbus --> sliceutil
    eventmesh --> config
    eventmesh --> events
    eventmesh --> jsonfile
    eventmesh --> workspace
    events --> config
    federation --> config
    federation --> events
    federation --> manifests
    federation --> providers
    federation --> registry
    federation --> repo
    federation --> reports
    federation --> workflow
    graph --> config
    graph --> events
    graph --> fsutil
    graph --> jsonfile
    graph --> sliceutil
    graph --> workspace
    intelligence --> capabilities
    intelligence --> config
    intelligence --> events
    intelligence --> fsutil
    intelligence --> gitx
    intelligence --> jsonfile
    intelligence --> sliceutil
    intelligence --> workspace
    jobs --> config
    jobs --> events
    jobs --> providers
    jobs --> registry
    jobs --> reports
    kernel --> config
    kernelregistry --> config
    kernelregistry --> events
    kernelregistry --> fsutil
    kernelregistry --> jsonfile
    kernelregistry --> sliceutil
    kernelregistry --> workspace
    kernelruntime --> config
    kernelruntime --> discoveryengine
    kernelruntime --> eventbus
    kernelruntime --> jsonfile
    kernelruntime --> kernelregistry
    manifests --> config
    manifests --> workspace
    manual --> config
    manual --> events
    manual --> fsutil
    manual --> jsonfile
    manual --> workspace
    modules --> config
    modules --> events
    modules --> fsutil
    modules --> jsonfile
    modules --> sliceutil
    modules --> workspace
    patchengine --> config
    patchengine --> jsonfile
    planner --> config
    planner --> events
    planner --> jsonfile
    planner --> sliceutil
    planner --> workspace
    plugins --> config
    plugins --> workspace
    providers --> config
    providers --> jsonfile
    registry --> config
    registry --> gitx
    registry --> jsonfile
    registry --> manifests
    registry --> workspace
    repo --> config
    repo --> gitx
    repo --> manifests
    repo --> workspace
    reports --> config
    reports --> gitx
    reports --> manifests
    reports --> workspace
    runtime --> config
    runtime --> events
    runtime --> jobs
    runtime --> supervisor
    scheduler --> config
    scheduler --> events
    scheduler --> registry
    scheduler --> reports
    servicecontracts --> config
    servicecontracts --> events
    servicecontracts --> fsutil
    servicecontracts --> jsonfile
    servicecontracts --> workspace
    services --> config
    services --> events
    services --> state
    state --> config
    supervisor --> config
    supervisor --> events
    supervisor --> jobs
    supervisor --> services
    supervisor --> state
    sync --> config
    sync --> gitx
    sync --> workspace
    workflow --> config
    workflow --> jsonfile
    workspace --> config
```

## Architectural Invariants

| Invariant | Status |
|---|---|
| no-circular-imports | PASS |
| commands-have-handlers | PASS |
| commands-have-help | PASS |
| no-duplicate-commands | FAIL |
| no-orphaned-packages | FAIL |
| modules-have-source | PASS |
| capabilities-have-evidence | PASS |
| service-contracts-have-owner | FAIL |

