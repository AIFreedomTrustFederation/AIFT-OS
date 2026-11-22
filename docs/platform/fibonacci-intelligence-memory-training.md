# AIFT Recursive Fibonacci Intelligence — Memory, Context, Training and Federation v1

> **Proposed, versioned engineering contract — 2026-11-22.** This document records operator-directed architecture, not an implemented memory engine, retrained Aetherion model, verified long-context capability, working distributed cluster, or release approval. Source tests and measured artifacts outrank declarations. Adopt it through separate, reviewable implementation PRs and actual acceptance evidence.

**Contract identifier:** aift.memory.fibonacci.v1  
**Authority:** AIFT-OS defines cross-federation orchestration and the compatibility contract; each producing repository owns its own implementation and data.  
**Design language:** Fibonacci recurrence, golden ratio and Tree of Emergence are explicit scheduling, hierarchy and visual metaphors. They are **not** scientific proof that sacred geometry increases intelligence or changes hardware capacity.

## 1. Objective and first integration milestone

Make user-facing AIFT conversation and project history effectively unbounded by **persistent storage and retrieval**, even though any single model request has finite context, throughput, memory and generation limits. Maintain sovereign, offline-capable, attributable knowledge without inventing tokens, sources, operations or successful tests.

The immediate executable target is one verified local path:

~~~text
Living Universe / Federation Console UI
    -> AIFT-OS bounded session and app-capability contract
    -> AIFT-Runtime observed local service and health
    -> llama.cpp / Aetherion OpenAI-compatible loopback provider
    -> user-visible answer + source/evidence + exact error classification
~~~

On 2026-11-22, the operator observed a 3.397B-parameter Aetherion GGUF answering a short synthetic request through llama.cpp at 127.0.0.1:8080 with aetherion:latest. The llama.cpp log reported a trained context of 32,768 and 1,024 active tokens **per sequence slot** across four slots. A separate AIFT-OS Console on 127.0.0.1:8788 answered HTTP health but its browser conversation reported a degraded inference connection when port 8080 was refused. These are time-bound observations, **not** evidence that the complete UI-to-model path has passed or that either service remains online now.

The current AIFT-OS Go CLI app launch command is plan-only, and its Console health response does not prove model inference. Do not equate a manifest, HTTP 200, declared "ready" status or a successful direct model request with integrated launch or generation.

## 2. Mathematical progression and candidate tiers

Define F(0)=0, F(1)=1 and F(n+2)=F(n+1)+F(n). Let phi=(1+sqrt(5))/2, approximately 1.618. Candidate context windows in **tokens** follow 512 × F(n) beginning with F(3)=2.

| Stage | F(n) | Candidate tokens | Status and condition |
| --- | ---: | ---: | --- |
| F3 | 2 | 1,024 | Operator-observed per-slot llama.cpp allocation; verify repeatability |
| F4 | 3 | 1,536 | Proposed; measure memory and completion success |
| F5 | 5 | 2,560 | Proposed; regress short-context accuracy |
| F6 | 8 | 4,096 | Proposed; long-context accuracy gate |
| F7 | 13 | 6,656 | Proposed; RAM/thermal/latency gate |
| F8 | 21 | 10,752 | Proposed; long-horizon evaluation gate |
| F9 | 34 | 17,408 | Proposed; device and model compatibility gate |
| F10 | 55 | 28,160 | Proposed; remains below reported trained length, not proven usable |
| F11 | 89 | 45,568 | Experimental; exceeds reported 32,768 trained context |

This is a **candidate ladder**, not an automatic resize sequence or guaranteed model behavior. Record model ID/digest, exact tokenizer, training metadata, actual n_ctx/n_ctx_seq, slot/concurrency configuration, available and peak RAM, free storage, temperature/thermal throttling where measurable, p50/p95 latency, prompt processing rate, generated-token rate, crash/restart behavior and task-quality scores. Only advance after a repeated representative benchmark meets an operator-approved threshold. Always retain the previous stable tier and a tested rollback.

Changing llama-server --ctx-size changes **inference allocation only**, not the weights' original training context. Never describe the reported n_ctx_train=32768 as a newly trained Fibonacci context. Use a measured per-sequence context, not a guessed aggregate from the number of slots.

## 3. Adaptive token budgeting

For each request use the real model tokenizer and chat-template overhead. Allocate finite per-request capacity to:
- Recent exact turns: starting relative weight 34;
- Retrieved source-backed relevant excerpts: weight 21;
- Rolling evidence-linked summaries: weight 13;
- Reserved output/continuation budget: weight 21;
- System/safety and task constraints: weight 8.

