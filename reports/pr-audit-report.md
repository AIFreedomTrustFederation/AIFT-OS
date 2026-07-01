# AIFT-OS Pull Request Audit Report

Generated: 2026-07-01T06:29:08-07:00

Branch audited from: fix/windows-verification-bootstrap

Commit before report: 2b5cd51

Repository: https://github.com/AIFreedomTrustFederation/AIFT-OS

## Verification Commands

- git status: clean before report generation
- git branch --show-current: fix/windows-verification-bootstrap
- git remote -v: origin https://github.com/AIFreedomTrustFederation/AIFT-OS.git
- go test ./...: PASS on fix/windows-verification-bootstrap
- pnpm install: PASS
- pnpm run typecheck: PASS
- pnpm run lint: FAIL, root package has no lint script
- pnpm test: FAIL, root package/workspace has no test script
- pnpm run build: FAIL, artifacts/mockup-sandbox cannot load @rollup/rollup-win32-x64-msvc from local node_modules

## Open PR Summary

| PR | Title | Branch | Mergeable | Checks After Audit | Recommendation |
| --- | --- | --- | --- | --- | --- |
| #16 | phase2: establish central federation runtime layout | phase2-central-federation-runtime | yes | CI test FAIL, CodeRabbit PASS | Hold until coverage is restored |
| #17 | phase3: add recursive federation discovery engine | phase3-recursive-discovery-engine | yes | CI test FAIL, CodeRabbit PASS | Hold until coverage is restored |
| #18 | phase4: build federation runtime graph | phase4-federation-runtime-graph | yes | CI test FAIL, CodeRabbit PENDING | Hold until coverage is restored and review completes |
| #19 | phase5: upgrade existing runtime supervisor | phase5-upgrade-existing-runtime | yes | CI test FAIL, CodeRabbit PASS | Hold until coverage is restored |
| #20 | phase7: repair doctor output compatibility | phase7-doctor-git-housekeeping | conflicting | CI test FAIL, CodeRabbit PASS | Close/rebuild or manually resolve conflicts |
| #21 | test: repair tests from repository reality | phase7-repair-tests-from-reality | conflicting | CI test FAIL, CodeRabbit PASS | Close/rebuild or manually resolve conflicts |
| #22 | phase9: repair real repo interfaces | phase9-repair-real-repo-interfaces | conflicting | CI test FAIL, CodeRabbit PASS | Close/rebuild or manually resolve conflicts |

## Fixes Applied

- PR #16: committed and pushed 421c599, Fix phase script harness portability.
- PR #17: committed and pushed d108d1e, Fix phase script harness portability.
- PR #18: committed and pushed 70d560d, Fix phase script harness portability.
- PR #19: committed and pushed 7eb0729, Fix phase script harness portability.

The fix changed Termux-only script shebangs to `#!/usr/bin/env bash` and added explicit `# no-harness` annotations for standalone phase bootstrap utilities. Local `go test ./...` passed on each updated branch after merging current main into a temporary audit worktree.

## Failure Categorization

### PR #16

- Cause category: H, incorrect test assumption; C, CI quality gate configuration; B, cross-platform script portability.
- Before fix: `TestMutatingScriptsSourceHarness` failed because phase2 standalone scripts neither sourced `scripts/lib/aift-run.sh` nor declared an exemption.
- After fix: local `go test ./...` passed.
- GitHub CI after fix: FAIL in `scripts/check-coverage.sh`; current coverage 29.6% versus baseline 33.3%.
- Files changed by fix: `scripts/phase2-central-runtime-scan.sh`, `scripts/phase2-clean-generated-runtime.sh`, `scripts/phase2-status.sh`.
- Mergeable: no, because required CI is still red.

### PR #17

- Cause category: H, incorrect test assumption; C, CI quality gate configuration; F, generated registry drift risk.
- Before fix: phase2 and phase3 standalone scripts failed the harness test.
- After fix: local `go test ./...` passed.
- GitHub CI after fix: FAIL in coverage gate; current coverage 29.6% versus baseline 33.3%.
- Files changed by fix: phase2 scripts plus `scripts/phase3-recursive-discovery.sh`.
- Mergeable: no, because required CI is still red.

