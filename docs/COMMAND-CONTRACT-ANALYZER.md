# Offline command-contract analysis

Run from the repository root with explicit local checkout directories:

```sh
python3 -B scripts/analyze_command_contracts.py . ../AIFT-Forge ../AIFT-Runtime
python3 -B -m unittest discover -s tests -p test_command_contracts.py -v
```

This dependency-free analyzer reads root `aift.repo.json` and, for explicit
`npm run NAME` references, root `package.json`. It never invokes commands from
either file, installs dependencies, contacts providers, or changes inputs.
JSON output identifies source hashes, declared command strings, and findings.
Treat command strings as untrusted data when displaying or processing output.

Findings identify optional gates, diagnostic-only verification, absent or no-op
npm scripts, missing local entrypoints, machine-specific paths, dynamic shell,
possible mutations and simple failure masking. Symlink roots, metadata files,
and entrypoint path components are refused. Metadata reads are bounded to 2 MB.
Do not analyze concurrently changing or adversarially writable directories:
these checks are not a race-resistant filesystem sandbox.

Exit 0 means analysis completed, including reports containing findings. Exit 1
means at least one root or command manifest was unavailable or invalid. Argument
errors use argparse exit 2. Every report says `verification: not_run`; neither
exit 0 nor absence of recognized issues establishes readiness or authorization.
Invalid package metadata is reported when an npm command needs it.

This is deliberately a heuristic, not a shell parser or permission gate. It does
not resolve nested workspace packages, npm shorthand, script call chains,
arbitrary shell control flow, executable permissions, tool installations, or
runtime prerequisites. An existing script can still fail or have side effects.
Mutations hidden inside scripts may be missed. Root metadata hashes identify
the bytes inspected, not a Git commit or authenticated provenance.

Use this report alongside `scripts/upgrade_inventory.py`, repository AGENTS.md,
and canonical QA results. Findings should guide scoped follow-up proposals,
never automatic command execution or repository authority reclassification.
