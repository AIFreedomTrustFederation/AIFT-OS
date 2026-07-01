# AIFT-OS Architecture

Generated: 2026-07-01T14:05:14Z

## Package Categories

### Foundation

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `config` | 0 | 44 | yes |
| `fsutil` | 0 | 9 | yes |
| `gitx` | 0 | 5 | yes |
| `jsonfile` | 0 | 17 | yes |
| `sliceutil` | 0 | 7 | yes |
| `version` | 0 | 1 | no |
| `workspace` | 1 | 17 | yes |

### Runtime

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `api` | 10 | 1 | yes |
| `daemon` | 4 | 0 | no |
| `jobs` | 5 | 3 | no |
| `runtime` | 4 | 1 | no |
| `scheduler` | 2 | 1 | no |
| `supervisor` | 5 | 1 | no |

### Events

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `eventbus` | 2 | 1 | yes |
| `eventmesh` | 4 | 0 | yes |
| `events` | 1 | 18 | yes |

### Analysis

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `capabilities` | 5 | 1 | yes |
| `discoveryengine` | 6 | 1 | yes |
| `graph` | 6 | 0 | yes |
| `intelligence` | 8 | 0 | yes |

### Kernel

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `kernel` | 1 | 0 | no |
| `kernelregistry` | 6 | 1 | yes |
| `kernelruntime` | 5 | 0 | no |
| `modules` | 6 | 0 | yes |
| `patchengine` | 2 | 0 | yes |

### Data

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `manifests` | 2 | 5 | yes |
| `registry` | 5 | 3 | yes |
| `repo` | 4 | 1 | yes |
| `reports` | 4 | 3 | yes |
| `state` | 1 | 3 | yes |

### Federation

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `federation` | 8 | 0 | no |
| `sync` | 3 | 1 | no |
| `workflow` | 2 | 1 | yes |

### Operations

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `doctor` | 1 | 0 | no |
| `manual` | 5 | 0 | no |

### Planning

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `planner` | 5 | 0 | yes |
| `servicecontracts` | 5 | 0 | yes |
| `services` | 3 | 2 | yes |

### Extensions

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `plugins` | 2 | 0 | no |
| `providers` | 2 | 3 | yes |

### Other

| Package | Dependencies | Dependents | Tests |
|---|---:|---:|---:|
| `ai` | 2 | 0 | no |
| `builder` | 4 | 0 | no |
| `capability` | 1 | 2 | no |
| `cli` | 1 | 0 | yes |
| `compiler` | 1 | 2 | no |
| `fedbuild` | 2 | 0 | no |
| `lifecycle` | 1 | 1 | no |
| `providerregistry` | 1 | 0 | yes |
| `readiness` | 5 | 0 | yes |
| `schedulerplan` | 3 | 0 | yes |

## Command Registry

| Command | Status | Handler | Help |
|---|---|---|---|
| `bootstrap` | active | yes | yes |
| `federation` | planned | yes | yes |
| `help` | active | yes | yes |
| `registry` | active | yes | yes |
| `repo` | planned | yes | yes |
| `status` | active | yes | yes |
| `verify` | active | yes | yes |
| `workflow` | planned | yes | yes |

## Package Dependency Graph

```mermaid
graph TD
    ai --> compiler
    ai --> config
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
    builder --> compiler
    builder --> config
    builder --> lifecycle
    builder --> scheduler
    capabilities --> config
    capabilities --> events
    capabilities --> fsutil
    capabilities --> jsonfile
    capabilities --> workspace
    capability --> config
    cli --> config
    compiler --> config
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
    fedbuild --> capability
    fedbuild --> config
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
    lifecycle --> config
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
    providerregistry --> config
    providers --> config
    providers --> jsonfile
    readiness --> config
    readiness --> events
    readiness --> fsutil
    readiness --> jsonfile
    readiness --> workspace
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
    scheduler --> capability
    scheduler --> config
    schedulerplan --> config
    schedulerplan --> events
    schedulerplan --> jsonfile
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
| no-duplicate-commands | PASS |
| no-orphaned-packages | PASS |
| modules-have-source | PASS |
| capabilities-have-evidence | PASS |
| service-contracts-have-owner | PASS |

