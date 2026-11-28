# Embodied E1 simulated motion gate

**INTEGRATION DRAFT** pending independent verification of the hardened head. This experiment returns Python dictionaries. An accepted proposal yields `SIMULATE_ONLY` and a simulated endpoint. A rejected proposal yields `STOP` without an endpoint. The demo reports `physical_actuation=false`.

## Gate and receiver contract

- Construct one `SafetyGate` per simulation session. The device name must be a nonblank string. Speed, clearance and the symmetric square geofence limit must be positive finite scalars. Maximum age may be zero. Maximum duration and transport deadline must be positive integers. Invalid configuration raises `ValueError` at construction.
- Times, sequences and intervals use built-in integers from 0 through 2⁶³−1. Durations must also be positive. Scalar measurements and velocities accept finite built-in floats or integers of at most 1023 bits. Booleans never count as numbers.
- The receiver supplies `now_ms` separately to the gate, dictionary adapter and JSON adapter. The publisher's `clock_ms` must match that sample. This single-clock fixture checks grant expiry and observation age at evaluation time. Equal receiver timestamps permit distinct increasing sequences. A denied proposal does not consume a sequence, but the receiver clock still advances.
- Both the observed start and the computed endpoint must be inside the inclusive square. Speed, duration and required clearance have inclusive limits. The gate checks displacement against the remaining distance to each edge before returning the unrounded endpoint.
- E-stop, lost or malformed network/watchdog signals, invalid receiver time and receiver clock rollback latch STOP for the lifetime of the gate. E-stop takes precedence. Restored signals, fresh grants and new sequence numbers cannot clear a latch. Recovery means the fixture owner explicitly ends that session and creates a fresh gate (and explicitly arms a fresh transport if used). Other direct proposal denials permit a corrected retry.
- The in-memory transport additionally latches on missing, late, duplicate, reordered or rejected frames, unarmed use and publisher clock or sequence mismatch. Its scheduler must supply ticks. The serialized boundary accepts UTF-8 JSON bytes up to 4096 bytes, enforces exact fields and types, and rejects duplicate keys and nonfinite numbers.

## Evidence and reproduction

This change depends on PR #107 at `f086411007f72e259bacfa1aab723275de5bc69c`. The independent Android ARM64 Ubuntu PRoot run reported 77 passing baseline tests in 0.023 s with Python 3.14.4. The baseline also passed all 77 locally with Python 3.12.14. The hardened suite passed 98 locally, comprising those 77 regression methods and 21 new methods. The original dictionary and byte adapter tests now supply the receiver clock explicitly. Both local demos returned a nominal endpoint of `[0.01, 0.0]`, then a latched STOP. The draft PR records its exact head, canonical verification and hosted CI evidence separately.

From the repository root on the draft head, use Python 3.10+, the Node/pnpm versions in CI and package.json, and Go from go.mod.

```sh
git rev-parse HEAD
python3 -B -m unittest discover -s experiments/embodied -p 'test_*.py' -v
python3 -B experiments/embodied/safety_sim.py
pnpm install --frozen-lockfile --ignore-scripts
pnpm run typecheck
make verify
git diff --check f086411007f72e259bacfa1aab723275de5bc69c...HEAD
git diff --check
test -z "$(git status --porcelain)"
```

The canonical Go suite uses loopback HTTP test servers. If the isolated architecture copy reports an error obtaining VCS status, use `GOFLAGS=-buildvcs=false make verify` to omit Go binary VCS stamping in that copy.

## Limits and rollback

The receiver clock, heartbeat flags, observations and unsigned device grants are fixture inputs. The supervisor must preserve the session and its monotonic clock. The model covers one floating-point step with static clearance. It provides no independent timer, sensor calibration, authentication, continuous dynamics or physical safety certification. ROS/DDS, devices and actuators are outside this experiment. Performance, thermal behavior and energy consumption have not been measured.

To roll back the integration draft, close the child PR. To reproduce the previous simulator, check out the #107 commit above in a separate worktree. Neither procedure requires changing the parent branch or a device.
