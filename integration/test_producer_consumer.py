"""Real exporter/consumer subprocesses with disposable QA fixtures, no mocks."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/verify_producer_integration.py"


class ProducerConsumerTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.env = {k: v for k, v in os.environ.items() if not k.startswith("GITHUB_")}
        self.roots = []
        self.shas = []
        for name, variable, exporter in (
            ("runtime", "AIFT_TEST_RUNTIME_ROOT", "run_runtime_verification.py"),
            ("forge", "AIFT_TEST_FORGE_ROOT", "run_forge_verification.py"),
        ):
            source = Path(os.environ[variable]) / "scripts" / exporter
            root = self.root / name
            (root / "scripts").mkdir(parents=True)
            shutil.copyfile(source, root / "scripts" / exporter)
            (root / "qa-local.sh").write_text("#!/bin/sh\nexit 0\n")
            (root / "package.json").write_text(json.dumps({"private": True, "scripts": {
                "qa:local": 'node -e "process.exit(0)"'}}))
            self.git(root, "init", "-q")
            self.commit(root)
            self.roots.append(root)
            self.shas.append(self.git(root, "rev-parse", "HEAD"))

    def git(self, root, *args):
        return subprocess.check_output(["git", "-C", str(root), *args], env=self.env,
                                       text=True, stderr=subprocess.PIPE).strip()

    def commit(self, root):
        self.git(root, "add", ".")
        self.git(root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid",
                 "commit", "-qm", "fixture")

    def invoke(self):
        return subprocess.run([sys.executable, "-B", str(SCRIPT),
            "--runtime-root", str(self.roots[0]), "--runtime-sha", self.shas[0],
            "--forge-root", str(self.roots[1]), "--forge-sha", self.shas[1]],
            env=self.env, capture_output=True, text=True, timeout=60)

    def test_real_exporters_preserve_partial_unsigned_status(self):
        result = self.invoke()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertEqual(report["status"], "partial")
        self.assertEqual(report["provenance"], "unsigned_supplied_json_not_authenticated")
        self.assertEqual(set(report["producer_exit_codes"].values()), {0})
        for item in report["repositories"]:
            self.assertEqual(item["reported_source"]["type"], "local")
            self.assertEqual(len(item["unverified_gates"]), 3)
        for root in self.roots:
            self.assertEqual(self.git(root, "status", "--porcelain"), "")

    def test_either_real_producer_failure_survives_consumption(self):
        for i in range(2):
            with self.subTest(producer=i):
                root = self.roots[i]
                if i == 0:
                    (root / "qa-local.sh").write_text("#!/bin/sh\nexit 7\n")
                else:
                    (root / "package.json").write_text(json.dumps({"scripts": {
                        "qa:local": 'node -e "process.exit(7)"'}}))
                self.commit(root)
                self.shas[i] = self.git(root, "rev-parse", "HEAD")
                result = self.invoke()
                self.assertEqual(result.returncode, 1, result.stderr)
                report = json.loads(result.stdout)
                self.assertEqual(report["status"], "failed")
                producer = ("AIFT-Runtime", "AIFT-Forge")[i]
                other = ("AIFT-Forge", "AIFT-Runtime")[i]
                self.assertNotEqual(report["producer_exit_codes"][producer], 0)
                self.assertEqual(report["producer_exit_codes"][other], 0)
                (root / "qa-local.sh").write_text("#!/bin/sh\nexit 0\n")
                (root / "package.json").write_text(json.dumps({"private": True, "scripts": {
                    "qa:local": 'node -e "process.exit(0)"'}}))
                self.commit(root)
                self.shas[i] = self.git(root, "rev-parse", "HEAD")

    def test_stale_sha_blocks_before_either_qa_runs(self):
        self.shas[1] = "f" * 40
        result = self.invoke()
        self.assertEqual(result.returncode, 2)
        self.assertEqual(result.stdout, "")
        self.assertNotIn("Verification evidence:", result.stderr)

    def test_dirty_root_blocks_before_either_qa_runs(self):
        (self.roots[1] / "untracked").write_text("dirty")
        result = self.invoke()
        self.assertEqual(result.returncode, 2)
        self.assertIn("must be clean", result.stderr)
        self.assertEqual(result.stdout, "")

    def run_with_forge_qa(self, command):
        forge = self.roots[1]
        (forge / "package.json").write_text(json.dumps({"scripts": {"qa:local": command}}))
        self.commit(forge)
        self.shas[1] = self.git(forge, "rev-parse", "HEAD")
        return self.invoke()

    def test_runtime_dirtied_after_its_qa_blocks_report(self):
        result = self.run_with_forge_qa("touch ../runtime/late-file")
        self.assertTrue((self.roots[0] / "late-file").is_file())
        self.assertEqual(result.returncode, 2, result.stderr)
        self.assertIn("input checkout must be clean", result.stderr)
        self.assertEqual(result.stdout, "")

    def test_runtime_commit_changed_after_its_qa_blocks_report(self):
        result = self.run_with_forge_qa(
            "git -C ../runtime -c user.name=Fixture -c user.email=fixture@example.invalid "
            "commit --allow-empty -qm late-revision"
        )
        self.assertNotEqual(self.git(self.roots[0], "rev-parse", "HEAD"), self.shas[0])
        self.assertEqual(self.git(self.roots[0], "status", "--porcelain"), "")
        self.assertEqual(result.returncode, 2, result.stderr)
        self.assertIn("input checkout does not match expected commit", result.stderr)
        self.assertEqual(result.stdout, "")

    def test_missing_exporter_blocks(self):
        (self.roots[1] / "scripts/run_forge_verification.py").unlink()
        self.commit(self.roots[1])
        self.shas[1] = self.git(self.roots[1], "rev-parse", "HEAD")
        result = self.invoke()
        self.assertEqual(result.returncode, 2)
        self.assertIn("entrypoint missing", result.stderr)


if __name__ == "__main__":
    unittest.main()
