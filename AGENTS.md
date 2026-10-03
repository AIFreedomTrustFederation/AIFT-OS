# Repository Instructions

These instructions apply to the entire repository.

## Mission and boundaries

- Preserve AIFT-OS as the truthful federation control plane described in
  `README.md`: discover and coordinate sovereign repositories without absorbing
  their authority.
- Inspect actual repository metadata, runtime state, and command results before
  reporting readiness or capability.
- Require human approval before destructive synchronization, deployment,
  credential, custody, financial, legal, or safety-affecting operations.
- Do not convert planned, missing, blocked, or failed state into a successful
  status.

## Required verification

Run the canonical local gate from the repository root:

```sh
corepack pnpm install --frozen-lockfile --ignore-scripts
corepack pnpm run typecheck
make verify
git diff --check
test -z "$(git status --porcelain)"
```

Local verification intentionally isolates generated coverage and architecture
outputs. Do not set `AIFT_KEEP_ARTIFACTS=1`; that CI-only mode retains the
declared reports for artifact upload.

While preparing a commit, use `git status --short` instead of the final
clean-tree assertion and confirm that only the intended files are changed.
