# AIFT-OS Architecture Evolution

AIFT-OS began as a shell-based federation control plane.

The project evolved through these stages:

1. Directory cleanup and control-plane layout.
2. Shell command dispatcher.
3. Federation registry and reports.
4. Plugin and manifest architecture.
5. Go kernel introduction.
6. Launcher stabilization.
7. Runtime service layer.
8. Internal API and supervisor foundation.

The current direction is a Go-based federation operating system with shell scripts only for bootstrap, install, tests, and archival migration utilities.
