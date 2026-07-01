# AIFT-OS Functionality Gap Report

Generated: 2026-07-01T06:29:08-07:00

Branch: fix/windows-verification-bootstrap

Evidence commit: 2b5cd51

## 1. Current Working Functionality

- `go test ./...` passes on the current branch.
- `pnpm install` passes.
- `pnpm run typecheck` passes.
- Active CLI commands in `cmd/aift/main.go`: `help`, `status`, `verify`, `registry`, `bootstrap`.
- Planned CLI commands in `cmd/aift/main.go`: `federation`, `repo`, `workflow`.
- `status`, `verify`, `registry`, and `bootstrap` currently return JSON from local repository checks and command metadata.

Commands not proven active: `doctor`, `monitor`, `repair`, and `report`.

## 2. Broken or Incomplete Functionality

### Root lint and test scripts

- Current status: missing
- Evidence: `pnpm run lint` fails with missing script; `pnpm test` fails with missing test command.
- Files involved: `package.json`
- Root cause: root package defines `build`, `typecheck`, and `typecheck:libs`, but no `lint` or `test`.
- Required fix: add truthful lint/test scripts wired to real tools, or add CI documentation stating they are intentionally unavailable.
- Priority: P1

### Frontend build

- Current status: broken in current local install state
- Evidence: `pnpm run build` fails in `artifacts/mockup-sandbox` because Rollup cannot find `@rollup/rollup-win32-x64-msvc`.
- Files involved: `artifacts/mockup-sandbox`, root pnpm workspace, lockfile/install state
- Root cause: Windows native optional Rollup dependency is absent from `node_modules`; `pnpm install --force` did not restore it.
- Required fix: refresh the dependency install state cleanly, confirm the optional dependency is represented in the lockfile, and add CI that exercises Windows builds.
- Priority: P1

### PR coverage gate

- Current status: broken for open phase PRs
- Evidence: PR #16-#18 report 29.6% coverage and PR #19 reports 28.7%, below the 33.3% baseline.
- Files involved: `scripts/check-coverage.sh`, `coverage-baseline.txt`, new runtime/discovery scripts and packages in PR branches
- Root cause: new code was added without enough tests to preserve baseline coverage.
- Required fix: add targeted tests for the new behavior rather than lowering the baseline.
- Priority: P0 for merge readiness

### Phase repair PRs

- Current status: broken/conflicting
- Evidence: PR #20-#22 have merge conflicts and failing CI; PR #22 has duplicate `Verify` definitions and unresolved repair types.
- Files involved: `cmd/aift/main.go`, `internal/repair`, `internal/ai`, generated capability/event files
- Root cause: old branch assumptions no longer match the cleaned active CLI and current internal package APIs.
- Required fix: rebuild those branches on current main, quarantine stale generated code, and rewire package APIs against one source of truth.
- Priority: P0

## 3. Needed New CLI Functionality

- `aift status`: active but minimal. Needs structured repo, runtime, dependency, and capability checks with stable schema tests.
- `aift verify`: active but minimal. Needs a real verification engine with capability detection, platform-specific skips, and generated artifact drift checks.
- `aift bootstrap`: active but minimal. Needs real federation bootstrap discovery with provenance and deterministic output.
- `aift registry`: active for CLI command metadata. Needs unified registry generation for repos, workflows, services, providers, models, and artifacts.
- `aift federation`: planned. Needs real federation registry and health graph operations before it can be active.
- `aift repo`: planned. Needs repo discovery, dirty-state handling, manifest validation, and safe sync operations.
- `aift workflow`: planned. Needs a workflow runner with capability checks, timeouts, logs, and truthful skip states.
- `aift doctor`: missing from active CLI. Needs diagnostic checks for CLI path, config, registry drift, dependencies, shell/tool availability, and platform compatibility.
- `aift monitor`: missing. Needs morning monitor integration, snapshot persistence, and change detection.
- `aift repair`: missing from active CLI. Needs non-destructive repair planning and explicit confirmation boundaries.
- `aift report`: missing. Needs report generation backed by real registry and verification state.

## 4. Needed Internal Packages