The weights (34:21:13:21:8) are **heuristic proportions**, not fixed percentages or authority. Enforce minimum safety/system budget and generation reserve first; then normalize or redistribute within the **observed** usable context. Reject impossible budgets clearly and preserve original turns. Never silently truncate user instructions or drop source attribution. Surface token counts, retained and omitted material, compaction reason, tier, max generated tokens, and the continuation cursor.

"Reasoning tokens" internal to a model are not directly controllable by this contract. Scale **explicit** planner, executor and verifier work with bounded batches (1, 2, 3, 5, 8, ...), deadlines, cancellation, checkpointing, resource admission and human review for high-risk steps. A model suggestion never becomes an authorized act merely because an extra reasoning pass was allocated.

## 4. Recursive memory without irreversible forgetting

1. **Roots:** append-only original conversations, files, source revisions, events, operator decisions, model outputs and hashes. Persist durable history independently of the next prompt window.
2. **Trunk:** scoped session, project, user and repository identities, revision graph and access controls. Do not conflate different people or trust domains.
3. **Branches:** rolling summaries and retrieval indexes at candidate Fibonacci turn/age boundaries 1, 2, 3, 5, 8, 13, 21, ...; every summary records parent IDs, original source spans, prompt/model/version, date, quality and whether it is stale.
4. **Leaves:** ranked retrieved passages and exact quotes with pointers and provenance. Relevance cannot override newer verified contradictory evidence.
5. **Flowers:** planned work and unverified hypotheses retain a separate status and source.
6. **Fruits:** verified usable outputs with reproduction command, artifact hash, tests and rollback, never repositories themselves.

History is **logically unbounded across sessions**, subject to explicit storage capacity, retention/deletion policy, backup/recovery, export and access controls. Summaries and embeddings are disposable derived indexes that can be rebuilt; original evidence remains the authoritative record unless the user chooses deletion. Explain when data was excluded due to context limits. Avoid accidental retention of secrets, cross-project memory leakage, unsafe prompt injection from retrieved content, duplicate records and implicit unapproved data sharing.

Proposed evidence contract: aift.memory.fibonacci.v1 records stage, model artifact digest, tokenizer, training_context_reported, inference_context_observed, per-slot allocation, token counts, retrieval provenance, compaction lineage, storage state, status and reason, benchmark environment, verified test outcome, source commit, timestamp and rollback target. This is a **proposed schema**, not an active endpoint or a claim that AIFT-OS currently emits such records.

## 5. Fibonacci training-context curriculum: separate from inference

Training and continued fine-tuning are **not performed** by inference context flags. Candidate curriculum sequence-length buckets follow the same F3 through F10 ladder; F11+ requires separately justified context-extension work. Every stage must establish:
- Authorized trainable model weights or licensed adapters, architecture and tokenizer compatibility, exact checkpoint lineage, evaluation set isolation, training-data permission and provenance;
- Sequence packing/bucketing, measured effective attended length, optimizer configuration, context-extension method, resource budget and checkpoint restoration;
- Repeated next-token-loss, short-context regression, long-range retrieval and multi-document question tests, instruction-following, hallucination/false recall, safety and reasoning checks;
- GPU/CPU RAM, throughput, energy/thermal observations, wall time, cost and rollback evidence;
- Independently reviewed threshold and acceptance decision before promotion.

The 2.0 GB quantized GGUF observed on Android is an **inference artifact**, not proof of an accessible training checkpoint or trainable long-context model. Above the model's native trained length, consider RoPE scaling/interpolation, YaRN-like methods, continued pretraining or different architectures only as separately licensed and evaluated experiments. Context-extension claims require measured generalization, not merely successful allocation or an API reply.

## 6. Infrastructure and workload federation

Scale resource shards, retrieval batches and optional model workers recursively only as **measured nodes** join the federation. Owners and boundaries:
- **AIFT-OS:** versioned admission/tier policy, exact status, orchestration, evidence/compatibility ledger.
- **AIFT-Runtime:** local inference process health, supervised restart, token counters, storage/retrieval adapters and offline resilience.
- **TheMindofAll:** authorized model manifests, weights/adapters, tokenizer/training/evaluation cards, dataset and run provenance.
- **AIFT-Forge:** safe reusable SDK, deterministic offline test fixtures and bounded agent/planner contracts.
- **VPS:** explicitly authorized provider nodes, isolated hosted services, storage/backup, capacity evidence and recovery; no implicit phone-to-cloud route.
- **AIFT-Genesis:** canonical Tree relations and identity inheritance without granting training, custody or model authority.
- **Living Universe / UXI:** source-backed, accessible spiral/tree exploration and honest token/memory visibility; public interface must not expose private operator internals.
- **AetherianGovernance / AI-Freedom-Trust:** existing consent, doctrine and constitutional boundaries remain independently owned.

