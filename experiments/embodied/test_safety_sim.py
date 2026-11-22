"""Negative and nominal tests for simulated motion. No external dependencies."""
import unittest
from safety_sim import SafetyGate, Observation, Proposal, Grant


class GateTests(unittest.TestCase):
    def setUp(self):
        self.gate = SafetyGate()
        self.obs = Observation(1, 1000, 0, 0, 2)
        self.proposal = Proposal(1, .1, 0, 100)
        self.grant = Grant("virtual-robot-01", "sim_motion", 2000)

    def check(self, reason, *, obs=None, proposal=None, grant=None, now_ms=1000,
              network_ok=True, watchdog_ok=True):
        result = self.gate.evaluate(
            self.obs if obs is None else obs,
            self.proposal if proposal is None else proposal,
            self.grant if grant is None else grant,
            now_ms=now_ms, network_ok=network_ok, watchdog_ok=watchdog_ok)
        self.assertFalse(result["accepted"])
        self.assertEqual(result["action"], "STOP")
        self.assertEqual(result["reason"], reason)

    def test_nominal_simulated_move(self):
        r = self.gate.evaluate(self.obs, self.proposal, self.grant, now_ms=1000)
        self.assertTrue(r["accepted"])
        self.assertEqual(r["action"], "SIMULATE_ONLY")
        self.assertEqual(r["next_position"], [.01, 0])
        self.check("sequence_invalid")

    def test_estop_latched(self):
        self.gate.stop()
        self.check("estop")

    def test_network_loss(self):
        self.check("heartbeat_lost", network_ok=False)

    def test_watchdog_loss(self):
        self.check("heartbeat_lost", watchdog_ok=False)

    def test_revoked(self):
        self.check("capability_denied", grant=Grant("virtual-robot-01", "sim_motion", 2000, True))

    def test_expired(self):
        self.check("capability_denied", now_ms=2000)

    def test_wrong_device(self):
        self.check("capability_denied", grant=Grant("another-device", "sim_motion", 2000))

    def test_wrong_operation(self):
        self.check("capability_denied", grant=Grant("virtual-robot-01", "real_motion", 2000))

    def test_stale_observation(self):
        self.check("stale_observation", now_ms=1201)

    def test_future_observation(self):
        self.check("stale_observation", obs=Observation(1, 1001, 0, 0, 2))

    def test_replayed_or_mismatched_sequence(self):
        self.check("sequence_invalid", proposal=Proposal(2, .1, 0, 100))

    def test_speed_limit(self):
        self.check("speed_limit", proposal=Proposal(1, .31, 0, 100))

    def test_duration_limit(self):
        self.check("duration_limit", proposal=Proposal(1, .1, 0, 501))

    def test_collision(self):
        self.check("collision_clearance", obs=Observation(1, 1000, 0, 0, .51),
                   proposal=Proposal(1, .2, 0, 100))

    def test_geofence(self):
        self.check("geofence", obs=Observation(1, 1000, 1.99, 0, 2),
                   proposal=Proposal(1, .2, 0, 100))

    def test_nonfinite(self):
        self.check("nonfinite_input", proposal=Proposal(1, float("nan"), 0, 100))

    def test_malformed(self):
        self.check("malformed_input", proposal=Proposal(1, .1, 0, True))

    def test_invalid_clock(self):
        self.check("invalid_clock", now_ms=True)


if __name__ == "__main__":
    unittest.main()
