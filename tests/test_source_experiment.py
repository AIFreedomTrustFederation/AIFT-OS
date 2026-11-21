import importlib.util
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

SPEC = importlib.util.spec_from_file_location('runner', Path(__file__).resolve().parents[1] / 'scripts/run_source_experiment.py')
runner = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(runner)


class ExperimentTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.git('init', '-q')
        (self.root / 'input.txt').write_text('original')
        self.git('add', '.')
        self.git('-c', 'user.name=Fixture', '-c', 'user.email=fixture@example.invalid', 'commit', '-qm', 'fixture')
        self.sha = self.git('rev-parse', 'HEAD').strip()

    def git(self, *args):
        return subprocess.check_output(['git', '-C', str(self.root), *args], text=True)

    def invoke(self, code, **kwargs):
        return runner.run(self.root, self.sha, [sys.executable, '-c', code], **kwargs)

    def test_success_and_isolated_writes(self):
        result = self.invoke("from pathlib import Path; Path('input.txt').write_text('changed'); print('ok')")
        self.assertEqual(result['status'], 'success')
        self.assertEqual((self.root / 'input.txt').read_text(), 'original')
        self.assertEqual(self.git('status', '--porcelain'), '')

    def test_failure_preserves_exit(self):
        result = self.invoke('raise SystemExit(7)')
        self.assertEqual(result['status'], 'failed')
        self.assertEqual(result['exit_code'], 7)

    def test_timeout(self):
        self.assertEqual(self.invoke('import time; time.sleep(5)', timeout=.05)['status'], 'timed_out')

    def test_output_bounded(self):
        result = self.invoke("print('x'*100000)", output_limit=100)
        self.assertEqual(result['status'], 'output_limit')
        self.assertLessEqual(len(result['stdout']), 100)

    def test_dirty_and_stale_inputs_block_before_execution(self):
        self.assertEqual(runner.run(self.root, 'f'*40, [sys.executable, '-c', 'print(1)'])['status'], 'blocked')
        (self.root / 'dirty').touch()
        self.assertEqual(self.invoke('print(1)')['status'], 'blocked')

    def test_symlink_snapshot_refused(self):
        (self.root / 'link').symlink_to('/etc/passwd')
        self.git('add', '.')
        self.git('-c', 'user.name=Fixture', '-c', 'user.email=fixture@example.invalid', 'commit', '-qm', 'symlink')
        self.sha = self.git('rev-parse', 'HEAD').strip()
        self.assertEqual(self.invoke('print(1)')['status'], 'blocked')

    def test_environment_credentials_not_inherited(self):
        import os
        from unittest.mock import patch
        with patch.dict(os.environ, {'EXPERIMENT_SECRET': 'should-not-inherit'}):
            result = self.invoke("import os; assert 'EXPERIMENT_SECRET' not in os.environ")
        self.assertEqual(result['status'], 'success')

    def test_missing_executable_blocks(self):
        self.assertEqual(runner.run(self.root, self.sha, ['/nonexistent/command'])['status'], 'blocked')

    def test_concurrent_input_mutation_invalidates_result(self):
        code = "from pathlib import Path; Path(%r).write_text('changed')" % str(self.root / 'input.txt')
        result = self.invoke(code)
        self.assertEqual(result['status'], 'blocked')
        self.assertEqual(result['error'], 'Input checkout changed during experiment')

    def test_invalid_limits_block(self):
        self.assertEqual(self.invoke('print(1)', timeout=0)['status'], 'blocked')
        self.assertEqual(self.invoke('print(1)', output_limit=0)['status'], 'blocked')


if __name__ == '__main__':
    unittest.main()
