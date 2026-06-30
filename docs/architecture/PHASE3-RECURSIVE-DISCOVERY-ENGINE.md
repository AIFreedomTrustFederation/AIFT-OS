# AIFT-OS Phase 3: Recursive Discovery Engine

AIFT-OS discovers top-level Git repositories first.

Each repository is then inspected internally for apps, packages, services, workflows, schemas, docs, commands, manifests, and capabilities.

## Design Rule

AIFT-OS never hardcodes repository names.

## Discovery Layers

1. Federation workspace
2. Git repositories
3. Repository manifests
4. Package managers
5. Apps and packages
6. Commands and scripts
7. Workflows
8. Docs and schemas
9. Capabilities
10. Federation graph

## Source of Truth

The source of truth is discovered reality on disk.

Generated runtime data belongs in AIFT-OS registry and runtime directories.
