# AIFT Capability-Aware Build

AIFT must never fail a federation build simply because the current machine lacks a runtime.

Instead, it discovers local capabilities first, then classifies modules honestly:

- active: runnable here
- planned: valid but waiting for runtime/tooling
- blocked: invalid or unsafe
- unsupported: no provider can handle it

This keeps AIFT module-agnostic, provider-agnostic, runtime-agnostic, and sync/async capable.
