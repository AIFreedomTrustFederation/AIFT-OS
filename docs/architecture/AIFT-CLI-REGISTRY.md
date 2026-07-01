# AIFT CLI Registry

AIFT commands must be registry-driven, deduplicated, and discoverable.

Rules:

- No duplicate help commands.
- No fake commands.
- No hardcoded phase-only behavior.
- Commands are discoverable by registry.
- Modules may later register commands directly.
