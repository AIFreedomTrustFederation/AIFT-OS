"""In-memory fault-injection transport; simulation-only, no sockets, ROS or actuators.

Run: python3 -m unittest discover -s experiments/embodied -p 'test_*.py' -v
"""
import copy
import json
import unittest

from ros_contract_adapter import evaluate_json_envelope, _unique_object, _reject_constant
from safety_sim import SafetyGate

def frame(seq, timestamp_ms):
    return {"observation": {"sequence": seq, "timestamp_ms": timestamp_ms,
                            "x": 0.0, "y": 0.0, "obstacle_m": 2.0},
            "proposal": {"sequence": seq, "vx_mps": 0.1, "vy_mps": 0.0,
                         "duration_ms": 100},
            "grant": {"device": "virtual-robot-01", "operation": "sim_motion",
                      "expires_ms": 10000, "revoked": False},
            "clock_ms": timestamp_ms, "network_ok": True, "watchdog_ok": True}

def encoded(message):
    return json.dumps(message, allow_nan=False, sort_keys=True).encode("utf-8")

class InMemoryTransport:
    """Strict ordered, deadline-bounded delivery. STOP latches on any gap/fault.

    The clock and grant are fixture assertions, not authenticated time or identity.
    """
    def __init__(self, *, deadline_ms=200):
        self.gate = SafetyGate()
        self.expected = 1
        self.deadline_ms = deadline_ms
        self.last_delivery_ms = None
        self.latched = False

    def deliver(self, payload, *, delivery_ms, sequence):
        stop = {"accepted": False, "action": "STOP", "simulated": True,
                "reason": "transport_fault"}
        if self.latched:
            return stop.copy()
        if (type(delivery_ms) is not int or type(sequence) is not int or
                delivery_ms < 0 or sequence != self.expected or
                (self.last_delivery_ms is not None and
                 (delivery_ms < self.last_delivery_ms or
                  delivery_ms - self.last_delivery_ms > self.deadline_ms))):
            self.latched = True
            self.gate.stop()
            return stop.copy()
        # Bind transport sequence to both embedded sequences before evaluation.
        try:
            if type(payload) is not bytes or len(payload) > 4096:
                raise ValueError("invalid payload")
            message = json.loads(payload.decode("utf-8", errors="strict"),
                                 object_pairs_hook=_unique_object,
                                 parse_constant=_reject_constant)
            if (type(message) is not dict or
                    type(message.get("observation")) is not dict or
                    type(message.get("proposal")) is not dict or
                    type(message["observation"].get("sequence")) is not int or
                    type(message["proposal"].get("sequence")) is not int or
                    message["observation"]["sequence"] != sequence or
                    message["proposal"]["sequence"] != sequence):
                raise ValueError("sequence mismatch")
        except (ValueError, TypeError, UnicodeError, RecursionError, AttributeError):
            self.latched = True
            self.gate.stop()
            return stop.copy()
        result = evaluate_json_envelope(self.gate, payload)
        if result["action"] != "SIMULATE_ONLY":
            self.latched = True
            self.gate.stop()
            return result
        self.expected += 1
        self.last_delivery_ms = delivery_ms
        return result

    def missing(self):
        self.latched = True
        self.gate.stop()
        return {"accepted": False, "action": "STOP", "simulated": True,
                "reason": "transport_missing"}

class TransportFaultTests(unittest.TestCase):
    def test_nominal_two_frames(self):
        t = InMemoryTransport()
        self.assertEqual(t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)["action"], "SIMULATE_ONLY")
        self.assertEqual(t.deliver(encoded(frame(2, 1100)), delivery_ms=1100, sequence=2)["action"], "SIMULATE_ONLY")

    def test_duplicate_latches(self):
        t = InMemoryTransport()
        t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)
        self.assertEqual(t.deliver(encoded(frame(1, 1000)), delivery_ms=1001, sequence=1)["action"], "STOP")
        self.assertEqual(t.deliver(encoded(frame(2, 1100)), delivery_ms=1100, sequence=2)["action"], "STOP")

    def test_reordered_latches(self):
        t = InMemoryTransport()
        self.assertEqual(t.deliver(encoded(frame(2, 1100)), delivery_ms=1100, sequence=2)["action"], "STOP")
        self.assertEqual(t.deliver(encoded(frame(1, 1000)), delivery_ms=1101, sequence=1)["action"], "STOP")

    def test_delayed_latches(self):
        t = InMemoryTransport()
        t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)
        self.assertEqual(t.deliver(encoded(frame(2, 1300)), delivery_ms=1300, sequence=2)["action"], "STOP")

    def test_drop_latches(self):
        t = InMemoryTransport()
        t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)
        self.assertEqual(t.missing()["action"], "STOP")
        self.assertEqual(t.deliver(encoded(frame(2, 1100)), delivery_ms=1100, sequence=2)["action"], "STOP")

    def test_malformed_latches(self):
        t = InMemoryTransport()
        self.assertEqual(t.deliver(b'{"bad":', delivery_ms=1000, sequence=1)["action"], "STOP")
        self.assertEqual(t.deliver(encoded(frame(1, 1000)), delivery_ms=1001, sequence=1)["action"], "STOP")

    def test_clock_rollback_latches(self):
        t = InMemoryTransport()
        t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)
        self.assertEqual(t.deliver(encoded(frame(2, 1100)), delivery_ms=999, sequence=2)["action"], "STOP")

    def test_transport_sequence_mismatch_latches(self):
        t = InMemoryTransport()
        self.assertEqual(t.deliver(encoded(frame(2, 1000)), delivery_ms=1000, sequence=1)["action"], "STOP")
        self.assertEqual(t.deliver(encoded(frame(1, 1001)), delivery_ms=1001, sequence=1)["action"], "STOP")

    def test_proposal_sequence_mismatch_latches(self):
        t = InMemoryTransport()
        message = frame(1, 1000)
        message["proposal"]["sequence"] = 2
        self.assertEqual(t.deliver(encoded(message), delivery_ms=1000, sequence=1)["action"], "STOP")

    def test_no_physical_effects(self):
        t = InMemoryTransport()
        result = t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)
        self.assertTrue(result["simulated"])
        self.assertNotIn("physical_actuation", result)

if __name__ == "__main__":
    unittest.main()
