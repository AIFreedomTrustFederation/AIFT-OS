# Phase 11: Federation Manual Contracts

Every repository now receives a UNIX-style manual source structure and a `.aift/manual.json` contract.

## Principle

Every repo owns its manual source.

BookSmith owns PDF generation, publication packets, proofing, and static web export.

AIFT-OS owns discovery, truth, contracts, reporting, and orchestration readiness.

## Commands

- `aift manual init-all`
- `aift manual scan`
- `aift manual report`
- `aift manual repo <repo>`

## Generated Per Repo

- `.aift/manual.json`
- `docs/manual/README.md`
- `docs/manual/source/index.md`
- `docs/manual/source/man0/`
- `docs/manual/source/man1/`
- `docs/manual/source/man2/`
- `docs/manual/source/man3/`
- `docs/manual/source/man4/`
- `docs/manual/source/man5/`
- `docs/manual/source/man6/`
- `docs/manual/source/man7/`
- `docs/manual/source/man8/`
- `docs/manual/source/man9/`
- `docs/manual/assets/`

## Generated In AIFT-OS

- `registry/manuals.json`
- `reports/manuals.md`

## Status Truth

- `manual.source` becomes `ready` when source folders exist.
- `manual.pdfBuild` remains `planned` until BookSmith exposes a verified `manual.build.pdf` capability.
- `manual.webPublish` remains `planned` until the website/static export integration is proven.
