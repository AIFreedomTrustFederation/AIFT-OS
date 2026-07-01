# AIFT Federation Scheduler

The Federation Scheduler classifies every discovered repository without pretending unavailable tooling is failure.

## States

- ACTIVE: verified and runnable here
- PLANNED: valid but waiting for optional runtime or dependency
- WAITING: prerequisite or capability not available
- BLOCKED: real source/configuration problem
- UNSUPPORTED: no provider can handle this repository yet

Missing local tools are not fatal.
Unsupported runtimes are not fatal.
Only real defects are blocked.
