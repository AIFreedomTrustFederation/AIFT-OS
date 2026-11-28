"""Strict ROS 2-shaped dictionary adapter; simulation-only, no ROS imports or I/O."""
import json

from safety_sim import Observation, Proposal, Grant, SafetyGate, deny, finite_scalar, valid_counter

SCHEMAS = {
    "observation": (Observation, {"sequence": int, "timestamp_ms": int, "x": (int, float), "y": (int, float), "obstacle_m": (int, float)}),
    "proposal": (Proposal, {"sequence": int, "vx_mps": (int, float), "vy_mps": (int, float), "duration_ms": int}),
    "grant": (Grant, {"device": str, "operation": str, "expires_ms": int, "revoked": bool}),
}

def evaluate_envelope(gate: SafetyGate, envelope, *, now_ms=None):
    """The receiver supplies now_ms. The publisher can only echo that time."""
    if type(envelope) is not dict:
        return gate.supervise(now_ms) or deny("invalid_envelope")
    stopped = gate.supervise(now_ms, network_ok=envelope.get("network_ok"),
                             watchdog_ok=envelope.get("watchdog_ok"))
    if stopped is not None:
        return stopped
    if set(envelope) != {"observation", "proposal", "grant", "clock_ms", "network_ok", "watchdog_ok"}:
        return deny("invalid_envelope")
    if not valid_counter(envelope["clock_ms"]) or envelope["clock_ms"] != now_ms:
        return deny("clock_mismatch")
    parsed = {}
    for name, (cls, fields) in SCHEMAS.items():
        data = envelope[name]
        if type(data) is not dict or set(data) != set(fields):
            return deny("invalid_envelope")
        if any(type(data[k]) not in (expected if type(expected) is tuple else (expected,)) for k, expected in fields.items()):
            return deny("invalid_envelope")
        parsed[name] = cls(**data)
    return gate.evaluate(parsed["observation"], parsed["proposal"], parsed["grant"],
                         now_ms=now_ms)


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

def _finite_json_float(token):
    value = float(token)
    if not finite_scalar(value):
        raise ValueError("nonfinite JSON number")
    return value


def decode_envelope(payload):
    """Decode bounded, unambiguous UTF-8 JSON or raise ValueError."""
    if type(payload) is not bytes or len(payload) > MAX_ENVELOPE_BYTES:
        raise ValueError("invalid payload")
    try:
        return json.loads(payload.decode("utf-8", errors="strict"),
                          object_pairs_hook=_unique_object, parse_constant=_reject_constant,
                          parse_float=_finite_json_float)
    except RecursionError as exc:
        raise ValueError("excessive JSON nesting") from exc


def evaluate_json_envelope(gate: SafetyGate, payload, *, now_ms=None):
    try:
        message = decode_envelope(payload)
    except ValueError:
        return gate.supervise(now_ms) or deny("invalid_envelope")
    return evaluate_envelope(gate, message, now_ms=now_ms)
