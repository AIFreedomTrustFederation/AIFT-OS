"""Offline negative fixtures; no repository command is ever invoked."""
import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

SCRIPT = Path(__file__).resolve().parents[1] / 'scripts/analyze_command_contracts.py'
SPEC = importlib.util.spec_from_file_location('contracts', SCRIPT)
contracts = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(contracts)


class CommandContractsTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)

    def inspect(self, command, package=None):
        (self.root / 'aift.repo.json').write_text(json.dumps({'commands': {'verify': command}}))
        if package is not None:
            (self.root / 'package.json').write_text(json.dumps(package))
        return contracts.analyze(self.root)

    def codes(self, report):
        return {f['code'] for c in report['commands'] for f in c['findings']}

    def test_commands_not_executed_and_inputs_unchanged(self):
        marker = self.root / 'executed'
        command = 'touch ' + str(marker) + '; echo success'
        self.inspect(command)
        before = (self.root / 'aift.repo.json').read_bytes()
        report = contracts.analyze(self.root)
        self.assertFalse(marker.exists())
        self.assertEqual(before, (self.root / 'aift.repo.json').read_bytes())
        self.assertEqual(report['verification'], 'not_run')

    def test_optional_missing_script_is_reported(self):
        codes = self.codes(self.inspect('npm run test --if-present', {'scripts': {}}))
        self.assertTrue({'optional_gate', 'missing_npm_script'} <= codes)

    def test_present_no_op_script_is_not_verification(self):
        self.assertIn('no_op_npm_script', self.codes(self.inspect('npm run test', {'scripts': {'test': 'echo passed'}})))

    def test_git_status_is_diagnostic(self):
        self.assertIn('diagnostic_only_gate', self.codes(self.inspect('git status --short')))

    def test_dynamic_machine_specific_and_masked_command(self):
        codes = self.codes(self.inspect('bash -c \'source "$HOME/AIFT/runtime/common.sh" && verify\' || true'))
        self.assertTrue({'dynamic_shell', 'machine_specific_path', 'failure_masking'} <= codes)

    def test_missing_and_symlink_entrypoints(self):
        self.assertIn('missing_entrypoint', self.codes(self.inspect('./verify.sh')))
        (self.root / 'verify.sh').symlink_to('/etc/passwd')
        self.assertIn('refused_entrypoint', self.codes(self.inspect('./verify.sh')))

    def test_parent_directory_symlink_refused(self):
        (self.root / '.aift').symlink_to('/tmp')
        self.assertIn('refused_entrypoint', self.codes(self.inspect('.aift/verify.sh')))

    def test_malformed_and_nonstandard_metadata_unverifiable(self):
        for value in ('[]', '{', '{"commands":NaN}', '{"commands":{}}'):
            with self.subTest(value=value):
                (self.root / 'aift.repo.json').write_text(value)
                self.assertEqual(contracts.analyze(self.root)['analysis'], 'unverifiable')

    def test_invalid_package_scripts_reported(self):
        self.assertIn('package_unavailable', self.codes(self.inspect('npm run test', {'scripts': []})))

    def test_no_known_issue_does_not_become_pass(self):
        report = self.inspect('python3 -m unittest discover')
        self.assertIn('manual_review', self.codes(report))
        self.assertEqual(report['verification'], 'not_run')

    def test_mutation_is_flagged(self):
        self.assertIn('possible_mutation', self.codes(self.inspect('git pull --ff-only')))

    def test_cli_exit_and_repeatability(self):
        self.inspect('git status --short')
        cmd = [sys.executable, '-B', str(SCRIPT), str(self.root)]
        first = subprocess.run(cmd, capture_output=True, text=True)
        second = subprocess.run(cmd, capture_output=True, text=True)
        self.assertEqual(first.returncode, 0)
        self.assertEqual(first.stdout, second.stdout)
        missing = subprocess.run(cmd + [str(self.root / 'absent')], capture_output=True, text=True)
        self.assertEqual(missing.returncode, 1)
        self.assertEqual(len(json.loads(missing.stdout)['repositories']), 2)


if __name__ == '__main__':
    unittest.main()
