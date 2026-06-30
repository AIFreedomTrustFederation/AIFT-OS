# Phase 6: Runtime Execution Engine

AIFT-OS now moves from discovery into execution.

This phase adds a truth-based runtime execution planner.

The planner discovers repositories and executable packages from disk without hardcoded repository names.

It detects common execution managers such as pnpm, npm, Go, Cargo, Python, and Make.

It writes the discovered execution plan to:

registry/execution-plan.json

This phase does not automatically launch long-running services. Planned services remain planned until explicitly started and verified.
