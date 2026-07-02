# AIFT Runtime Wiring Report

## Purpose

This report documents the first manifest-first runtime wiring pass for AIFT-OS.

Rule:

```text
Registry declares.
Verifier proves.
Launcher starts only what is proven.
```

## Canonical Runtime Context

Declared in `manifests/aift-runtime.json`:

- Host: local Linux machine
- Primary agent: OpenHands
- Model runtime: Ollama
- Fallback CLI editor-agent: Aider
- Editors: VS Code, VSCodium, Neovim
- Explicitly not assumed as runtime host: phone Termux

Termux can still be used as a remote-control shell, but this repo must not assume that Termux is the host where OpenHands, Ollama, Aider, or desktop editors are installed.

## Registry Entries Added

- `registry/providers/ollama.json`
- `registry/agents/openhands.json`
- `registry/agents/aider.json`
- `registry/editors/vscode.json`
- `registry/editors/vscodium.json`
- `registry/editors/neovim.json`

## Current Status

All new runtime entries are `planned`.

Nothing in this pass marks any provider, agent, editor, or app as active.

## Duplicate and Wiring Policy

The next verifier layer must reject or report:

- duplicate IDs across provider, agent, editor, app, and service registries
- missing declared dependencies
- app commands that reference tools not proven locally
- editor commands that are declared but missing locally
- providers that are declared but not reachable locally
- more than one active owner for the same app
- more than one service claiming the same port once service manifests exist

## Verification Commands to Implement Next

- `aift runtime verify`
- `aift providers verify`
- `aift agents verify`
- `aift editors verify`
- `aift apps verify`

These commands must inspect the local Linux host and report discovered, missing, planned, or failed truthfully.

## Local Verification Targets

Expected checks:

```bash
go version
node --version
npm --version
pnpm --version
ollama --version
ollama list
openhands --version
aider --version
code --version
codium --version
nvim --version
```

No command above should be treated as active unless it succeeds locally.

## PR Status

This report belongs to the draft PR for the AIFT apps layer and runtime manifest wiring.

The implementation is intentionally conservative: declarations exist, but activation is blocked until local verification is implemented and passes.
