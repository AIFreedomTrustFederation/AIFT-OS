# Launcher and Shell Lessons

## Launcher lessons

- Do not track compiled binaries such as `bin/aiftd`.
- Keep the shell launcher thin.
- Use one workspace launcher: `~/AIFT/aift`.
- Use one repo launcher: `aift-os.sh`.
- Use one compiled binary: `bin/aiftd`.
- Avoid recursive wrapper chains.
- Never pass the binary path as a user command argument.

## Shell lessons

- Termux can handle long scripts, but ChatGPT-generated heredocs can become fragile.
- POSIX `sh` does not support Bash brace expansion such as `mkdir -p internal/{api,state}`.
- Use explicit directory lists in portable scripts.
- Prefer small idempotent scripts once the system grows.

## Go migration lessons

- Keep CLI parsing in `cmd/aift`.
- Keep business logic in `internal/*`.
- Generated outputs belong in `registry/`, `reports/`, `logs/`, and `var/`.
- Runtime state belongs in `var/`.
