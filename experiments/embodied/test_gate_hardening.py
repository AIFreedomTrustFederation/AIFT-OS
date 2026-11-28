"""Fault injection against public simulation entry points. No external I/O."""
from dataclasses import FrozenInstanceError, replace
import json
import math
import sys
import unittest

from ros_contract_adapter import evaluate_envelope, evaluate_json_envelope
from safety_sim import Grant, MAX_COUNTER, Observation, Proposal, SafetyGate
from test_inmemory_transport import InMemoryTransport, encoded, frame
from test_mock_pipeline import simulated_cycle


class GateHardeningTests(unittest.TestCase):
    def setUp(self):
        self.gate = SafetyGate()
        self.obs = Observation(1, 1000, 0, 0, 2)
        self.proposal = Proposal(1, 0.1, 0, 100)
        self.grant = Grant("virtual-robot-01", "sim_motion", 2000)

    def evaluate(self, *, gate=None, obs=None, proposal=None, grant=None, now_ms=1000, **health):
        return (gate or self.gate).evaluate(obs or self.obs, proposal or self.proposal,
                                            grant or self.grant, now_ms=now_ms, **health)

    def assert_stop(self, result, reason=None):
        self.assertIs(result["accepted"], False)
        self.assertEqual(result["action"], "STOP")
        self.assertIs(result["simulated"], True)
        self.assertNotIn("next_position", result)
        if reason is not None:
            self.assertEqual(result["reason"], reason)

    def test_oversized_position(self):
        for axis in ("x", "y", "obstacle_m"):
            for value in (10**400, -(10**400)):
                with self.subTest(axis=axis, value=value):
                    self.assert_stop(self.evaluate(obs=replace(self.obs, **{axis: value})))
        self.assertTrue(self.evaluate()["accepted"])

    def test_oversized_velocity(self):
        for axis in ("vx_mps", "vy_mps"):
            for value in (10**400, -(10**400)):
                with self.subTest(axis=axis, value=value):
                    self.assert_stop(self.evaluate(proposal=replace(self.proposal, **{axis: value})))
        self.assertTrue(self.evaluate()["accepted"])

    def test_nan_speed_limit(self):
        with self.assertRaises(ValueError):
            SafetyGate(max_speed=math.nan)

    def test_nan_geofence(self):
        with self.assertRaises(ValueError):
            SafetyGate(bound_m=math.nan)

    def test_nan_clearance(self):
        with self.assertRaises(ValueError):
            SafetyGate(clearance_m=math.nan)

    def check_health_latch(self, field):
        self.assert_stop(self.evaluate(**{field: False}), "heartbeat_lost")
        self.assert_stop(self.evaluate(), "heartbeat_lost")
        self.assert_stop(self.evaluate(obs=replace(self.obs, sequence=2),
                                      proposal=replace(self.proposal, sequence=2)), "heartbeat_lost")
        self.assertTrue(self.evaluate(gate=SafetyGate())["accepted"])
        self.assert_stop(self.evaluate(), "heartbeat_lost")

    def test_network_failure_is_terminal(self):
        self.check_health_latch("network_ok")

    def test_watchdog_failure_is_terminal(self):
        self.check_health_latch("watchdog_ok")

    def test_estop_precedence_and_no_reset(self):
        self.assert_stop(self.evaluate(watchdog_ok=None), "heartbeat_lost")
        self.gate.stop()
        self.assert_stop(self.evaluate(), "estop")
        with self.assertRaises(AttributeError):
            self.gate.estop = False

    def test_configuration_domain(self):
        for field in ("max_speed", "clearance_m", "bound_m"):
            for bad in (True, None, "1", -1, 0, 10**400, math.inf, -math.inf):
                with self.subTest(field=field, value=bad), self.assertRaises(ValueError):
                    SafetyGate(**{field: bad})
        for field in ("max_age_ms", "max_duration_ms"):
            for bad in (True, 1.0, -1, None, "1", math.nan, math.inf, MAX_COUNTER + 1):
                with self.subTest(field=field, value=bad), self.assertRaises(ValueError):
                    SafetyGate(**{field: bad})
        for device in (None, True, "", " "):
            with self.assertRaises(ValueError):
                SafetyGate(device=device)
        with self.assertRaises(ValueError):
            SafetyGate(max_duration_ms=0)
        with self.assertRaises(FrozenInstanceError):
            self.gate.bound_m = math.nan

    def test_inward_step_cannot_hide_invalid_start(self):
        for axis, velocity in (("x", "vx_mps"), ("y", "vy_mps")):
            for sign in (-1, 1):
                self.assert_stop(self.evaluate(obs=replace(self.obs, **{axis: 2.01 * sign}),
                                               proposal=replace(self.proposal, **{velocity: -.3 * sign})),
                                 "geofence")

    def test_inclusive_geofence_and_unrounded_endpoint(self):
        for sign in (-1, 1):
            gate = SafetyGate(bound_m=0.9999999)
            obs = replace(self.obs, x=sign * gate.bound_m)
            stopped = replace(self.proposal, vx_mps=0.0)
            self.assertEqual(self.evaluate(gate=gate, obs=obs, proposal=stopped)["next_position"],
                             [obs.x, 0.0])
            self.assert_stop(self.evaluate(gate=SafetyGate(bound_m=gate.bound_m), obs=obs,
                                           proposal=replace(stopped, vx_mps=sign * 1e-16)), "geofence")

    def test_clearance_speed_duration_and_age_edges(self):
        obs = replace(self.obs, obstacle_m=.625, timestamp_ms=800)
        proposal = replace(self.proposal, vx_mps=.25, duration_ms=500)
        for bad_obs in (replace(obs, obstacle_m=math.nextafter(.625, 0)),
                        replace(obs, timestamp_ms=799)):
            self.assert_stop(self.evaluate(gate=SafetyGate(max_speed=.25), obs=bad_obs, proposal=proposal))
        for bad_proposal in (replace(proposal, vx_mps=math.nextafter(.25, 1)),
                             replace(proposal, duration_ms=501)):
            self.assert_stop(self.evaluate(gate=SafetyGate(max_speed=.25), obs=obs, proposal=bad_proposal))
        self.assertTrue(self.evaluate(gate=SafetyGate(max_speed=.25), obs=obs, proposal=proposal)["accepted"])

    def test_counter_and_arithmetic_extremes(self):
        for value in (-1, MAX_COUNTER + 1, True, 1.0):
            self.assert_stop(self.evaluate(obs=replace(self.obs, sequence=value)))
            self.assert_stop(self.evaluate(proposal=replace(self.proposal, duration_ms=value)))
            self.assert_stop(self.evaluate(grant=replace(self.grant, expires_ms=value)))
            self.assert_stop(self.evaluate(gate=SafetyGate(), now_ms=value), "invalid_clock")
        huge = sys.float_info.max
        gate = SafetyGate(max_speed=huge, bound_m=huge, max_duration_ms=MAX_COUNTER)
        self.assert_stop(self.evaluate(gate=gate, obs=replace(self.obs, obstacle_m=huge),
                                       proposal=Proposal(1, huge, 0, MAX_COUNTER)))
        self.assert_stop(self.evaluate(gate=gate, proposal=Proposal(1, huge, huge, 1)))
        self.assertTrue(self.evaluate(gate=SafetyGate(max_age_ms=0),
                                      obs=replace(self.obs, sequence=MAX_COUNTER),
                                      proposal=replace(self.proposal, sequence=MAX_COUNTER))["accepted"])

    def test_rollback_after_denial_cannot_resurrect_grant(self):
        self.assert_stop(self.evaluate(now_ms=2000), "capability_denied")
        self.assert_stop(self.evaluate(now_ms=1000), "clock_rollback")
        self.assert_stop(self.evaluate(now_ms=2001), "clock_rollback")

    def test_same_receiver_tick_allows_distinct_sequences(self):
        self.assertTrue(self.evaluate()["accepted"])
        self.assertTrue(self.evaluate(obs=replace(self.obs, sequence=2),
                                      proposal=replace(self.proposal, sequence=2))["accepted"])

    def test_receiver_time_is_required_for_both_adapters(self):
        message = frame(1, 1000)
        message["grant"]["expires_ms"] = 1050
        for adapter, data in ((evaluate_envelope, message), (evaluate_json_envelope, encoded(message))):
            self.assert_stop(adapter(SafetyGate(), data))
            self.assert_stop(adapter(SafetyGate(), data, now_ms=1100))
            self.assert_stop(adapter(SafetyGate(), data, now_ms=True))

    def test_malformed_bytes_still_advance_receiver_clock(self):
        self.assert_stop(evaluate_json_envelope(self.gate, b"[]", now_ms=1001))
        self.assert_stop(evaluate_json_envelope(self.gate, b"{", now_ms=1002))
        self.assert_stop(evaluate_envelope(self.gate, frame(1, 1001), now_ms=1001), "clock_rollback")

    def test_heartbeat_loss_with_malformed_proposal_latches(self):
        message = frame(1, 1000)
        message.update(network_ok=False, proposal=None)
        self.assert_stop(evaluate_json_envelope(self.gate, encoded(message), now_ms=1000))
        self.assert_stop(self.evaluate(), "heartbeat_lost")

    def test_json_exponent_overflow(self):
        payload = json.dumps(frame(1, 1000)).replace('"vx_mps": 0.1', '"vx_mps": 1e400').encode()
        self.assert_stop(evaluate_json_envelope(self.gate, payload, now_ms=1000))

    def test_transport_numeric_domains_and_terminal_stop(self):
        for value in (0, -1, True, 1.0, math.nan, math.inf, MAX_COUNTER + 1):
            with self.assertRaises(ValueError):
                InMemoryTransport(deadline_ms=value)
        for method in ("arm", "tick", "deliver"):
            transport = InMemoryTransport()
            if method != "arm":
                transport.arm(1000)
            if method == "deliver":
                result = transport.deliver(encoded(frame(1, 1000)), delivery_ms=MAX_COUNTER + 1, sequence=1)
            else:
                result = getattr(transport, method)(MAX_COUNTER + 1)
            self.assert_stop(result)
            self.assert_stop(transport.arm(1000))

    def test_mock_fixture_denies_invalid_inputs_before_arithmetic(self):
        for kwargs in ({"now_ms": None}, {"observation_age_ms": "1"}, {"consent": "yes"}, {"estop": 0}):
            self.assertFalse(simulated_cycle(**kwargs)["accepted"])