### PR #18

- Cause category: H, incorrect test assumption; C, CI quality gate configuration; F, generated registry drift risk.
- Before fix: phase2, phase3, and phase4 standalone scripts failed the harness test.
- After fix: local `go test ./...` passed.
- GitHub CI after fix: FAIL in coverage gate; current coverage 29.6% versus baseline 33.3%.
- Files changed by fix: phase2 scripts plus `scripts/phase3-recursive-discovery.sh` and `scripts/phase4-federation-runtime-graph.sh`.
- Mergeable: no, because required CI is still red and CodeRabbit was still pending at last check.

### PR #19

- Cause category: H, incorrect test assumption; C, CI quality gate configuration; A, insufficient test coverage for new runtime code.
- Before fix: phase2, phase3, phase4, and phase5 standalone scripts failed the harness test.
- After fix: local `go test ./...` passed.
- GitHub CI after fix: FAIL in coverage gate; current coverage 28.7% versus baseline 33.3%.
- Files changed by fix: phase2, phase3, phase4 scripts plus `scripts/phase5-upgrade-existing-runtime-check.sh`.
- Mergeable: no, because required CI is still red.

### PR #20

- Cause category: A, real code/test defect; G, merge conflict.
- Conflicts with main: `.aift/capabilities.json`, `cmd/aift/main.go`, `var/events/events.jsonl`.
- CI failures from log: syntax errors in several tests, unused variables, report path failures, and doctor tests that assume a `cmd/aift` path shape that no longer matches the active CLI layout.
- Mergeable: no.

### PR #21

- Cause category: H, incorrect assumptions in tests; G, merge conflict.
- Conflicts with main: `.aift/capabilities.json`, deleted/modified tests in `internal/capabilities` and `internal/eventbus`, `var/events/events.jsonl`.
- CI failure from log: `tests/reality` expected a runtime architecture document that is not present in the branch state.
- Mergeable: no.

### PR #22

- Cause category: A, real code defect; G, merge conflict.
- Conflicts with main: `.aift/capabilities.json`, `cmd/aift/main.go`, `internal/ai/ai.go`, `var/events/events.jsonl`.
- CI failures from log: duplicate `Verify` declarations in `internal/repair`, unresolved repair symbols such as `Context`, `Issue`, `Blocked`, and integration panic caused by build failure.
- Mergeable: no.

## Review Comments

- CodeRabbit was inspected through PR metadata. Actionable comments were visible for PR #16 around portable shebangs, generated artifact cleanup, and ignore ordering.
- PR #16 shebang portability was fixed. The generated cleanup behavior and ignore ordering still need focused review before merge because changing cleanup semantics can delete untracked local artifacts if implemented carelessly.
- Other visible CodeRabbit comments were mostly rate-limit or summary comments in the inspected metadata; unresolved thread-level state should be checked again before final merge.

## Checks Before

- All seven PRs had failing GitHub `test` checks before this audit.
- PR #20, #21, and #22 were already marked conflicting.
- PR #16, #17, #18, and #19 were marked mergeable but failed tests.

## Checks After

- PR #16-#19: local `go test ./...` passes after the pushed harness portability fix.
- PR #16-#19: GitHub CI still fails because coverage is below `coverage-baseline.txt`.
- PR #20-#22: not auto-fixed because conflicts and source-level compile/test defects require manual branch reconstruction.

## Merge Recommendation

- PR #16: hold. Add targeted tests or reduce untested surface until coverage meets the existing 33.3% baseline.
- PR #17: hold. Add tests for discovery output and registry generation before merge.
- PR #18: hold. Add tests for runtime graph generation and generated registry provenance before merge.
- PR #19: hold. Add tests for runtime, services, state, and supervisor code before merge.
- PR #20: close/rebuild unless the branch owner wants a manual conflict resolution pass.
- PR #21: close/rebuild unless the branch owner wants a manual conflict resolution pass.
- PR #22: close/rebuild unless the repair package API is redesigned against current main.

No functionality was marked working unless proven by tests or command output.
