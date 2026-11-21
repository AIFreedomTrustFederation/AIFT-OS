# First gate experiment: missing-stage detection

Observed locally on 2026-11-21 using `scripts/demo_gate_experiment.py`.
Six temporary Git fixtures were exported and executed with installed npm/Node;
no dependencies were installed and no upstream project builds were executed.

| Fixture | Optional baseline | Mandatory candidate |
| --- | --- | --- |
| Healthy stages | success, exit 0 | success, exit 0 |
| Missing typecheck stage | success, exit 0 (missed defect) | failed, exit 1 (detected defect) |
| Test stage exits 7 | failed, exit 7 | failed, exit 7 |

All six expected outcomes matched. This establishes the narrow value of the
existing Forge PR #15 command correction on synthetic stages. It is not a
statistical reliability estimate, real typecheck/build result, or speedup claim.
The demo emits fixture SHAs, argv, outcomes, timings and logs as JSON for reruns.

Command sources:

- [Baseline manifest](https://github.com/AIFreedomTrustFederation/AIFT-Forge/blob/584781172a48c2d8ed1aaff70ae543a70f7822d6/aift.repo.json)
- [Candidate manifest](https://github.com/AIFreedomTrustFederation/AIFT-Forge/blob/c98b856ca6046c953417064a0411861f8fd98e6c/aift.repo.json)
- [Existing Forge PR #15](https://github.com/AIFreedomTrustFederation/AIFT-Forge/pull/15), still open when inspected.

Related live verification inspection:

- Analyzer PR #78 was already merged when inspected. Its original head
  `c7762c336e22b97eb83618fb5c6b6e943c1b10d6` had successful `test`, `collector`,
  and `bootstrap` checks, while `live-producer-evidence` was skipped.
  [Test run](https://github.com/AIFreedomTrustFederation/AIFT-OS/actions/runs/44889250835).
- Forge PR #15 head `c98b856ca6046c953417064a0411861f8fd98e6c` had successful
  [qa](https://github.com/AIFreedomTrustFederation/AIFT-Forge/actions/runs/44214568120).
  Review also found README states the active root typecheck/build scripts only
  print success. A mandatory chain alone does not repair those implementations.

Complete-checkout analysis and canonical OS validation remain unavailable in
this cloud workspace: cloning fails through the configured proxy, and direct
GitHub access fails DNS resolution. The connected GitHub tools provided pinned
source and live CI observations. The runner's local fixtures are complete local
Git repositories; they are not substitutes for the upstream checkouts.
