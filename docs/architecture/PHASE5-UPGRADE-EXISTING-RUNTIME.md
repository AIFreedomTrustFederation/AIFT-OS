# AIFT-OS Phase 5: Upgrade Existing Runtime

AIFT-OS already has runtime, supervisor, services, scheduler, and verify commands.

Phase 5 upgrades the existing runtime instead of creating a second system.

Rules:

- Do not hardcode repository names.
- Discover repositories from the workspace.
- Discover services from repository evidence.
- Keep AIFT-OS as the runtime owner.
- Treat all other repositories as mounted source packages.
- Runtime state belongs in AIFT-OS var/runtime-state.json.
- Registry output belongs in AIFT-OS registry and reports.
- Verify must not leave mounted repositories dirty.
