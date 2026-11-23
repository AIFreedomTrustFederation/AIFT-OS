"""Simulation-only strict ROS 2-shaped contract tests, no ROS transport."""
import unittest
from safety_sim import SafetyGate
from ros_contract_adapter import evaluate_envelope

def fixture():
    return {"observation": {"sequence": 1, "timestamp_ms": 1000, "x": 0.0, "y": 0.0, "obstacle_m": 2.0},
            "proposal": {"sequence": 1, "vx_mps": 0.1, "vy_mps": 0.0, "duration_ms": 100},
            "grant": {"device": "virtual-robot-01", "operation": "sim_motion", "expires_ms": 2000, "revoked": False},
            "clock_ms": 1000, "network_ok": True, "watchdog_ok": True}

class AdapterTests(unittest.TestCase):
    def test_nominal(self):
        result = evaluate_envelope(SafetyGate(), fixture())
        self.assertEqual(result["action"], "SIMULATE_ONLY")
        self.assertTrue(result["simulated"])
    def test_missing_field(self):
        data = fixture(); del data["proposal"]["duration_ms"]
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")
    def test_extra_field(self):
        data = fixture(); data["grant"]["admin"] = True
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")
    def test_boolean_sequence(self):
        data = fixture(); data["observation"]["sequence"] = True
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")
    def test_string_watchdog(self):
        data = fixture(); data["watchdog_ok"] = "true"
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")
    def test_nonfinite(self):
        data = fixture(); data["proposal"]["vx_mps"] = float("nan")
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")
    def test_replay(self):
        gate = SafetyGate(); data = fixture()
        self.assertTrue(evaluate_envelope(gate, data)["accepted"])
        self.assertEqual(evaluate_envelope(gate, data)["action"], "STOP")
    def test_revoked(self):
        data = fixture(); data["grant"]["revoked"] = True
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")
    def test_estop(self):
        gate = SafetyGate(); gate.stop()
        self.assertEqual(evaluate_envelope(gate, fixture())["action"], "STOP")
    def test_network_loss(self):
        data = fixture(); data["network_ok"] = False
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")

    def test_top_level_unknown_field(self):
        data = fixture(); data["physical_actuation"] = True
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")
    def test_nested_wrong_container(self):
        data = fixture(); data["observation"] = []
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")
    def test_future_observation(self):
        data = fixture(); data["observation"]["timestamp_ms"] = 1001
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")
    def test_expired_capability(self):
        data = fixture(); data["grant"]["expires_ms"] = 1000
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")
    def test_geofence(self):
        data = fixture(); data["observation"]["x"] = 2.0
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")
    def test_rejection_does_not_consume_sequence(self):
        gate = SafetyGate(); data = fixture()
        data["proposal"]["vx_mps"] = 100.0
        self.assertEqual(evaluate_envelope(gate, data)["action"], "STOP")
        data["proposal"]["vx_mps"] = 0.1
        self.assertEqual(evaluate_envelope(gate, data)["action"], "SIMULATE_ONLY")
    def test_invalid_clock_boolean(self):
        data = fixture(); data["clock_ms"] = True
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")
    def test_zero_duration(self):
        data = fixture(); data["proposal"]["duration_ms"] = 0
        self.assertEqual(evaluate_envelope(SafetyGate(), data)["action"], "STOP")
