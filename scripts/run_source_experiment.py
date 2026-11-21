#!/usr/bin/env python3
"""Run an explicitly supplied argv against an isolated, pinned source snapshot."""
import argparse
import io
import json
import os
from pathlib import Path, PurePosixPath
import re
import signal
import subprocess
import tarfile
import tempfile
import threading
import time


def git(root, *args):
    return subprocess.check_output(['git', '--no-optional-locks', '-C', str(root), *args],
                                   stderr=subprocess.PIPE, timeout=30)


def run(root, sha, argv, timeout=30, output_limit=65536, tool_paths=()):
    record = {'schema': 'aift.source-experiment.v1', 'expected_sha': sha,
              'argv': argv, 'status': 'blocked', 'scope': 'isolated_source_command',
              'limitations': ['Not an OS security sandbox; trusted commands only.',
                              'No production readiness or authorization certification.']}
    try:
        root = Path(root).resolve()
        if os.name != 'posix':
            raise ValueError('POSIX process groups required')
        if not re.fullmatch('[0-9a-f]{40}', sha):
            raise ValueError('Full lowercase 40-character commit SHA required')
        if not argv or any(not isinstance(x, str) or not x for x in argv):
            raise ValueError('Explicit nonempty argv required')
        if not 0 < timeout <= 300 or not 0 < output_limit <= 1048576:
            raise ValueError('Timeout must be 0..300 seconds and output limit 1..1048576 bytes')
        if any(not Path(p).is_absolute() or not Path(p).is_dir() for p in tool_paths):
            raise ValueError('Tool paths must be explicit absolute directories')
        record['tool_paths'] = list(tool_paths)
        if Path(git(root, 'rev-parse', '--show-toplevel').decode().strip()).resolve() != root:
            raise ValueError('Repository root required')
        if git(root, 'rev-parse', 'HEAD').decode().strip() != sha:
            raise ValueError('HEAD does not match expected SHA')
        if git(root, 'status', '--porcelain', '--untracked-files=all').strip():
            raise ValueError('Clean input checkout required')
        # Bound snapshot size before reading its archive into memory.
        listing = git(root, 'ls-tree', '-r', '-l', sha).decode()
        total = 0
        for line in listing.splitlines():
            metadata = line.split('\t', 1)[0].split()
            if metadata[0] not in ('100644', '100755'):
                raise ValueError('Symlinks and submodules are unsupported')
            total += int(metadata[3])
        if total > 20_000_000:
            raise ValueError('Snapshot exceeds 20 MB source limit')
        archive = git(root, 'archive', '--format=tar', sha)
        with tempfile.TemporaryDirectory(prefix='aift-experiment-') as tmp:
            workspace = Path(tmp) / 'source'
            workspace.mkdir()
            with tarfile.open(fileobj=io.BytesIO(archive)) as source:
                for member in source:
                    path = PurePosixPath(member.name)
                    if path.is_absolute() or '..' in path.parts:
                        raise ValueError('Unsafe archive path')
                    target = workspace.joinpath(*path.parts)
                    if member.isdir():
                        target.mkdir(parents=True, exist_ok=True)
                    elif member.isfile():
                        target.parent.mkdir(parents=True, exist_ok=True)
                        target.write_bytes(source.extractfile(member).read())
                        target.chmod(member.mode & 0o777)
                    else:
                        raise ValueError('Unsupported archive member')
            home = Path(tmp) / 'home'
            home.mkdir()
            env = {'PATH': os.pathsep.join([*tool_paths, os.defpath]), 'HOME': str(home), 'TMPDIR': tmp,
                   'LANG': 'C.UTF-8', 'PYTHONDONTWRITEBYTECODE': '1',
                   'npm_config_ignore_scripts': 'true', 'npm_config_offline': 'true'}
            started = time.monotonic()
            process = subprocess.Popen(argv, cwd=workspace, env=env,
                                       stdin=subprocess.DEVNULL, stdout=subprocess.PIPE,
                                       stderr=subprocess.PIPE, start_new_session=True)
            exceeded = threading.Event()
            logs = {'stdout': bytearray(), 'stderr': bytearray()}
            def stop():
                try:
                    os.killpg(process.pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
            def drain(stream, key):
                while True:
                    chunk = stream.read(4096)
                    if not chunk:
                        break
                    remaining = output_limit - len(logs[key])
                    logs[key].extend(chunk[:remaining])
                    if len(chunk) > remaining:
                        exceeded.set()
                        stop()
                stream.close()
            threads = [threading.Thread(target=drain, args=(stream, key), daemon=True)
                       for stream, key in ((process.stdout, 'stdout'), (process.stderr, 'stderr'))]
            for thread in threads:
                thread.start()
            timed_out = False
            try:
                process.wait(timeout=timeout)
            except subprocess.TimeoutExpired:
                timed_out = True
            finally:
                # Also terminate children left running after the command exits.
                stop()
                process.wait()
                for thread in threads:
                    thread.join(timeout=5)
            record.update(exit_code=process.returncode,
                          elapsed_seconds=round(time.monotonic() - started, 6),
                          stdout=logs['stdout'].decode(errors='replace'),
                          stderr=logs['stderr'].decode(errors='replace'))
            record['status'] = ('timed_out' if timed_out else 'output_limit' if exceeded.is_set()
                                else 'success' if process.returncode == 0 else 'failed')
        # Recheck input after execution, not just before snapshotting.
        if (git(root, 'rev-parse', 'HEAD').decode().strip() != sha or
                git(root, 'status', '--porcelain', '--untracked-files=all').strip()):
            record['status'] = 'blocked'
            record['error'] = 'Input checkout changed during experiment'
    except (OSError, ValueError, subprocess.SubprocessError, tarfile.TarError) as error:
        record['status'] = 'blocked'
        record['error'] = str(error)
    return record


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--repo', required=True)
    parser.add_argument('--sha', required=True)
    parser.add_argument('--timeout', type=float, default=30)
    parser.add_argument('--tool-path', action='append', default=[])
    parser.add_argument('argv', nargs=argparse.REMAINDER)
    args = parser.parse_args()
    argv = args.argv[1:] if args.argv[:1] == ['--'] else args.argv
    result = run(args.repo, args.sha, argv, args.timeout, tool_paths=args.tool_path)
    print(json.dumps(result, indent=2, sort_keys=True))
    return 0 if result['status'] == 'success' else 1


if __name__ == '__main__':
    raise SystemExit(main())
