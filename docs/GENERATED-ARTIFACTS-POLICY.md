# Generated Artifacts Policy

AIFT-OS treats generated discovery outputs as runtime state, not source truth.

## Committed

Commit source code, schemas, tests, manuals, and scripts that generate or validate truth.

## Ignored

Do not normally commit generated registry snapshots or generated reports:

- `registry/*.json`
- `reports/*.md`

These files are regenerated from discovered repository state.

## Rule

The registry is a cache of discovered truth.

The source of truth is the filesystem, Git state, manifests, contracts, tests, health checks, and validation evidence.

## Validation

Scripts must verify generated artifacts exist after generation, but should not stage ignored runtime artifacts unless the operator intentionally force-adds them.
