# Phase 7: Tests From Repository Reality

AIFT-OS tests must not assume fields, constructors, package shapes, or generated state.

Tests must discover the real repository API before asserting behavior.

Rules:

- Do not manually construct structs with guessed fields.
- Prefer real constructors such as config.Load().
- Test command surfaces through exported functions where possible.
- Keep generated runtime artifacts ignored or restored.
- Generated federation state must not cause permanent dirty worktrees.
- Tests should validate true behavior, not fake capability.
