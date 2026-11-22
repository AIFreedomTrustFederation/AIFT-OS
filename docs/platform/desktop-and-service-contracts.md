# AIFT Desktop, Hosted Session and Service Contracts

> **Design-only technical interface guidance.** No API endpoint, privileged app, identity provider, file service, hardware device or hosted OS is claimed deployed by this document. Implemented commands remain what the owning repositories actually ship.

## A single user experience, multiple truthful capability profiles

| Interaction | Local Linux AIFT session | Hosted browser session | Pre-boot UEFI | Robot/phone prototype |
|---|---|---|---|---|
| Windowed apps | Native compositor or webview as implemented | Browser virtual windows | **Not an OS desktop** | Device-specific adapted UI |
| Files | Scoped host filesystem service | Isolated user data service | Read-only EFI capabilities where explicitly supported | Per-device explicit grants |
| Hardware sensors | Supported drivers and user permissions | No direct access absent approved bridge | Limited UEFI environment | Permissions, safety controllers, sensor validity |
| AI | Local or explicitly authorized provider | Scoped hosted local/remote model | Not presumed available | Bounded plan through independent safety checks |
| Session | OS login + local user | Authenticated isolated server session | Boot target selection only | Hardware-specific session and consent |
| Updates | Signed/versioned native image/system packages | Hosted deployment and rollback | Boot entry/firmware changes require separate review | Vendor/model-specific OTA or controlled test |

## Proposed API / package boundaries

All endpoints below are **illustrative names and schemas**, not verified deployed HTTP routes. Before implementation the owning repository must publish actual API source, schema version and tests.

- **/v1/platform/capabilities:** supported target, runtime kind, host model, available permissions, version and unavailable reason; never infer privileges from browser user agent.
- **/v1/platform/health:** observed boot/service ID, process state, measurement timestamp, actual dependencies and degraded conditions; status is not optimistic.
- **/v1/apps:** discover signed or otherwise verified packages and compatible entrypoints, owner, files scope and capability requests; no arbitrary shell command in public manifest.
- **/v1/sessions:** authenticated creation, expiry, isolation/tenant identity, optional persistence and termination; require CSRF/session protections and no credential leakage.
- **/v1/activities:** consent-reviewed model proposals, real executed action evidence, cancellation and future progress; never represent a plan as completed execution.
- **/v1/federation/evidence:** verified producer repo/commit, schema, consumer, check run evidence and source links; unavailable sources remain unknown.

## App manifest (conceptual fields)

| Property | Purpose |
|---|---|
| app_id, owner, version | Stable source authority and upgrade scope |
| targets | Native Linux, hosted browser or supported device profile only |
| entrypoints | Verified launch targets in the selected platform; no unrestricted remote execution |
| artifact_digest | Reproducibility and update provenance |
| grants_required | Explicit user and host policy capabilities |
| filesystem_scope | Private, user-approved and system paths kept separate |
| identity_scope | Least-privilege access to user and federation services |
| offline_behavior | Defined local fallback and graceful unavailable state |
| update_and_rollback | Recovery and migration compatibility with observed schema |

## UXI acceptance: public Tree → real desktop

The public Living Universe HTML remains the canonical visual benchmark. The operating-system shell must implement an actual application journey before adding more decorative portals:

1. Discover an application from a verified manifest and mark unsupported target honestly.
2. User selects app; system obtains any necessary explicit grant.
3. Launcher starts a bounded process/isolated service **or** supported hosted application, and observes actual start success/failure.
4. Window receives keyboard/pointer/touch focus, can be closed and restored where platform-supported, and displays real app state.
5. File explorer can list only authorized content and enforces denied access in tests.
6. Settings and notifications reflect real services; empty unavailable states are legitimate.
7. Disconnect/terminate session without losing user data unexpectedly; document user-initiated export and deletion.

Accessibility: keyboard-only traversal, visible focus, screen-reader labels, sensible contrast, reduced-motion mode, low-power fallback, touch targets and no hidden modal interception; verify on Android-sized and desktop widths without claiming this proves native-window-manager behavior.

## Semantic Tree contract

- **Roots:** real repositories, source and provenance. Never fruits.
- **Trunk:** actual tested federation relationships.
- **Branches:** implemented systems and subsystems.
- **Leaves:** documented knowledge and capabilities.
- **Flowers:** emerging work with explicit draft evidence.
- **Fruits:** actual functioning visitor-facing product, with a working destination and lineage.

Navigation includes reversible ancestry, search, a source-backed inspector, accurate readiness badges and no fabricated XP or simulated operational telemetry. The immersive UI must not invent governance authority or represent experimental services as production.

## Security and operational split

Browsers have no kernel privileges. Native agents run at minimum privilege, behind authenticated service APIs. A phone controller is not a robot safety controller. UEFI boot menus cannot depend on an OS compositor or cloud session. No host shell/installer/firmware actions may be bound to a public UI without a separately approved, narrowly scoped and reviewed privileged service.

Build one complete proof slice before scaling: **AIFT booted local service → actual capability API → working desktop app**, and separately **authenticated hosted session → same app contract**. Measure and compare both, preserving their capability differences.

Return to [platform roadmap](README.md) and [verification criteria](verification-and-release.md).
