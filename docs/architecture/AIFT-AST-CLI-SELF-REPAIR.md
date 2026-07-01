# AIFT AST CLI Self-Repair

This phase replaces brittle text patching with Go AST/source-aware CLI repair.

It enforces:

- every switch command appears exactly once in help
- every help command corresponds to a real switch command
- duplicate help commands are removed
- duplicate switch cases are removed
- generated command registry reports are written

This does not fake commands or invent functionality.
