# MoBox Android inference recovery: merge acceptance contract

Companion infrastructure candidate: [VPS PR #22](https://github.com/AIFreedomTrustFederation/VPS/pull/22).

## Required before merging PR #102

- [ ] Exact-head `make verify`, Go unit tests and race tests pass; correct gofmt failures.
- [ ] A browser-originated message completes with `mode=local-inference` on Samsung Internet, not just a Python-created session.
- [ ] A canceled browser request does not cancel a bounded server-managed generation.
- [ ] Job submission is idempotent under retries, double taps and concurrent requests; different prompts must never silently share one active job.
- [ ] Job IDs and results remain discoverable after refresh. Explicitly define restart semantics; current in-memory job map is **not durable** and must not be described as such.
- [ ] Server restart during a job produces an accurate recoverable or interrupted status, never an indefinite spinner or fabricated response.
- [ ] Frontend resumes polling or retrieves persisted session state after navigation/reload; never submits a duplicate prompt to recover a lost job ID.
- [ ] Local model unavailable, busy, timed out, canceled, and context-overflow errors are distinguishable, without exposing sensitive prompts or filesystem paths in public UI.
- [ ] Bounded queue, concurrency, request sizes, generated token count, memory and model deadlines are covered by tests.
- [ ] Server shutdown cancels outstanding work cleanly and prevents accepting new jobs during drain.
- [ ] Repeated status polling does not leak full conversation history; same-origin and loopback-only boundaries remain enforced.
- [ ] No automatic remote model fallback. Any off-device inference requires explicit user consent and a separately verified authenticated, encrypted transport.
- [ ] Review the system prompt and 1024-token context behavior with actual model capacity.
- [ ] Mobile UX provides waiting state, elapsed time, retry guidance and a privacy-safe diagnostic copy action.
- [ ] Real Android app-switch/background/reopen test, plus local restart and no-network tests pass.

## VPS separation

VPS PR #22 is a **draft opt-in Linux host profile**, not a running server or an
Android Termux supervisor. Actual persistent hosting needs a real Linux machine
and operator-approved service installation. Android startup requires a separate
Termux-compatible design with explicit Android process-lifecycle limitations.
Never treat passing CI as proof of a deployed model or failover.

## Current verified baseline

A Python HTTP client completed a MoBox session via port 8788 in 14.31 seconds,
with `AIFT_READY` and `mode=local-inference`. Samsung Internet message submissions
returned `context canceled`. Reading a Python-created saved conversation in the
browser worked. This establishes a backend baseline, not browser acceptance.
