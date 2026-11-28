"""E1 mock-only sensor -> planner -> deterministic gate -> simulated action tests.

No ROS transport, model, network, sensors or physical actuator connections.
Run: python3 -m unittest discover -s experiments/embodied -p 'test_*.py' -v
"""
import unittest
from safety_sim import SafetyGate, Observation, Proposal, Grant, valid_counter


def simulated_cycle(*, sequence=1, now_ms=1000, obstacle_m=2.0,
                    requested_speed=0.1, consent=True, network_ok=True,
                    watchdog_ok=True, estop=False, observation_age_ms=0):
    """Untrusted mock planner may request movement; gate alone authorizes a label."""
    timestamp = None
    if valid_counter(now_ms) and valid_counter(observation_age_ms):
        timestamp = now_ms - observation_age_ms
    observation = Observation(sequence, timestamp, 0.0, 0.0, obstacle_m)
    proposal = Proposal(sequence, requested_speed, 0.0, 100)
    grant = Grant("virtual-robot-01", "sim_motion", now_ms + 1000 if valid_counter(now_ms) else None,
                  revoked=consent is not True)
    gate = SafetyGate()
    if estop is not False:
        gate.stop()
    result = gate.evaluate(observation, proposal, grant, now_ms=now_ms,
                           network_ok=network_ok, watchdog_ok=watchdog_ok)
    return {"sensor": "MOCK", "planner": "UNTRUSTED_MOCK",
            "action": result["action"], "accepted": result["accepted"],
            "reason": result["reason"], "physical_actuation": False}


class MockPipelineTests(unittest.TestCase):
    def assert_stopped(self, **kwargs):
        result = simulated_cycle(**kwargs)
        self.assertEqual(result["action"], "STOP")
        self.assertFalse(result["accepted"])
        self.assertFalse(result["physical_actuation"])

    def test_nominal_path_is_simulation_only(self):
        result = simulated_cycle()
        self.assertTrue(result["accepted"])
        self.assertEqual(result["action"], "SIMULATE_ONLY")
        self.assertFalse(result["physical_actuation"])

    def test_untrusted_planner_overspeed(self):
        self.assert_stopped(requested_speed=1.0)

    def test_sensor_reports_insufficient_clearance(self):
        self.assert_stopped(obstacle_m=0.1)

    def test_stale_sensor_fails_closed(self):
        self.assert_stopped(observation_age_ms=201)

    def test_consent_revocation_fails_closed(self):
        self.assert_stopped(consent=False)

    def test_network_failure_fails_closed(self):
        self.assert_stopped(network_ok=False)

    def test_watchdog_failure_fails_closed(self):
        self.assert_stopped(watchdog_ok=False)

    def test_estop_fails_closed(self):
        self.assert_stopped(estop=True)

    def test_malformed_mock_planner_output_fails_closed(self):
        self.assert_stopped(requested_speed="fast")


if __name__ == "__main__":
    unittest.main()
