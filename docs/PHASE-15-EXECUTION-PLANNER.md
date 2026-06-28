# Phase 15: Federation Execution Planner

AIFT-OS now has a truthful execution planner.

## Principle

The planner decides readiness.

The runtime executes only after the planner says something is ready.

## States

- `READY` — at least one ready/v1 capability and no blockers
- `DETECTED` — detected but not proven executable
- `PLANNED` — no proven capabilities
- `BLOCKED` — service requirements are unmet or planned
- `BROKEN` — broken capabilities exist

## Commands

- `aift plan build`
- `aift plan summary`
- `aift plan repo <repo>`
- `aift plan ready`
- `aift plan blocked`
- `aift plan report`

## Generated

- `registry/execution-plan.json`
- `reports/execution-plan.md`
- `reports/execution-blockers.md`

## Truth Rule

The planner must explain why a repository is ready, planned, detected, blocked, or broken.
