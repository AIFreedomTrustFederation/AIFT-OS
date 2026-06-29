# Phase 6: Core Services

AIFT-OS now includes the first real operating-system service layer.

Implemented:

- event log
- scheduler
- provider registry
- runtime one-shot start
- daemon entry point
- local HTTP API skeleton
- version command
- events command
- providers command

Commands:

- aift version
- aift start
- aift events
- aift providers
- aift serve :8787
- aift daemon :8787

The daemon currently runs the scheduler and local API. This is the foundation for future federation services.