- `internal/capabilities`: present but needs complete platform/tool detection and schema tests.
- `internal/eventbus`: present with tests, needs integration with real monitor/runtime events.
- `internal/federation`: present but not proven as a complete engine.
- `internal/gitx`: present, needs cross-platform command timeout and error-shape tests.
- `internal/verify`: missing as a dedicated package; current verify logic lives in CLI-level checks.
- `internal/workflow`: present defaults only; needs execution, skip, timeout, and result persistence.
- `internal/registry`: present but needs deterministic generation and drift validation.
- `internal/monitor`: missing; needed for project steward and morning briefing automation integration.
- `internal/reporting`: missing as named package; `internal/reports` exists but needs a stable report API.
- `internal/runtime`: present but not sufficiently covered.
- `internal/config`: present and tested, needs environment and repo-root precedence tests.

## 5. Needed Cross-Platform Support

- Shell assumptions: tests now skip shell smoke tests on Windows when bash/sh is unavailable, but phase scripts still assume bash when executed.
- Path assumptions: use `filepath.Join` in Go; continue auditing generated registries that contain Unix/Termux paths.
- Executable detection: centralize `exec.LookPath` checks for bash, sh, docker, podman, ollama, aiftd, git, node, and pnpm.
- Permissions: add tests for executable bit behavior on Windows and Unix.
- Process spawning: add timeout/cancellation coverage around git and runtime commands.
- Symlinks: add tests for environments without symlink permission on Windows.
- Line endings: keep shell scripts LF in the repo and add `.gitattributes` if needed.
- Android/Termux: replace hard-coded Termux shebangs with portable shebangs and detect Termux-specific capabilities at runtime.

## 6. Needed GitHub/CI Functionality

- Add Windows, Linux, and macOS matrix jobs.
- Add explicit Go test workflow steps separate from coverage threshold checks.
- Add pnpm install, typecheck, build, lint, and test steps only when scripts exist or after adding real scripts.
- Add generated artifact drift checks for CLI registry and bootstrap output.
- Add capability probes for bash, sh, docker, podman, ollama, and aiftd before integration tests use them.
- Add coverage reporting that identifies packages responsible for drops.
- Avoid lowering coverage baseline unless a reviewed change proves the baseline is obsolete.

## 7. Needed Federation Features

- Repo discovery with manifest validation and dirty-state detection.
- Federation registry with stable schemas and deterministic JSON.
- Project monitoring snapshots for active repositories.
- Capability detection with planned/active validation.
- Artifact provenance for generated registries and reports.
- Report generation from real internal state, not stale snapshots.
- Morning monitor integration that uses current workspace state and writes concise memory.

## 8. Needed Tests

- Unit tests for `aift status`, `aift verify`, `aift registry`, and `aift bootstrap` schemas.
- CLI binary tests that build `bin/aift` and execute the real binary.
- Windows tests for path separators, missing shell, and missing executable cases.
- Termux tests or simulated path tests for mobile shell assumptions.
- Registry tests for deterministic output and generated artifact drift.
- Workflow tests for unavailable capabilities producing honest skip/planned results.
- GitHub Actions tests for matrix behavior and missing optional tools.
- Runtime/supervisor tests for PR #19 coverage restoration.
- Discovery/graph tests for PR #17 and PR #18 coverage restoration.

## 9. Implementation Roadmap

- Phase 21: stabilize active CLI, keep one dispatcher, add CLI schema tests, and ensure generated CLI artifacts are deterministic.
- Phase 22: build a real verification engine with central capability detection and truthful skip/planned states.
- Phase 23: implement the federation registry engine with repo, service, provider, workflow, and artifact provenance records.
- Phase 24: implement repo monitor snapshots and morning briefing integration.
- Phase 25: implement the workflow runner with timeouts, logs, dry-run behavior, and capability checks.
- Phase 26: implement the repair engine as non-destructive plans first, with explicit execution boundaries.
- Phase 27: add release readiness checks, cross-platform matrix CI, generated drift gates, and documentation completeness checks.

## 10. Final Merge Recommendation

- PR #16: hold until coverage meets baseline with real tests.
- PR #17: hold until discovery tests restore coverage.
- PR #18: hold until graph/registry tests restore coverage and review completes.
- PR #19: hold until runtime/supervisor tests restore coverage.
- PR #20: close/rebuild or manually resolve conflicts on current main.
- PR #21: close/rebuild or manually resolve conflicts on current main.
- PR #22: close/rebuild; repair interfaces are currently broken against current main.

AIFT-OS currently has a stable minimal CLI and passing Go/typecheck verification on the audited branch, but it is not yet a complete truthful federation operating system. The main gaps are real verification, registry, monitoring, workflow, repair, and cross-platform CI coverage.
