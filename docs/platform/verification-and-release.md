# Verification, Safety, Release and Multi-Repository Handoffs

> This document defines **proposed acceptance gates**. Do not promote any planned milestone to implemented, installed, safe, deployed or certified without matching evidence from its owner.

## Evidence record template

For every milestone and dependent PR, publish a dated, inspectable record:

| Field | Required value |
|---|---|
| milestone / owner | AIFT milestone ID and repository that owns result |
| source / target | Exact default/base/head SHA, impacted producer/consumer version |
| environment | Native VM, QEMU/OVMF, browser, Android emulator, robot simulator, staging host or exact physical device model |
| test | Reproducible commands, configuration, dependencies and expected negative cases |
| result | PASS, FAIL, PENDING, BLOCKED, UNRUN or NEEDS APPROVAL; never collapse categories |
| artifacts | Logs, checks, screenshots if relevant, image/source hashes and accessible links |
| limitations | Unsupported hardware, browser-only tests, mock/fixture vs real services, unrun checks |
| owner actions | Required authority, release/recovery steps, human approval for physical/high-risk actions |
| rollback | Virtual-image, OS/installer and service migration recovery documented and verified |
| next | Single executable dependency-linked action; no duplicate PR based on stale head |

## Acceptance matrices: which evidence counts?

| Claim | Minimum meaningful proof |
|---|---|
| AIFT boots | Exact built image + QEMU cold boot, kernel/init logs, reboot and image digest |
| AIFT boot manager works | Test target selection and clean failure recovery in virtual firmware; physical NVRAM untouched |
| AIFT native services work | Started services, real API replies, logs, offline/restart/denial tests |
| AIFT desktop works | App manifest → real launch → working window and file/status UI on a native session |
| Hosted OS works | Authenticated isolated session and real backend APIs, storage/restore and security denial |
| Windows dual-boot works | Independent original Windows boot verified in disposable authorized virtual disk |
| Intel Mac works | Exact model, supported firmware/security settings and tested install/recovery path |
| Apple Silicon works | Exact model and Asahi-supported boot path, not generic PC UEFI |
| Future phone OS works | Named device/virtual target with app, network, permissions, power, update/rollback tests; telephony if claiming a real phone |
| Robot acts safely | Independent controller/safety architecture plus simulated fault tests; real safety requires separate hardware engineering and approval |
| AI autonomy works | Local model proposal → policy gate → approved deterministic action → exact outcome with negative case |
| Deployment ready | Artifact provenance, independent review, recovery, permission/release gate and observed live environment |

## Coordinated hourly engines

The scheduled automations are engineering coordination prompts, not guaranteed cloud computers, privileged GitHub Apps, permanent robots, firmware flash tools or physical test rigs.

| Engine | Role | Safety boundary |
|---|---|---|
| Federation Continuous Engineering | OS/native Linux/UEFI/QEMU and integration contracts | Draft PR/isolated build, no physical boot modifications |
| UXI Continuous Experience | Living Universe + OS desktop/apps and preboot mock designs | Browser preview ≠ firmware or native service |
| Cloudflare Universal Release Readiness | Native/hosted releases, artifact/rollback evidence | No production deployment or signing policy change without approval |
| VPS opt-in merger | Only eligible, explicitly labeled routine VPS PRs | Exclude firmware, boot, install, signing, security, privileged OS infrastructure |
| Embodied Frontier OS | Simulated robot/mobile, portable intelligence, safety gate | No real actuation, surveillance sensors or phone flashing without approval |

All read the same versioned roadmap and reference exact GitHub SHAs. If automation environment lacks a real runner, model, browser or simulator, the report must label testing UNRUN and produce concrete next steps rather than inventing success.

## Repository authority and high-risk changes

- **AIFT-OS:** platform integration coordination; no takeover of owner repos.
- **AIFT-Genesis / AI-Freedom-Trust / AetherianGovernance:** constitutional, identity, doctrine and governance contracts need separate explicit review and cannot be silently rewritten under a docs sync.
- **VPS:** deployments, routing, secrets, release and host safety require explicit authorization.
- **Aether_Coin_biozonecurrency and business repos:** payments, custody, financial claims, customer data and physical delivery/medical/regulated actions remain behind owner rules.
- **mobox / OpenMontage / other upstream-derived repos:** preserve licenses, upstream attribution and specific project instructions; do not relabel upstream forks as core OS.
- **Robot/phone:** human safety, sensors, firmware, bootloader and physical actuator control remain separate opt-in workstreams.

Documentation-only draft PRs may outline future contracts without granting new execution rights. Never treat an informational cross-repository link as implemented integration.

## First safe development slice

1. Review and approve architecture and exact ownership/compatibility matrix in a documentation draft.
2. Create a bounded BOOT-01 implementation plan in AIFT-OS with pinned Linux/build environment, QEMU/OVMF profile and virtual disk.
3. Build a minimal bootable development image **in a disposable VM**; log serial and kernel startup; verify shutdown/reboot; hash artifacts.
4. Start one supervised AIFT-Runtime service with real status API and offline restart test.
5. Add a basic AIFT graphical desktop launcher that calls only this verified service; keep browser-only previews labeled.
6. Release Readiness verifies recovered boot after an interrupted virtual update and checks artifacts and operator requirements.
7. Only after virtual proof should independent phone, Apple and robotics simulations graduate toward exact hardware model research.

## QA for documentation-only architecture PR

- All Markdown links resolve to valid repository paths or publicly accessible references; proposed docs are in the same PR.
- Mermaid blocks are syntactically plausible and SVGs are well-formed, accessible, free of external resource loads, text labels remain legible at small scale.
- No README/status claim is changed from conceptual to implemented.
- Inspect AGENTS.md, verify only intended files changed, run the owning repository's canonical tests as available, use exact-head GitHub Actions.
- Report any CI blocked by GitHub workflow approvals explicitly; do not bypass review/branch protections.

Return to [platform overview](README.md), [boot profiles](boot-and-device-compatibility.md), [frontier devices](embodied-frontier-os.md) and [desktop contracts](desktop-and-service-contracts.md).
