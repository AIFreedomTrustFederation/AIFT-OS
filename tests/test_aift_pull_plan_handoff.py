"""Offline checks for operator-review digest binding and no-overwrite handoff."""
import base64
import hashlib
import importlib.util
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest

ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location(
    "aift_pull_plan_handoff", ROOT / "scripts/aift_pull_plan_handoff.py"
)
handoff = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(handoff)


class HandoffTests(unittest.TestCase):
    def test_verified_plan_bytes_are_preserved_and_private(self):
        plan = b'{"schema_version":2,"old":"before","new":"after"}\n'
        encoded = base64.b64encode(plan).decode()
        digest = hashlib.sha256(plan).hexdigest()
        with TemporaryDirectory() as tmp:
            destination = Path(tmp) / "plan.json"
            self.assertEqual(handoff.decode_reviewed_plan(encoded, digest, destination), digest)
            self.assertEqual(destination.read_bytes(), plan)
            self.assertEqual(destination.stat().st_mode & 0o777, 0o600)
            # Even if the same plan is replayed, it must not replace a file.
            with self.assertRaises(FileExistsError):
                handoff.decode_reviewed_plan(encoded, digest, destination)
            self.assertEqual(destination.read_bytes(), plan)

    def test_mismatched_digest_or_malformed_data_never_creates_file(self):
        plan = b'{"schema_version":2}'
        encoded = base64.b64encode(plan).decode()
        digest = hashlib.sha256(plan).hexdigest()
        examples = (
            (encoded, "0" * 64), (encoded, digest.upper()),
            (encoded, ""), (encoded, "x" * 64),
            ("not-base64!", digest), ("", digest),
            ("A" * 11001, digest),
            (base64.b64encode(b"x" * 8193).decode(), digest),
        )
        with TemporaryDirectory() as tmp:
            for i, (payload, supplied_digest) in enumerate(examples):
                destination = Path(tmp) / f"invalid-{i}.json"
                with self.subTest(i=i), self.assertRaises(handoff.InvalidHandoff):
                    handoff.decode_reviewed_plan(payload, supplied_digest, destination)
                self.assertFalse(destination.exists())

    def test_symlink_handoff_is_blocked(self):
        data = b'{"schema_version":2}'
        digest = hashlib.sha256(data).hexdigest()
        with TemporaryDirectory() as tmp:
            root = Path(tmp)
            target = root / "sensitive.txt"
            target.write_text("untouched", encoding="utf-8")
            link = root / "handoff.json"
            link.symlink_to(target)
            with self.assertRaises(OSError):
                handoff.decode_reviewed_plan(base64.b64encode(data).decode(), digest, link)
            self.assertEqual(target.read_text(encoding="utf-8"), "untouched")

    def test_hosted_workflow_requires_digest_before_write_token(self):
        source = (ROOT / ".github/workflows/aift-pull-doc-publisher.yml").read_text()
        self.assertIn("plan_sha256:", source)
        self.assertIn("required: true", source)
        self.assertIn("PLAN_SHA256:", source)
        self.assertIn("scripts/aift_pull_plan_handoff.py", source)
        self.assertLess(source.index("scripts/aift_pull_plan_handoff.py"),
                        source.index("scripts/aift_secure_cloud_worker.py preflight"))
        self.assertLess(source.index("scripts/aift_secure_cloud_worker.py preflight"),
                        source.index("actions/create-github-app-token@"))


if __name__ == "__main__":
    unittest.main()
