# Phase 14: Federation Service Contracts

AIFT-OS now gives every repository a service contract.

## Principle

Event contracts describe how repositories communicate.

Service contracts describe what repositories provide, require, and may eventually run.

A planned service is not executed.

## Commands

- `aift service-contracts init-all`
- `aift service-contracts scan`
- `aift service-contracts list`
- `aift service-contracts repo <repo>`
- `aift service-contracts report`

## Per Repo

- `.aift/services.json`
- `.aift/services/`

## Generated

- `registry/service-contracts.json`
- `reports/service-contracts.md`

## Truth Rule

AIFT-OS records service promises, but does not treat them as executable until matching capabilities and health checks are ready or v1.
