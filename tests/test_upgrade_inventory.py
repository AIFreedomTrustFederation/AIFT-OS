"""Inventory regressions use disposable fixtures, never live services."""
import importlib.util
from pathlib import Path
import subprocess
import tempfile
import unittest

SPEC = importlib.util.spec_from_file_location("inventory", Path(__file__).resolve().parents[1] / "scripts/upgrade_inventory.py")
inventory = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(inventory)


class InventoryTests(unittest.TestCase):
    def test_missing_directory_is_not_ready(self):
        with tempfile.TemporaryDirectory() as tmp:
            record = inventory.inspect(Path(tmp) / "missing")
        self.assertEqual(record["error"], "missing_directory")
        self.assertEqual(record["verification"], "not_run")

    def test_commands_are_never_executed_and_files_unchanged(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            marker = root / "executed"
            content = '{"commands":{"verify":"touch ' + str(marker) + '"}}'
            (root / "aift.repo.json").write_text(content)
            record = inventory.inspect(root)
            self.assertFalse(marker.exists())
            self.assertEqual((root / "aift.repo.json").read_text(), content)
            self.assertEqual(record["files"]["README.md"]["state"], "missing")
            self.assertEqual(record["git_evidence"], "unavailable")

    def test_invalid_json_and_non_object(self):
        for content in ('{', '[]', 'null'):
            with self.subTest(content=content), tempfile.TemporaryDirectory() as tmp:
                (Path(tmp) / "package.json").write_text(content)
                self.assertEqual(inventory.inspect(tmp)["files"]["package.json"]["state"], "invalid_json_object")

    def test_symlink_is_refused(self):
        with tempfile.TemporaryDirectory() as tmp:
            (Path(tmp) / "README.md").symlink_to("/etc/passwd")
            self.assertEqual(inventory.inspect(tmp)["files"]["README.md"]["state"], "refused_symlink")

    def test_revision_and_dirty_evidence(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            def git(*args):
                return subprocess.check_output(["git", "-C", tmp, *args], text=True).strip()
            git("init", "-q")
            git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "--allow-empty", "-qm", "fixture")
            record = inventory.inspect(root)
            self.assertEqual(record["head"], git("rev-parse", "HEAD"))
            self.assertEqual(record["worktree"], "clean")
            (root / "README.md").write_text("changed")
            self.assertEqual(inventory.inspect(root)["worktree"], "dirty")
            child = root / "child"
            child.mkdir()
            self.assertEqual(inventory.inspect(child)["git_evidence"], "unavailable")
