# Embodied E1: deterministic simulated motion gate (research prototype)

**Status:** INTEGRATION DRAFT / tests UNRUN on exact GitHub head until CI or an independent runner executes them. This is not ROS 2 integration, a phone OS, or a certified/physical safety controller.

## Target and authority
- Milestone: E1 preliminary pure-Python simulator contract; device `virtual-robot-01`; no sensors, ROS transport, actuators or network.
- Base: `31012097ec2f8dfdded176693d9362c62bc71f44` (AIFT-OS main observed 2026-11-22).
- Trust boundary: model output is only an untrusted proposal; deterministic gate decides; `SIMULATE_ONLY` is a data label and never triggers hardware.
- Device-specific grant: `device`, `operation=sim_motion`, `expires_ms`, `revoked`. These are simulation assertions, not signed capabilities or identity proof.
- Dependencies: AIFT-OS owns this experimental contract; future ROS 2 adapter requires separate simulator/owner integration. AIFT-Runtime owns model adapter, not this gate.

## Hypothesis and baseline
A deterministic gate can reject invalid/stale/replayed/motion-unsafe proposals independently of model availability. Baseline is **no embodied executable gate in the inspected E0 roadmap**; no measured comparison against a ROS 2 simulator is claimed. The bounded goal is 1 accepted nominal move and 17 negative cases, including e-stop, network/watchdog loss, expiry/revocation, device/operation mismatch, stale/future perception, sequence mismatch, speed, duration, clearance, geofence, non-finite and malformed input. An untrusted model could propose any value, including NaN; the gate must reject it.

## Reproduction
From repo root, Python 3.10+:

```sh
python3 -m unittest discover -s experiments/embodied -p 'test_*.py' -v
python3 experiments/embodied/safety_sim.py
git diff --check
```

Expected nominal response: `accepted=true`, `action=SIMULATE_ONLY`, `next_position=[0.01,0]`. Expected latched stop response: `accepted=false`, `action=STOP`, `reason=estop`. Tests are source-defined expected behavior, **not an observed result** until executed.

## Limitations and next gate
The simulator uses a single static observation and one-step kinematics, not a ROS 2 URDF robot, dynamics, collision mesh, real watchdog, signed permission, calibrated sensor, continuous trajectory or safety-rated motor controller. `network_ok` and `watchdog_ok` are injected booleans. This cannot establish actual physical safety, real-time deadlines or an independent process boundary. Memory, energy, thermal and battery metrics are not measured.

**Rollback:** close the draft and delete its feature branch; main and device state are untouched.

**One next action:** execute the 18 stdlib tests on exact PR head and attach the command output, Python version and SHA before considering a ROS 2 simulator adapter.
