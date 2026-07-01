# Phase 9: Real Repository Repair

This phase repairs the repair engine against the repository as it actually exists.

Rules:

- Preserve existing public callers.
- Restore compatibility functions instead of breaking callers.
- Do not run Git commands inside non-Git temp test directories.
- Treat Git repair as optional when no Git worktree exists.
- Verify with `go test ./...`.
