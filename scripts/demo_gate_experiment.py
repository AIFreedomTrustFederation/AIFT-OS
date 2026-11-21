#!/usr/bin/env python3
"""Disposable npm fixture comparing the reviewed Forge command contracts."""
import argparse
import json
from pathlib import Path
import shutil
import subprocess
import tempfile

from run_source_experiment import run

BASELINE = 'npm run lint --if-present && npm run typecheck --if-present && npm run test --if-present && npm run build --if-present'
CANDIDATE = 'npm run qa:local'


def demonstrate():
    npm = shutil.which('npm')
    node = shutil.which('node')
    if not npm or not node:
        raise ValueError('npm and node must already be installed')
    tools = sorted({str(Path(npm).parent), str(Path(node).parent)})
    records = []
    with tempfile.TemporaryDirectory(prefix='aift-gate-demo-') as tmp:
        for mode in ('healthy', 'missing_stage', 'failed_stage'):
            for variant, command in (('baseline', BASELINE), ('candidate', CANDIDATE)):
                root = Path(tmp) / (mode + '-' + variant)
                root.mkdir()
                scripts = {'qa:local': 'npm run lint && npm run typecheck && npm run test && npm run build'}
                for stage in ('lint', 'typecheck', 'test', 'build'):
                    if mode == 'missing_stage' and stage == 'typecheck':
                        continue
                    scripts[stage] = 'node -e "process.exit(%d)"' % (7 if mode == 'failed_stage' and stage == 'test' else 0)
                (root / 'package.json').write_text(json.dumps({'private': True, 'scripts': scripts}))
                (root / 'aift.repo.json').write_text(json.dumps({'commands': {'verify': command}}))
                def git(*args):
                    return subprocess.check_output(['git', '-C', str(root), *args], text=True).strip()
                git('init', '-q')
                git('add', '.')
                git('-c', 'user.name=Fixture', '-c', 'user.email=fixture@example.invalid', 'commit', '-qm', 'gate fixture')
                result = run(root, git('rev-parse', 'HEAD'), ['/bin/sh', '-c', command],
                             timeout=30, tool_paths=tools)
                result.update(fixture=mode, variant=variant)
                records.append(result)
    expected = [('healthy', 'baseline', 'success'), ('healthy', 'candidate', 'success'),
                ('missing_stage', 'baseline', 'success'), ('missing_stage', 'candidate', 'failed'),
                ('failed_stage', 'baseline', 'failed'), ('failed_stage', 'candidate', 'failed')]
    observed = [(r['fixture'], r['variant'], r['status']) for r in records]
    return {'schema': 'aift.gate-comparison.v1', 'scope': 'disposable_npm_fixture',
            'expectations_met': observed == expected, 'records': records,
            'upstream': {'repository': 'AIFreedomTrustFederation/AIFT-Forge',
                         'baseline_sha': '584781172a48c2d8ed1aaff70ae543a70f7822d6',
                         'candidate_sha': 'c98b856ca6046c953417064a0411861f8fd98e6c'},
            'limitations': ['Fixture commits are not upstream revisions.',
                            'Synthetic exit-only stages do not test real lint, compilation or builds.']}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.parse_args()
    try:
        report = demonstrate()
    except (OSError, ValueError, subprocess.SubprocessError) as error:
        print(json.dumps({'status': 'blocked', 'error': str(error)}))
        return 1
    print(json.dumps(report, indent=2, sort_keys=True))
    return 0 if report['expectations_met'] else 1


if __name__ == '__main__':
    raise SystemExit(main())
