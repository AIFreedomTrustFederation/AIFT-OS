# AIFT Apps Layer Report

## Summary

Implemented the real AIFT apps layer discovery surface for the `aift` CLI.

This phase adds truthful app manifest discovery from:

- `registry/apps/*.json`
- `.aift/apps/*.json`
- sibling federation repositories at `*/.aift/apps/*.json`

It also adds the command group:

- `aift apps list`
- `aift apps inspect <id>`
- `aift apps launch <id> --plan`

`launch` does not start processes. It returns a structured planned JSON response with `active: false` until local verification is implemented.

## Operator Runtime Context

The intended operator environment is:

- Primary: OpenHands
- Model runtime: Ollama
- Fallback CLI editor-agent: Aider
- Editors: VS Code, VSCodium, Neovim
- Host: local Linux machine
- Explicitly not assumed: phone Termux

This context is recorded in `registry/apps/booksmith-studio.json` metadata and should be used by later local verification work.

## Files Changed

- `cmd/aift/main.go`
  - Added `AIFTApp` manifest model.
  - Added app discovery helpers.
  - Added duplicate app ID detection.
  - Added malformed JSON handling with explicit errors.
  - Added `apps` command group.
  - Added planned-only launch response.
  - Added `registry/apps` to bootstrap and verification discovery.

- `cmd/aift/main_test.go`
  - Added tests for registry/local/sibling app discovery.
  - Added tests for duplicate app ID handling.
  - Added tests proving launch is planned and not active.
  - Added tests proving malformed JSON fails honestly.

- `registry/apps/booksmith-studio.json`
  - Added BookSmith Studio as a planned AIFT app manifest.
  - Preserves the launch command as metadata only; AIFT-OS does not execute it yet.
  - Records the local Linux/OpenHands/Ollama/Aider/editor context for future verification.

## Truthfulness Guarantees

- Duplicate app IDs are not silently collapsed.
- Duplicate manifests are marked with `duplicate: true` and listed in `duplicate_ids`.
- `inspect` reports duplicate matches instead of pretending one app is authoritative.
- `launch` requires `--plan` and returns `status: planned`, `active: false`.
- Malformed app JSON fails with `malformed app manifest ...` instead of being ignored.

## Verification

Performed in a reconstructed local Go module matching the updated files because the GitHub repository could not be cloned from the execution container.

Commands run:

```bash
go test ./...
go build ./cmd/aift
```

Result:

```text
ok   github.com/AIFreedomTrustFederation/AIFT-OS/cmd/aift  0.019s
```

Both requested commands passed in the reconstructed local module.

## Remaining Planned Work

- Implement local app verification before enabling active launch.
- Verify app working directories against discovered federation repositories on the local Linux host.
- Detect and report OpenHands, Ollama, Aider, VS Code/VSCodium, and Neovim capabilities honestly.
- Verify declared launch, health, and build commands before promoting an app from planned to active.
- Add richer manifest schema validation once the registry schema is finalized.
