"""Strict ROS 2-shaped dictionary adapter; simulation-only, no ROS imports or I/O."""
from safety_sim import Observation, Proposal, Grant, SafetyGate

SCHEMAS = {
    "observation": (Observation, {"sequence": int, "timestamp_ms": int, "x": (int, float), "y": (int, float), "obstacle_m": (int, float)}),
    "proposal": (Proposal, {"sequence": int, "vx_mps": (int, float), "vy_mps": (int, float), "duration_ms": int}),
    "grant": (Grant, {"device": str, "operation": str, "expires_ms": int, "revoked": bool}),
}

def evaluate_envelope(gate: SafetyGate, envelope):
    """Return inert SIMULATE_ONLY or STOP; never a physical actuation command."""
    deny = {"accepted": False, "reason": "invalid_envelope", "action": "STOP", "simulated": True}
    if type(envelope) is not dict or set(envelope) != {"observation", "proposal", "grant", "clock_ms", "network_ok", "watchdog_ok"}:
        return deny.copy()
    if type(envelope["clock_ms"]) is not int or any(type(envelope[k]) is not bool for k in ("network_ok", "watchdog_ok")):
        return deny.copy()
    parsed = {}
    for name, (cls, fields) in SCHEMAS.items():
        data = envelope[name]
        if type(data) is not dict or set(data) != set(fields):
            return deny.copy()
        if any(type(data[k]) not in (expected if type(expected) is tuple else (expected,)) for k, expected in fields.items()):
            return deny.copy()
        parsed[name] = cls(**data)
    return gate.evaluate(parsed["observation"], parsed["proposal"], parsed["grant"],
                         now_ms=envelope["clock_ms"], network_ok=envelope["network_ok"],
                         watchdog_ok=envelope["watchdog_ok"])

# Bounded serialized-message boundary for a future isolated ROS simulator bridge.
# JSON is an interchange fixture only; no ROS/DDS transport or network I/O.
import json

MAX_ENVELOPE_BYTES = 4096

def _unique_object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("duplicate JSON field")
        result[key] = value
    return result

def _reject_constant(value):
    raise ValueError("nonfinite JSON constant")

def evaluate_json_envelope(gate: SafetyGate, payload):
    """Reject malformed/ambiguous/unbounded messages before safety evaluation."""
    deny = {"accepted": False, "reason": "invalid_envelope", "action": "STOP", "simulated": True}
    if type(payload) is not bytes or len(payload) > MAX_ENVELOPE_BYTES:
        return deny.copy()
    try:
        message = json.loads(payload.decode("utf-8", errors="strict"),
                             object_pairs_hook=_unique_object,
                             parse_constant=_reject_constant)
    except (ValueError, UnicodeError, TypeError, RecursionError):
        return deny.copy()
    return evaluate_envelope(gate, message)
