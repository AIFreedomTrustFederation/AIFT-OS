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


class SafetyGate:
    """Deterministic deny-by-default simulated motion gate, not a safety controller."""

    def __init__(self, *, device="virtual-robot-01", max_speed=0.3,
                 max_age_ms=200, max_duration_ms=500, clearance_m=0.5,
                 bound_m=2.0):
        self.device = device
        self.max_speed = max_speed
        self.max_age_ms = max_age_ms
        self.max_duration_ms = max_duration_ms
        self.clearance_m = clearance_m
        self.bound_m = bound_m
        self.last_sequence = -1
        self.estop = False

    def stop(self):
        self.estop = True

    def evaluate(self, obs: Observation, proposal: Proposal, grant: Grant,
                 *, now_ms: int, network_ok=True, watchdog_ok=True):
        def deny(reason):
            return {"accepted": False, "reason": reason, "action": "STOP",
                    "simulated": True}
        if self.estop:
            return deny("estop")
        if network_ok is not True or watchdog_ok is not True:
            return deny("heartbeat_lost")
        if not isinstance(now_ms, int) or isinstance(now_ms, bool):
            return deny("invalid_clock")
        if (not isinstance(grant, Grant) or type(grant.device) is not str or
                type(grant.operation) is not str or type(grant.expires_ms) is not int or
                type(grant.revoked) is not bool or grant.revoked or
                grant.device != self.device or grant.operation != "sim_motion" or
                now_ms >= grant.expires_ms):
            return deny("capability_denied")
        if not isinstance(obs, Observation) or not isinstance(proposal, Proposal):
            return deny("malformed_input")
        if any(type(v) is not int for v in (obs.sequence, obs.timestamp_ms, proposal.sequence, proposal.duration_ms)):
            return deny("malformed_input")
        values = (obs.x, obs.y, obs.obstacle_m, proposal.vx_mps, proposal.vy_mps)
        if any(type(v) not in (int, float) or not math.isfinite(v) for v in values):
            return deny("nonfinite_input")
        if obs.sequence != proposal.sequence or obs.sequence <= self.last_sequence:
            return deny("sequence_invalid")
        if obs.timestamp_ms > now_ms or now_ms - obs.timestamp_ms > self.max_age_ms:
            return deny("stale_observation")
        if not 0 < proposal.duration_ms <= self.max_duration_ms:
            return deny("duration_limit")
        speed = math.hypot(proposal.vx_mps, proposal.vy_mps)
        if speed > self.max_speed:
            return deny("speed_limit")
        if obs.obstacle_m < self.clearance_m + speed * proposal.duration_ms / 1000:
            return deny("collision_clearance")
        dx = proposal.vx_mps * proposal.duration_ms / 1000
        dy = proposal.vy_mps * proposal.duration_ms / 1000
        if abs(obs.x + dx) > self.bound_m or abs(obs.y + dy) > self.bound_m:
            return deny("geofence")
        self.last_sequence = obs.sequence
        return {"accepted": True, "reason": "bounded_simulated_motion",
                "action": "SIMULATE_ONLY", "simulated": True,
                "next_position": [round(obs.x + dx, 6), round(obs.y + dy, 6)]}


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
