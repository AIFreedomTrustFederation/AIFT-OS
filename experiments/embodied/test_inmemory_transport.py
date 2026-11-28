"""In-memory fault-injection transport; simulation-only, no sockets, ROS or actuators.

Run: python3 -m unittest discover -s experiments/embodied -p 'test_*.py' -v
"""
import json
import unittest

from ros_contract_adapter import decode_envelope, evaluate_envelope
from safety_sim import SafetyGate, valid_counter

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
        if not valid_counter(deadline_ms) or deadline_ms == 0:
            raise ValueError("deadline_ms must be a positive simulated interval")
        self.gate = SafetyGate()
        self.expected = 1
        self._deadline_ms = deadline_ms
        self.last_delivery_ms = None
        self.latched = False
        self.last_tick_ms = None
        self.armed_at_ms = None

    @property
    def deadline_ms(self):
        return self._deadline_ms

    def arm(self, now_ms):
        """Explicitly begin startup supervision; no reset or rearm after a fault."""
        stop = {"accepted": False, "action": "STOP", "simulated": True,
                "reason": "transport_fault"}
        if self.latched or self.armed_at_ms is not None or self.last_delivery_ms is not None or not valid_counter(now_ms) or (self.last_tick_ms is not None and now_ms < self.last_tick_ms):
            self.latched = True
            self.gate.stop()
            return stop
        self.armed_at_ms = now_ms
        self.last_tick_ms = now_ms
        return {"accepted": True, "action": "SIMULATE_ONLY", "simulated": True,
                "reason": "watchdog_armed"}

    def tick(self, now_ms):
        """External scheduler must call this with monotonic milliseconds.

        This method does not create an independent timer or safety-rated watchdog.
        """
        stop = {"accepted": False, "action": "STOP", "simulated": True,
                "reason": "transport_timeout"}
        if self.latched:
            return stop.copy()
        if (self.armed_at_ms is None or not valid_counter(now_ms) or
                (self.last_tick_ms is not None and now_ms < self.last_tick_ms) or
                (self.last_delivery_ms is not None and
                 (now_ms < self.last_delivery_ms or
                  now_ms - self.last_delivery_ms > self.deadline_ms))):
            self.latched = True
            self.gate.stop()
            return stop.copy()
        if (self.armed_at_ms is not None and self.last_delivery_ms is None and
                now_ms - self.armed_at_ms > self.deadline_ms):
            self.latched = True
            self.gate.stop()
            return stop.copy()
        self.last_tick_ms = now_ms
        return {"accepted": True, "action": "SIMULATE_ONLY", "simulated": True,
                "reason": "watchdog_healthy"}

    def deliver(self, payload, *, delivery_ms, sequence):
        stop = {"accepted": False, "action": "STOP", "simulated": True,
                "reason": "transport_fault"}
        if self.latched:
            return stop.copy()
        if (self.armed_at_ms is None or not valid_counter(delivery_ms) or not valid_counter(sequence) or
                (self.last_tick_ms is not None and delivery_ms < self.last_tick_ms) or
                sequence != self.expected or
                (self.armed_at_ms is not None and self.last_delivery_ms is None and
                 (delivery_ms < self.armed_at_ms or delivery_ms - self.armed_at_ms > self.deadline_ms)) or
                (self.last_delivery_ms is not None and
                 (delivery_ms < self.last_delivery_ms or
                  delivery_ms - self.last_delivery_ms > self.deadline_ms))):
            self.latched = True
            self.gate.stop()
            return stop.copy()
        # Bind publisher clock and sequences to receiver fixture authority.
        try:
            message = decode_envelope(payload)
            if (type(message) is not dict or
                    type(message.get("clock_ms")) is not int or
                    message["clock_ms"] != delivery_ms or
                    type(message.get("observation")) is not dict or
                    type(message.get("proposal")) is not dict or
                    type(message["observation"].get("sequence")) is not int or
                    type(message["proposal"].get("sequence")) is not int or
                    message["observation"]["sequence"] != sequence or
                    message["proposal"]["sequence"] != sequence):
                raise ValueError("clock or sequence mismatch")
        except (ValueError, TypeError, UnicodeError, RecursionError, AttributeError):
            self.latched = True
            self.gate.stop()
            return stop.copy()
        result = evaluate_envelope(self.gate, message, now_ms=delivery_ms)
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
    def test_unarmed_delivery_latches(self):
        t = InMemoryTransport()
        self.assertEqual(t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)["action"], "STOP")
        self.assertEqual(t.arm(1001)["action"], "STOP")
        self.assertEqual(t.deliver(encoded(frame(1, 1002)), delivery_ms=1002, sequence=1)["action"], "STOP")

    def test_unarmed_tick_latches(self):
        t = InMemoryTransport()
        self.assertEqual(t.tick(1000)["action"], "STOP")
        self.assertEqual(t.arm(1001)["action"], "STOP")

    def test_nominal_two_frames(self):
        t = InMemoryTransport()
        t.arm(1000)
        self.assertEqual(t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)["action"], "SIMULATE_ONLY")
        self.assertEqual(t.deliver(encoded(frame(2, 1100)), delivery_ms=1100, sequence=2)["action"], "SIMULATE_ONLY")

    def test_duplicate_latches(self):
        t = InMemoryTransport()
        t.arm(1000)
        t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)
        self.assertEqual(t.deliver(encoded(frame(1, 1000)), delivery_ms=1001, sequence=1)["action"], "STOP")
        self.assertEqual(t.deliver(encoded(frame(2, 1100)), delivery_ms=1100, sequence=2)["action"], "STOP")

    def test_reordered_latches(self):
        t = InMemoryTransport()
        t.arm(1000)
        self.assertEqual(t.deliver(encoded(frame(2, 1100)), delivery_ms=1100, sequence=2)["action"], "STOP")
        self.assertEqual(t.deliver(encoded(frame(1, 1000)), delivery_ms=1101, sequence=1)["action"], "STOP")

    def test_delayed_latches(self):
        t = InMemoryTransport()
        t.arm(1000)
        t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)
        self.assertEqual(t.deliver(encoded(frame(2, 1300)), delivery_ms=1300, sequence=2)["action"], "STOP")

    def test_drop_latches(self):
        t = InMemoryTransport()
        t.arm(1000)
        t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)
        self.assertEqual(t.missing()["action"], "STOP")
        self.assertEqual(t.deliver(encoded(frame(2, 1100)), delivery_ms=1100, sequence=2)["action"], "STOP")

    def test_malformed_latches(self):
        t = InMemoryTransport()
        t.arm(1000)
        self.assertEqual(t.deliver(b'{"bad":', delivery_ms=1000, sequence=1)["action"], "STOP")
        self.assertEqual(t.deliver(encoded(frame(1, 1000)), delivery_ms=1001, sequence=1)["action"], "STOP")

    def test_clock_rollback_latches(self):
        t = InMemoryTransport()
        t.arm(1000)
        t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)
        self.assertEqual(t.deliver(encoded(frame(2, 1100)), delivery_ms=999, sequence=2)["action"], "STOP")

    def test_transport_sequence_mismatch_latches(self):
        t = InMemoryTransport()
        t.arm(1000)
        self.assertEqual(t.deliver(encoded(frame(2, 1000)), delivery_ms=1000, sequence=1)["action"], "STOP")
        self.assertEqual(t.deliver(encoded(frame(1, 1001)), delivery_ms=1001, sequence=1)["action"], "STOP")

    def test_proposal_sequence_mismatch_latches(self):
        t = InMemoryTransport()
        t.arm(1000)
        message = frame(1, 1000)
        message["proposal"]["sequence"] = 2
        self.assertEqual(t.deliver(encoded(message), delivery_ms=1000, sequence=1)["action"], "STOP")

    def test_silent_publisher_timeout_latches(self):
        t = InMemoryTransport(deadline_ms=200)
        t.arm(1000)
        self.assertEqual(t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)["action"], "SIMULATE_ONLY")
        self.assertEqual(t.tick(1200)["action"], "SIMULATE_ONLY")
        self.assertEqual(t.tick(1201)["action"], "STOP")
        self.assertEqual(t.deliver(encoded(frame(2, 1202)), delivery_ms=1202, sequence=2)["action"], "STOP")

    def test_tick_rollback_latches(self):
        t = InMemoryTransport()
        t.arm(1000)
        self.assertEqual(t.tick(1000)["action"], "SIMULATE_ONLY")
        self.assertEqual(t.tick(999)["action"], "STOP")

    def test_invalid_tick_latches(self):
        t = InMemoryTransport()
        t.arm(1000)
        self.assertEqual(t.tick(True)["action"], "STOP")

    def test_armed_startup_silence_latches(self):
        t = InMemoryTransport(deadline_ms=200)
        self.assertEqual(t.arm(1000)["action"], "SIMULATE_ONLY")
        self.assertEqual(t.tick(1200)["action"], "SIMULATE_ONLY")
        self.assertEqual(t.tick(1201)["action"], "STOP")
        self.assertEqual(t.deliver(encoded(frame(1, 1202)), delivery_ms=1202, sequence=1)["action"], "STOP")

    def test_armed_first_delivery_after_deadline_latches(self):
        t = InMemoryTransport(deadline_ms=200)
        t.arm(1000)
        self.assertEqual(t.deliver(encoded(frame(1, 1201)), delivery_ms=1201, sequence=1)["action"], "STOP")

    def test_armed_first_delivery_within_deadline(self):
        t = InMemoryTransport(deadline_ms=200)
        t.arm(1000)
        self.assertEqual(t.deliver(encoded(frame(1, 1200)), delivery_ms=1200, sequence=1)["action"], "SIMULATE_ONLY")

    def test_rearm_does_not_clear_latched_stop(self):
        t = InMemoryTransport(deadline_ms=200)
        t.arm(1000)
        t.tick(1201)
        self.assertEqual(t.arm(1202)["action"], "STOP")
        self.assertEqual(t.tick(1203)["action"], "STOP")

    def test_publisher_clock_mismatch_latches(self):
        for clock in (999, 1001, True, 1000.0, None):
            with self.subTest(clock=clock):
                t = InMemoryTransport()
                t.arm(1000)
                message = frame(1, 1000)
                message["clock_ms"] = clock
                self.assertEqual(t.deliver(encoded(message), delivery_ms=1000, sequence=1)["action"], "STOP")
                self.assertTrue(t.latched)
                self.assertEqual(t.deliver(encoded(frame(1, 1001)), delivery_ms=1001, sequence=1)["action"], "STOP")

    def test_expired_grant_cannot_use_old_publisher_clock(self):
        t = InMemoryTransport()
        t.arm(1000)
        message = frame(1, 1000)
        message["grant"]["expires_ms"] = 1050
        self.assertEqual(t.deliver(encoded(message), delivery_ms=1100, sequence=1)["action"], "STOP")
        self.assertTrue(t.latched)

    def test_matching_clock_expired_grant_latches(self):
        t = InMemoryTransport()
        t.arm(1000)
        message = frame(1, 1100)
        message["grant"]["expires_ms"] = 1050
        self.assertEqual(t.deliver(encoded(message), delivery_ms=1100, sequence=1)["action"], "STOP")
        self.assertTrue(t.latched)

    def test_no_physical_effects(self):
        t = InMemoryTransport()
        t.arm(1000)
        result = t.deliver(encoded(frame(1, 1000)), delivery_ms=1000, sequence=1)
        self.assertTrue(result["simulated"])
        self.assertNotIn("physical_actuation", result)

if __name__ == "__main__":
    unittest.main()
