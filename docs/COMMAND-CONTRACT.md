# Engineering command contract

Run commands from the repository root. Source QA remains separate from live
capability, security, dependency provenance and deployment acceptance.

| Command | Prerequisites | Effects and success meaning |
| --- | --- | --- |
| `pnpm install --frozen-lockfile --ignore-scripts` | Node and manifest-pinned pnpm; package registry access or populated cache | Installs locked dependencies; lifecycle scripts disabled. Not an offline guarantee. |
| `pnpm run typecheck` | Installed dependencies | Runs library and workspace TypeScript checks; may write ignored incremental build state. |
| `make verify` | Bash, Git, real Go compiler plus gofmt, Make, Python 3, tar and existing smoke-test prerequisites | Runs formatting, shell, Python and Go tests, isolated builds, smoke, coverage and architecture gates. Nonzero means stop. Local reports isolated; do not set CI-only AIFT_KEEP_ARTIFACTS. |
| `python3 -B scripts/upgrade_inventory.py .` | Python 3.10+ and Git; exact repository root | Prints offline inventory JSON. Exit 0 means inventory complete, not tests passed. Never executes metadata commands. |

The original preparation PR #75 passed hosted CI `test` at
`2faa401eef8733fd0b9faf352a8fca5ebcda8663` ([run](https://github.com/AIFreedomTrustFederation/AIFT-OS/actions/runs/44213254401)).
That job includes canonical `make verify`. Local execution remains blocked by
missing `gofmt`; `/usr/bin/go` here is not the Go compiler. Do not replace Go
verification with a stub. Hosted `live-producer-evidence` was skipped, not passed.

Offline Runtime/Forge fixture tests exercise the existing consumer only. They
are synthetic v1 payloads, not authenticated producer runs, live cross-repository
acceptance, or authorization to expand collector permissions.

After verification run `git diff --check` and inspect `git status --short`.
After committing, require a clean worktree. Hosted evidence must match the final
PR head; older green checks do not certify subsequent commits.
