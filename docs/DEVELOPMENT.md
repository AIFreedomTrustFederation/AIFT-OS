# Development Workflow

AIFT-OS uses one repository-native verification router so Termux, Linux, and GitHub Actions apply the same rules.

## Daily commands

```bash
make format
make verify-fast
make verify
make build
make smoke
```

- `make verify-fast` checks formatting, shell syntax, all Go tests, isolated command builds, and the artifact policy.
- `make verify` adds the MoBox UXI smoke test, coverage threshold, and architecture invariants.
- Local verification removes temporary coverage/build output and runs architecture generation in an isolated copy.
- CI sets `AIFT_KEEP_ARTIFACTS=1` so coverage and architecture reports remain available for upload.
- `make build` intentionally writes ignored executables to `bin/`.

## Optional pre-push gate

```bash
make install-hooks
```

The installer refuses to overwrite a different existing hook. The hook runs `make verify-fast`; use `make verify` before merging a substantial vertical slice.

## Artifact policy

Compiled executables and transient coverage output are never source:

- `/aift`
- `/aiftd`
- `/bin/`
- `/coverage.out`

Release automation may package binaries, but Git history keeps source and intentional reports only.

## Vertical slices

A playable development branch should connect:

```text
contract -> API -> renderer -> interaction -> governance -> evidence -> tests
```

Large speculative rewrites across unrelated repositories are avoided. Federation-wide behavior begins with a versioned contract, followed by role-specific adapters or manifests.
