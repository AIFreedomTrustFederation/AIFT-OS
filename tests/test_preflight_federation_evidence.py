import importlib.util
from pathlib import Path
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location(
    "preflight", Path(__file__).parents[1] / "scripts/preflight_federation_evidence.py")
preflight = importlib.util.module_from_spec(spec)
spec.loader.exec_module(preflight)

class PreflightTests(unittest.TestCase):
    def test_all_prerequisites_present(self):
        def fake(path):
            if "/variables/" in path:
                return {"name": preflight.VARIABLE, "value": "Iv1_example"}
            if "/secrets/" in path:
                return {"name": preflight.SECRET}
            return {"full_name": path.removeprefix("repos/")}
        self.assertTrue(all(preflight.diagnose(fake).values()))

    def test_missing_secret_fails_closed(self):
        def fake(path):
            if "/secrets/" in path:
                return None
            if "/variables/" in path:
                return {"name": preflight.VARIABLE, "value": "Iv1_example"}
            return {"full_name": path.removeprefix("repos/")}
        self.assertFalse(preflight.diagnose(fake)["private_key_secret_exists"])

    def test_missing_variable_and_wrong_producer_identity_fail_closed(self):
        def fake(path):
            if "/variables/" in path:
                return {"name": preflight.VARIABLE, "value": ""}
            if "/secrets/" in path:
                return {"name": preflight.SECRET}
            return {"full_name": "other/repo"}
        result = preflight.diagnose(fake)
        self.assertFalse(result["client_id_variable"])
        self.assertFalse(result["operator_can_see_.github"])
        self.assertFalse(result["operator_can_see_AIFT-Runtime"])

    def test_request_never_passes_secret_or_mutation_arguments(self):
        with patch.object(preflight.subprocess, "run") as run:
            run.return_value.returncode = 0
            run.return_value.stdout = b'{"name":"fixture"}'
            preflight.request("repos/example/actions/secrets/name")
            argv = run.call_args.args[0]
            self.assertEqual(argv[:3], ["gh", "api", "--hostname"])
            self.assertNotIn("--method", argv)
            self.assertNotIn("--input", argv)

if __name__ == "__main__":
    unittest.main()