Do not make geometric proximity a permissions rule, inference evidence, trust score or semantic ranking. Recursive federation requires version negotiation, local-first default, isolated authority scopes, encrypted or access-controlled storage, reviewed replication, deduplication, graceful offline behavior and audited user consent. Hosted and native capabilities must be reported separately.

## 7. Acceptance and verification matrix

| Gate | Minimum required result | Failure evidence and recovery |
| --- | --- | --- |
| LOCAL-01 | Known GGUF hash and aetherion:latest model list; a real bounded completion on 127.0.0.1:8080 | API refusal, HTTP error, timeout, malformed response, model mismatch |
| UXI-01 | Browser -> 8788 Console -> 8080 actual generation with recorded local-inference mode and visible answer | Preserve exact sanitized inference_failure_kind and error; never label every failure "offline" |
| CONTEXT-01 | Exact per-slot context and tokenizer budget measured at existing tier | Context rejection/overflow recovered without losing persistent history |
| MEMORY-01 | At least 100-turn offline test and recovery of the earliest and latest source-linked turns | Inject crash/corruption and prove journal and backup/rebuild behavior |
| RETRIEVE-01 | Provenance links, contradiction/revision tests, privacy boundaries and no fabricated recall | Measure false recall, omission, prompt injection and retrieval drift |
| TRAIN-01 | Separate immutable checkpoint, data rights, evaluation and resource evidence for each actual training tier | Roll back to last qualified checkpoint; do not publish untested weights |
| MOBILE-01 | Android/Termux memory, thermal, p95 latency, restart/low-battery and single-slot vs multi-slot comparison | Hold previous stable window and truthful degraded state |
| FED-01 | At least one real versioned producer->consumer contract test; denial and offline behavior | Version mismatch/absent node reported blocked; no privileged fallback |

Passing a syntax check or documentation link check is **not** proof of a trained model, resumed service, 100-turn durable memory, or hardware release. Use exact commit SHA, actual commands, log excerpts, host/arch, digests, environment and status PASS/FAIL/UNRUN/BLOCKED. Never invent evidence. Sensitive network, identity, keys, training data, deployment, physical-device, destructive and governance changes require explicit owner review.

## 8. Implementation sequence (separate from this documentation-only change)

**P0: Repair and prove the existing console path.** Verify simultaneous model/Console availability; inspect recorded inference errors, short model round trip, context rejection and start/restart behavior; maintain loopback-only exposure. The AIFT-OS inference error classification work in draft PR #95 is a related implementation prerequisite, not merged functionality.

**P1: Persist and budget.** Version a local token-budget planner, create a session/history retrieval adapter backed by existing persisted turns, add source-linked summaries and explicit continuation. Unit tests must demonstrate bounded memory, exact history preservation, redaction and rollback.

**P2: Verify first Fibonacci tier promotion.** Benchmark F3 baseline against proposed F4 in the same device environment. Measure per-slot n_ctx, p95 latency, memory, quality and crashes; do not auto-increase. Repeat against later tiers only after each qualifies.

**P3: Extend training and federation only when justified.** Publish training/evaluation data cards, opt-in checkpoints and versioned provider interoperability with independent safety, privacy and capacity gates.

## 9. Operator safety and non-goals

This documentation PR performs **no** model download, llama.cpp restart, memory allocation, training run, GitHub runner enrollment, credential operation, public binding, repository merge, firmware action or deployment. It does not change existing source code or test expectations. The initial work must remain local-first, source-backed and reviewable; all implementations belong in separate scoped draft PRs with verified exact-head CI and owner approvals.

Related sources: [AIFT platform overview](README.md), [desktop and service contracts](desktop-and-service-contracts.md), [Aetherion worker readiness](../aetherion-worker-readiness.md), [Living Federation Geometry](../LIVING-FEDERATION-GEOMETRY.md), [MoBox UXI operator guide](../MOBOX-UXI.md), and [AIFT-Genesis Fibonacci research](https://github.com/AIFreedomTrustFederation/AIFT-Genesis/blob/main/science/FIBONACCI.md).
