# Phase 12: Federation Knowledge Graph Engine

AIFT-OS now builds a truthful graph of the federation from real repository evidence.

## Sources

The graph discovers:

- repositories
- manifests
- capability contracts
- manual contracts
- manual pages
- package scripts
- Go modules
- npm dependencies
- BookSmith manual build declarations

## Commands

- `aift graph`
- `aift graph summary`
- `aift graph repo <name>`
- `aift graph type <type>`
- `aift graph status <status>`

## Generated Registry Files

- `registry/graph.json`
- `registry/graph.mermaid`
- `registry/graph.dot`
- `registry/graph.graphml`
- `registry/graph.cypher`
- `registry/graph.rdf`

## Generated Reports

- `reports/graph.md`
- `reports/graph-summary.md`
- `reports/dependency-tree.md`
- `reports/orphaned-capabilities.md`
- `reports/planned-vs-running.md`
- `reports/service-map.md`

## Principle

Nothing is simulated.

Nodes and edges are generated only from files and contracts that actually exist.
