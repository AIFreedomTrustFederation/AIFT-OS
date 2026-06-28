# AIFT Event Bus

The Event Bus is the nervous system of AIFT-OS.

It records meaningful operating-system activity as append-only runtime evidence.

## Principles

- Events are runtime state, not source truth.
- Events should be append-only.
- Events should support replay.
- Events should be generated from observable actions.
- Events should not pretend a capability is active unless validation proves it.

## Commands

- `aiftd event-bus publish <topic> <message> [key=value...]`
- `aiftd event-bus list`
- `aiftd event-bus replay [topic]`
- `aiftd event-bus report`

## Runtime artifacts

- `var/events/event-bus.jsonl`
- `reports/event-bus.md`

These are intentionally ignored generated runtime artifacts.
