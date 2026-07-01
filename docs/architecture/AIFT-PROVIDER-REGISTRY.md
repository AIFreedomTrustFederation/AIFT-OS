# AIFT Provider Registry

The Provider Registry is the runtime-agnostic foundation for AIFT-OS.

The scheduler should eventually stop knowing about Go, Node, Python, Rust, Make, Java, Docker, or any other runtime directly.

Instead, providers describe:

- detection files
- required capabilities
- build commands
- test commands
- sync support
- async support

This keeps the scheduler module-agnostic, runtime-agnostic, provider-agnostic, and compatible with both synchronous and asynchronous execution.
