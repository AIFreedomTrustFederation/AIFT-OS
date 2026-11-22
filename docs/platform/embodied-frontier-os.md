# AIFT Embodied Frontier OS — Robots, Future Phones and Portable Intelligence

> **Research and simulation roadmap, not shipped robot firmware, phone OS, physical device compatibility or functional-safety certification.** Real physical actuation, camera/mic access, phone bootloader changes and firmware flashing require separate informed operator consent and hardware-specific safety controls.

## North star

**One intelligence, many bodies; human sovereignty before autonomy.** An approved user workspace, privacy-preserving personal memory, intent and skills may move among a workstation, phone, wearable and robot. **Permissions never move implicitly with the data.** Real-world movement and sensor access must pass device-local, independently checkable authorization, valid perception and safety-state constraints.

~~~mermaid
flowchart TB
    U["Human owner: permission, revocation, override"] --> C["Portable encrypted workspace / audited skill"]
    C --> PH["Virtual phone / wearable"]
    C --> AG["Aetherion / approved model: proposes a bounded plan"]
    AG --> P["Device-specific policy checker"]
    P --> S{"Independent safety and capability gate"}
    S -->|allowed in simulator| R["ROS 2 robot / sensor and actuator simulator"]
    S -->|rejected, stale sensor, timeout| STOP["Safe stop / no action"]
    U --> STOP
    R --> E["Measured feedback and recorded evidence"]
    E --> C
~~~

**The model and the safety controller are different components.** ROS 2 middleware and ros2_control are useful integration infrastructure, but neither alone establishes certified physical safety. A language model cannot override an independent e-stop, watchdog, restricted region, force or speed boundary, sensor freshness requirement or operator revocation.

## Device targets and bounded expectations

| Class | Engineering baseline | Minimum negative tests | Hardware readiness |
|---|---|---|---|
| Robot simulation | ROS 2-compatible sensor/plan/action protocol; deterministic simulated motor gate | Network loss, malformed AI command, stale sensor, timeouts, collision or geofence violations, human stop | No physical action without independent safety engineering |
| MCU/RTOS testbed | Supported Zephyr/micro-ROS or similar target with validated real-time deadlines | Watchdog reset, missed deadlines, bus fault, command rejection | Specific board and controller review only |
| Future AIFT phone | AOSP-derived or comparable supported runtime; ARM64 emulator first; real app identity/session | Permission denial, offline sync, app restart, storage encryption boundaries, battery/thermal limits | Carrier/modem/IMS/emergency calling, camera, power, HAL and OTA must be proven on model |
| Wearable/ambient | Constrained local private sensor and notification interfaces | Consent expiration, data minimization, missing sensor, battery depletion | Specific device and privacy review |
| Multi-device federation | End-to-end encrypted or suitably protected consent-scoped task handoff | Revocation during transfer, conflict, rollback, invalid identity, loss of network | Never transfer a device/robot privilege with the user workspace |

## Proposed system interfaces

**Portable workspace envelope:** schema version, user-controlled identity reference, encrypted data scope, provenance, retention/forgetting policy, target-appropriate capabilities and approval time. The envelope is *not* a universal permission token.

**Device capability description:** exact device model or simulator, operation class, locality, scope, physical vs simulated, consent, independent controller and sensor state, expiry and incident logging. Deny unknown operations.

**Embodied task protocol:** observation source and age, bounded goal, model proposal, deterministic validator output, task timeout, controller sequence number, cancel/e-stop state, action outcome and fault evidence.

**AI uncertainty and fallback:** distinguish perception confidence, model uncertainty, calibration and physical-sensor validity; reject unsupported claims and degrade to safe no-action when an external dependency is unavailable.

**Resource-aware inference:** CPU/GPU/NPU/RAM/battery/thermals, privacy mode, network outage and sustainable performance budgets. Prefer quantified power/latency/accuracy baselines over self-assigned 'optimized' scores.

## Measurable frontier research tracks

1. **Continuum memory:** replicate *selected* user-authored state across virtual phone and desktop, with separate permissions, explicit deletion and audit. Measure disclosure and restore behavior.
2. **Consent-carrying skills:** make each skill request short-lived, operation-scoped grants; test denial, revocation and expired permissions.
3. **Embodied digital twin:** verify sensor-to-plan-to-independent-gate-to-simulated-action and capture baseline vs agent task completion and false-accept rates.
4. **Graceful AI degradation:** test whether simulator stops deterministically when model times out, sensors are missing or remote service fails.
5. **Adaptive interface:** responsive desktop-to-phone-to-robot control view that labels predicted, simulated and observed states separately.
6. **On-device inference:** measure cold start, token/second or task latency, memory, power/thermal proxy and offline task success with supported model artifacts.
7. **Multi-agent cooperation:** simulate task handoff without silent delegation of location, microphone/camera or actuator access.

## Embodied milestone chain

| Milestone | Demonstration | Current claim |
|---|---|---|
| E0 | Target matrix, threat model, baseline simulator, interface designs | Proposed |
| E1 | Reproducible ROS 2 robot simulation with independent negative-tested safety gate | Unverified |
| E2 | Virtual phone ↔ robot permission protocol with revocation | Unverified |
| E3 | A local-first agent completes a bounded simulated task behind the safety gate | Unverified |
| E4 | Consented cross-device memory/skill handoff and forgetting | Unverified |
| E5 | Android emulator app and performance/accessibility/offline tests | Unverified |
| E6 | Multi-device simulation and fault-injection acceptance | Unverified |
| E7 | Reproducible candidate builds and device-specific safe update/recovery | Unverified |
| E8 | Independently assessed physical safety and approved hardware trials | Unverified |

## Non-negotiable physical safety

- Simulation-only changes may be proposed through routine authorized draft PRs. **Never operate a physical actuator, drone, vehicle, industrial device or nearby-human robot through this automation.**
- Require human-supervised physical controls, independent e-stop and carefully scoped environment before hardware experiments.
- Consent for device data remains granular and revocable. No covert sensors, tracking people, surveillance, recording or unrestricted remote control.
- Do not turn a successful software test into a functional-safety compliance or certification claim. Seek applicable independent assessment and regulations before physical operation.
- Device manifests and AI skills cannot grant themselves their own privileges.

## Related AIFT repositories

TheMindofAll owns model/memory artifacts; AIFT-Runtime owns local runtime adapters; AIFT-Forge develops skill/application interfaces; AIFT-OS coordinates verified plans and evidence; VPS supports hosted infrastructure; AIFT-Genesis owns canonical federation metadata and identity templates. Each must opt into integration through documented versioned contracts, not implied code ownership.

## Primary documentation

[ROS 2](https://docs.ros.org/) · [ros2_control](https://control.ros.org/) · [micro-ROS](https://micro.ros.org/) · [Zephyr](https://docs.zephyrproject.org/) · [AOSP](https://source.android.com/) · [Linux kernel](https://docs.kernel.org/). Use pinned compatible versions and inspect license/support dates at implementation time.

Return to [universal OS roadmap](README.md), [boot targets](boot-and-device-compatibility.md) and [verification](verification-and-release.md).
