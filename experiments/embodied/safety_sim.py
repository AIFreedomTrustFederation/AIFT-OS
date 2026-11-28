#!/usr/bin/env python3
"""E1 simulated robot safety gate. Stdlib only; NEVER connects to hardware.

Run: python3 -m unittest discover -s experiments/embodied -p 'test_*.py' -v
Demo: python3 experiments/embodied/safety_sim.py
"""
from __future__ import annotations
from dataclasses import dataclass
import json
import math
import time


@dataclass(frozen=True)
class Observation:
    sequence: int
    timestamp_ms: int
    x: float
    y: float
    obstacle_m: float


@dataclass(frozen=True)
class Proposal:
    sequence: int
    vx_mps: float
    vy_mps: float
    duration_ms: int


@dataclass(frozen=True)
class Grant:
    device: str
    operation: str
    expires_ms: int
    revoked: bool = False


# Nonnegative signed-64 range for simulated times, intervals and sequences.
MAX_COUNTER = (1 << 63) - 1


def valid_counter(value):
    return type(value) is int and 0 <= value <= MAX_COUNTER


def finite_scalar(value):
    # Bound integers before calling any C double conversion (including isfinite).
    if type(value) is int:
        return value.bit_length() <= 1023
    return type(value) is float and math.isfinite(value)


def deny(reason):
    return {"accepted": False, "reason": reason, "action": "STOP", "simulated": True}


@dataclass(frozen=True, kw_only=True, eq=False)
class GateLimits:
    device: str = "virtual-robot-01"
    max_speed: float = 0.3
    max_age_ms: int = 200
    max_duration_ms: int = 500
    clearance_m: float = 0.5
    bound_m: float = 2.0

    def __post_init__(self):
        if type(self.device) is not str or not self.device.strip():
            raise ValueError("device must be a nonempty name")
        for name in ("max_speed", "clearance_m", "bound_m"):
            value = getattr(self, name)
            if not finite_scalar(value) or value <= 0:
                raise ValueError(f"{name} must be finite and positive")
        if not valid_counter(self.max_age_ms):
            raise ValueError("invalid max_age_ms")
        if not valid_counter(self.max_duration_ms) or self.max_duration_ms == 0:
            raise ValueError("invalid max_duration_ms")


class SafetyGate(GateLimits):
    """One simulation session. Critical faults require ending the session."""

    def __init__(self, **limits):
        super().__init__(**limits)
        self._last_sequence = -1
        self._last_clock = None
        self._fault = None

    @property
    def last_sequence(self):
        return self._last_sequence

    @property
    def estop(self):
        return self._fault == "estop"

    def stop(self):
        self._fault = "estop"

    def supervise(self, now_ms, *, network_ok=True, watchdog_ok=True):
        """Record receiver time even if the subsequent proposal is rejected."""
        if self._fault is None:
            if network_ok is not True or watchdog_ok is not True:
                self._fault = "heartbeat_lost"
            elif not valid_counter(now_ms):
                self._fault = "invalid_clock"
            elif self._last_clock is not None and now_ms < self._last_clock:
                self._fault = "clock_rollback"
            else:
                self._last_clock = now_ms
        return deny(self._fault) if self._fault is not None else None

    def evaluate(self, obs: Observation, proposal: Proposal, grant: Grant,
                 *, now_ms: int, network_ok=True, watchdog_ok=True):
        stopped = self.supervise(now_ms, network_ok=network_ok, watchdog_ok=watchdog_ok)
        if stopped is not None:
            return stopped
        if (type(grant) is not Grant or type(grant.device) is not str or
                type(grant.operation) is not str or not valid_counter(grant.expires_ms) or
                type(grant.revoked) is not bool or grant.revoked or
                grant.device != self.device or grant.operation != "sim_motion" or
                now_ms >= grant.expires_ms):
            return deny("capability_denied")
        if type(obs) is not Observation or type(proposal) is not Proposal:
            return deny("malformed_input")
        if any(not valid_counter(v) for v in (obs.sequence, obs.timestamp_ms, proposal.sequence, proposal.duration_ms)):
            return deny("malformed_input")
        values = (obs.x, obs.y, obs.obstacle_m, proposal.vx_mps, proposal.vy_mps)
        if any(not finite_scalar(v) for v in values):
            return deny("nonfinite_input")
        if obs.sequence != proposal.sequence or obs.sequence <= self.last_sequence:
            return deny("sequence_invalid")
        if obs.timestamp_ms > now_ms or now_ms - obs.timestamp_ms > self.max_age_ms:
            return deny("stale_observation")
        if not 0 < proposal.duration_ms <= self.max_duration_ms:
            return deny("duration_limit")
        if abs(obs.x) > self.bound_m or abs(obs.y) > self.bound_m:
            return deny("geofence")
        speed = math.hypot(proposal.vx_mps, proposal.vy_mps)
        if speed > self.max_speed:
            return deny("speed_limit")
        # Convert before multiplying: arbitrary-precision integer products can
        # overflow on division. Infinite float intermediates fail closed.
        travel = speed * proposal.duration_ms / 1000
        if obs.obstacle_m < self.clearance_m + travel:
            return deny("collision_clearance")
        dx = float(proposal.vx_mps) * proposal.duration_ms / 1000
        dy = float(proposal.vy_mps) * proposal.duration_ms / 1000
        # Test remaining room before addition can round a small outward step
        # back onto the boundary. Also check the coordinates we actually emit.
        if not (-self.bound_m - obs.x <= dx <= self.bound_m - obs.x and
                -self.bound_m - obs.y <= dy <= self.bound_m - obs.y):
            return deny("geofence")
        next_x, next_y = obs.x + dx, obs.y + dy
        if abs(next_x) > self.bound_m or abs(next_y) > self.bound_m:
            return deny("geofence")
        self._last_sequence = obs.sequence
        return {"accepted": True, "reason": "bounded_simulated_motion",
                "action": "SIMULATE_ONLY", "simulated": True,
                "next_position": [next_x, next_y]}


def demo():
    gate = SafetyGate()
    now = int(time.monotonic() * 1000)
    obs = Observation(1, now, 0, 0, 2)
    grant = Grant("virtual-robot-01", "sim_motion", now + 1000)
    safe = gate.evaluate(obs, Proposal(1, 0.1, 0, 100), grant, now_ms=now)
    gate.stop()
    stopped = gate.evaluate(Observation(2, now, 0, 0, 2),
                            Proposal(2, 0.1, 0, 100), grant, now_ms=now)
    print(json.dumps({"profile": "pure-python-virtual-robot", "safe": safe,
                      "estop": stopped, "physical_actuation": False}, sort_keys=True))


if __name__ == "__main__":
    demo()
