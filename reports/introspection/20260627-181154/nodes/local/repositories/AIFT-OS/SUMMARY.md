# AIFT-OS

## Git Status
```text
?? AI-Code-Training/scripts/phase-scripts/phase-16-introspection-kernel.sh
?? ci-truth-repair-report.sh
?? fix-introspection-and-push-federation.sh
?? phase-16-diagnostics.sh
?? phase-16-introspection-kernel.sh
?? reports/diagnostics-20260627-174933.tar.gz
?? reports/diagnostics/
?? reports/introspection/
```

## Detected Commands
```text
go:test=go test ./...
go:build=go build ./...
```

## Logs

### go-build.log
```text
LABEL=go-build
CMD=go build ./...
DATE=Sat Jun 27 18:12:07 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found
```

### go-test.log
```text
LABEL=go-test
CMD=go test ./...
DATE=Sat Jun 27 18:12:05 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found
?   	github.com/AIFreedomTrustFederation/AIFT-OS/cmd/aift	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/api	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/capabilities	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/config	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/daemon	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/doctor	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/eventmesh	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/events	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/federation	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/gitx	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/graph	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/intelligence	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/jobs	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernel	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/manual	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/planner	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/plugins	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/providers	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/repo	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/runtime	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/scheduler	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/servicecontracts	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/services	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/state	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/supervisor	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/sync	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/version	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/workflow	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace	[no test files]
```
