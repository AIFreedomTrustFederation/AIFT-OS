# Portable source experiments

This opt-in runner executes an operator-supplied argv in a temporary snapshot of
a clean local Git checkout at an exact SHA. It never reads executable commands
from manifests. It needs POSIX, Git and Python 3.10+; provision command tools
separately. No network fetching or dependency installation is built in.

```sh
python3 -B scripts/run_source_experiment.py --repo /absolute/checkout \
  --sha FULL_COMMIT_SHA --timeout 30 -- python3 -B -m unittest discover
python3 -B scripts/demo_gate_experiment.py > /tmp/gate-comparison.json
python3 -B -m unittest discover -s tests -p test_source_experiment.py -v
```

Use repeated `--tool-path /absolute/bin` to add explicitly provisioned tools.
Default PATH contains only standard OS directories. Inherited credentials and
environment variables are not passed to commands. HOME and TMPDIR are temporary;
npm lifecycle scripts are disabled and npm offline mode is enabled by default.
There is no Git checkout metadata or dependency installation in the snapshot.
Symlinks, submodules and snapshots larger than 20 MB are refused. Git archives
honor export attributes: commands see exported source, not necessarily every
tracked file. LFS files remain pointers unless tracked source contains bytes.

Results record argv, exact expected SHA, tool directories, exit code, elapsed
wall time and bounded stdout/stderr (64 KiB each). Timeout is at most 300 seconds.
The process group is killed after completion, timeout or excessive output.
Source HEAD and checkout cleanliness are checked before and after execution.
Exit 0 means only that the explicit command exited zero within limits; all
failed, blocked, timed-out and output-limit outcomes exit 1. CLI errors exit 2.

Temporary directories provide filesystem separation, **not security isolation**.
Commands still have the invoking user's host and network access and can escape
the directory or start detached processes. Use only reviewed, trusted source and
commands. Untrusted workloads require an external container/VM sandbox. CPU,
disk and memory are not hard-limited; this is not a multi-tenant execution service.
Logs can contain sensitive output: keep results local until reviewed.

The demonstration creates six disposable fixture commits with synthetic npm
stages: healthy, missing typecheck and failing test, each under the baseline
optional command and PR #15's mandatory command. It expects the baseline to
miss the absent stage, the candidate to catch it, and both to reject a failed
stage. Upstream SHAs identify the source of the command strings; fixture SHAs
identify the actual code executed. This demonstrates failure detection, not
actual Forge compilation, deployment readiness or a performance improvement.

The static analyzer in `analyze_command_contracts.py` can flag the optional
baseline before this experiment. It does not automatically authorize execution.
Forge PR #15 already proposes the correction; this experiment does not duplicate
that repository change or merge it.
