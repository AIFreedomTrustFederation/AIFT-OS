"""Offline contracts for live Aetherion readiness (no model, GitHub, or secrets)."""
import importlib.util
from pathlib import Path
import unittest

SCRIPT = Path(__file__).resolve().parents[1] / "scripts" / "aift_aetherion_readiness.py"
spec = importlib.util.spec_from_file_location("aift_aetherion_readiness", SCRIPT)
readiness = importlib.util.module_from_spec(spec)
spec.loader.exec_module(readiness)


class AetherionReadinessTests(unittest.TestCase):
    def test_five_exact_document_edits_succeed_without_publishing(self):
        calls = []

        def fake_infer(**kwargs):
            calls.append(kwargs)
            return {"content": readiness.EXPECTED, "summary": "single edit"}

        ticks = iter(range(11))
        report, code = readiness.evaluate(
            repetitions=5, model="aetherion:latest", endpoint="http://127.0.0.1:11434/v1",
            infer=fake_infer, clock=lambda: next(ticks),
        )
        self.assertEqual(code, 0)
        self.assertEqual(report["status"], "local_worker_inference_verified")
        self.assertEqual((report["attempted"], report["passed"], report["failed"]), (5, 5, 0))
        self.assertEqual(len(calls), 5)
        self.assertTrue(all(c["content"] == readiness.SOURCE for c in calls))
        self.assertEqual(report["publisher"], "not_checked")
        self.assertEqual(report["runner_registration"], "not_checked")
        self.assertEqual(report["production_deployment"], "not_checked")

    def test_wrong_text_fails_closed_even_with_valid_shape(self):
        report, code = readiness.evaluate(
            repetitions=1, model="aetherion:latest", endpoint="http://127.0.0.1:11434/v1",
            infer=lambda **_: {"content": readiness.EXPECTED + "extra"}, clock=iter([0, 1]).__next__,
        )
        self.assertEqual(code, 1)
        self.assertEqual(report["checks"]["local_worker_inference"], "failed")
        self.assertEqual(report["passed"], 0)

    def test_model_exception_is_not_leaked_into_evidence(self):
        def failing(**_):
            raise RuntimeError("sensitive-prompt-or-token")

        report, code = readiness.evaluate(
            repetitions=1, model="aetherion:latest", endpoint="http://127.0.0.1:11434/v1",
            infer=failing, clock=iter([0, 1]).__next__,
        )
        self.assertEqual(code, 1)
        self.assertNotIn("sensitive-prompt-or-token", str(report))

    def test_exact_head_mismatch_prevents_inference(self):
        def must_not_run(**_):
            self.fail("inference must not execute when SHA is wrong")

        report, code = readiness.evaluate(
            repetitions=5, model="aetherion:latest", endpoint="http://127.0.0.1:11434/v1",
            expected_sha="a" * 40, resolve_sha=lambda: "b" * 40, infer=must_not_run,
        )
        self.assertEqual(code, 1)
        self.assertEqual(report["attempted"], 0)
        self.assertEqual(report["checks"]["exact_head"], "mismatch")

    def test_matching_head_allows_smoke(self):
        report, code = readiness.evaluate(
            repetitions=1, model="aetherion:latest", endpoint="http://127.0.0.1:11434/v1",
            expected_sha="a" * 40, resolve_sha=lambda: "a" * 40,
            infer=lambda **_: {"content": readiness.EXPECTED},
            clock=iter([0, 1]).__next__,
        )
        self.assertEqual(code, 0)
        self.assertEqual(report["checks"]["exact_head"], "passed")
        self.assertEqual(report["source_sha"], "a" * 40)

    def test_repetition_limit(self):
        for count in (0, 11, -1):
            with self.subTest(count=count):
                with self.assertRaises(ValueError):
                    readiness.evaluate(repetitions=count, model="x", endpoint="y")

    def test_workflow_manual_only_on_trusted_main(self):
        workflow = (SCRIPT.parents[1] / ".github/workflows/aift-aetherion-readiness.yml").read_text()
        self.assertIn("github.event_name == 'workflow_dispatch'", workflow)
        self.assertIn("github.ref == 'refs/heads/main'", workflow)
        self.assertIn("aift-local-model", workflow)
        self.assertIn("persist-credentials: false", workflow)
        self.assertNotIn("APP_PRIVATE_KEY", workflow)
        self.assertNotIn("AIFT_AUTODEV_AI_ENABLED=true", workflow)
        self.assertNotIn("pull_request_target:", workflow)

    def test_existing_planner_is_guarded_by_live_readiness(self):
        workflow = (SCRIPT.parents[1] / ".github/workflows/aift-secure-cloud-worker.yml").read_text()
        self.assertIn("Gate planning on exact-head local Aetherion readiness", workflow)
        self.assertIn("scripts/aift_aetherion_readiness.py", workflow)
        self.assertIn("--expected-sha", workflow)
        self.assertIn("aift-secure-plan.json", workflow)


if __name__ == "__main__":
    unittest.main()
