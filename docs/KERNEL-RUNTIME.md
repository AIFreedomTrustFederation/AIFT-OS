# AIFT Kernel Runtime

The Kernel Runtime is the boot orchestrator for AIFT-OS.

It does not replace the operating system kernel of the host machine. It coordinates the federation-level AIFT kernel subsystems.

## Boot sequence

- Load configuration
- Run Discovery Engine
- Build Kernel Registry
- Publish kernel event
- Write boot report

## Commands

- `aiftd kernel boot`
- `aiftd kernel status`
- `aiftd kernel report`

## Runtime artifacts

- `registry/kernel-boot.json`
- `reports/kernel-boot.md`

These are ignored runtime state and should be regenerated from truth.
