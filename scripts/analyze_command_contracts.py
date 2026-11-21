#!/usr/bin/env python3
"""Statically inspect root command contracts. Never execute repository commands."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import shlex

LIMIT = 2_000_000


def read_object(root, name):
    path = root / name
    if path.is_symlink():
        raise ValueError(name + ': symlink refused')
    if not path.is_file():
        raise ValueError(name + ': missing file')
    if path.stat().st_size > LIMIT:
        raise ValueError(name + ': file too large')
    data = path.read_bytes()
    def reject(value):
        raise ValueError('non-standard JSON constant: ' + value)
    value = json.loads(data, parse_constant=reject)
    if not isinstance(value, dict):
        raise ValueError(name + ': JSON object required')
    return value, hashlib.sha256(data).hexdigest()


def finding(code, message):
    return {'code': code, 'message': message}


def analyze(root):
    root = Path(root)
    report = {'repository': root.name, 'verification': 'not_run',
              'analysis': 'complete', 'source_hashes': {}, 'commands': [],
              'findings': [], 'limitations': [
                  'Root manifests only; no nested workspace discovery.',
                  'Heuristic shell analysis, not a shell interpreter or safety proof.',
                  'Commands, dependencies and services were not executed or verified.']}
    if root.is_symlink() or not root.is_dir():
        report['analysis'] = 'unverifiable'
        report['findings'].append(finding('invalid_root', 'Explicit directory required; symlink roots refused.'))
        return report
    root = root.resolve()
    try:
        manifest, digest = read_object(root, 'aift.repo.json')
        report['source_hashes']['aift.repo.json'] = digest
        commands = manifest.get('commands')
        if not isinstance(commands, dict) or not commands:
            raise ValueError('aift.repo.json: nonempty commands object required')
    except (OSError, ValueError, UnicodeError) as error:
        report['analysis'] = 'unverifiable'
        report['findings'].append(finding('invalid_manifest', str(error)))
        return report
    package = None
    package_error = None
    try:
        package, digest = read_object(root, 'package.json')
        report['source_hashes']['package.json'] = digest
        if not isinstance(package.get('scripts', {}), dict):
            raise ValueError('package.json: scripts object required')
    except (OSError, ValueError, UnicodeError) as error:
        package_error = str(error)
    for name, command in sorted(commands.items()):
        entry = {'name': name, 'findings': [], 'coverage': 'unverified'}
        report['commands'].append(entry)
        flags = entry['findings']
        if not isinstance(command, str) or not command.strip():
            flags.append(finding('invalid_command', 'Nonempty command string required.'))
            continue
        entry['command'] = command
        if re.search(r'\$\(|`|\$\{|\beval\b|\b(?:bash|sh)\s+-c\b', command):
            flags.append(finding('dynamic_shell', 'Dynamic shell or nested shell requires manual inspection.'))
        if re.search(r'\$HOME\b|~/|/data/data/|/home/|/Users/', command):
            flags.append(finding('machine_specific_path', 'Command depends on a machine-specific workspace or home.'))
        if re.search(r'--if-present\b', command):
            flags.append(finding('optional_gate', 'Optional scripts can be absent without failing.'))
        if re.search(r'\|\|\s*(?:true|:|echo\b)|;\s*(?:true|echo\b)', command):
            flags.append(finding('failure_masking', 'A later successful command may hide a failure.'))
        if name == 'verify' and re.fullmatch(r'\s*(?:git\s+status(?:\s+--short|\s+--porcelain)?|true|:|echo\s+.*?)\s*', command):
            flags.append(finding('diagnostic_only_gate', 'This command does not establish application verification.'))
        if re.search(r'\b(?:install|bootstrap|repair|deploy|publish|push|pull|sync)\b|\brm\s|\bgit\s+(?:reset|clean|checkout|switch)\b|[>]', command):
            flags.append(finding('possible_mutation', 'Command may change files, dependencies, Git state or external state.'))
        try:
            tokens = shlex.split(command)
        except ValueError:
            flags.append(finding('unparsed_shell', 'Shell quoting could not be parsed.'))
            continue
        # Only explicit root npm script invocations. Do not infer workspace scripts.
        scripts = re.findall(r'\bnpm\s+run\s+([A-Za-z0-9_:.-]+)', command)
        for script in scripts:
            if package is None or package_error:
                flags.append(finding('package_unavailable', package_error or 'Package metadata unavailable.'))
            else:
                body = package.get('scripts', {}).get(script)
                if not isinstance(body, str) or not body.strip():
                    flags.append(finding('missing_npm_script', 'Root npm script is missing or invalid: ' + script))
                else:
                    entry.setdefault('npm_scripts', {})[script] = body
                    if re.fullmatch(r'\s*(?:true|:|echo\s+.*?)\s*', body):
                        flags.append(finding('no_op_npm_script', 'Script has no verification assertion: ' + script))
        for token in tokens:
            if not token.startswith(('./', '.aift/')):
                continue
            if any(c in token for c in '$`*?[]\n'):
                flags.append(finding('dynamic_entrypoint', 'Entrypoint cannot be resolved statically.'))
                continue
            path = root / token
            parts = Path(token).parts
            cursor = root
            symlink = False
            for part in parts:
                cursor = cursor / part
                symlink |= cursor.is_symlink()
            if symlink or not path.resolve().is_relative_to(root):
                flags.append(finding('refused_entrypoint', 'Entrypoint escapes root or contains a symlink: ' + token))
            elif not path.is_file():
                flags.append(finding('missing_entrypoint', 'Referenced root file is missing: ' + token))
        if not flags:
            flags.append(finding('manual_review', 'No recognized issue; behavior and prerequisites remain unverified.'))
    return report


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('repositories', nargs='+', help='Explicit local repository directories')
    args = parser.parse_args()
    reports = [analyze(path) for path in args.repositories]
    print(json.dumps({'schema': 'aift.command-contract-analysis.v1',
                      'scope': 'offline_static_analysis', 'repositories': reports},
                     indent=2, sort_keys=True))
    return 1 if any(r['analysis'] == 'unverifiable' for r in reports) else 0


if __name__ == '__main__':
    raise SystemExit(main())
