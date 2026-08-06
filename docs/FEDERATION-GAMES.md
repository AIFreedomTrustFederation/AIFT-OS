# Federation Games

AIFT-OS exposes two standalone, borderless game worlds backed by the same evidence APIs as MoBox UXI:

- `http://127.0.0.1:8787/tree` — Federation Tree of Life
- `http://127.0.0.1:8787/world` — Federation World Game
- `http://127.0.0.1:8787/` — governed MoBox console

The Tree and World buttons in MoBox launch these routes instead of placing the game inside dashboard panels.

## Borderless contract

Each game occupies the entire browser viewport. It has no permanent sidebar, dashboard frame, card grid, or administrative navigation. Controls float over the world and disappear into the visual field. A small home control returns to MoBox, and each game links directly to the other.

Supported interaction:

- one-finger or mouse drag to pan;
- pinch or mouse wheel to zoom;
- tap or keyboard activation to inspect a node;
- fullscreen toggle where the browser permits it;
- bottom-sheet node details;
- evidence-derived quests and progression;
- direct travel between Tree and World.

The browser's own address and system bars are controlled by the browser and operating system. The in-page experience is borderless. A future installable PWA shell may remove more browser chrome when launched from the home screen.

## Tree of Life game

The Tree game consumes `GET /v1/federation/tree` and `GET /v1/federation/geometry` and renders:

- one federation root;
- seven living layers;
- Tree of Life and Tree of Knowledge branches;
- repository leaves;
- readiness glow, growth, level, XP, capabilities, and quests;
- filters for life, knowledge, ready, and blocked repositories.

Selecting a repository opens an evidence panel and can return to MoBox for `/inspect <repository>`. The game has no action-execution endpoint.

## World game

The World game consumes `GET /v1/federation/world` and `GET /v1/federation/geometry` and renders:

- visible repository location declarations;
- private, unmapped, and invalid counts without exposing hidden coordinates;
- a temporary phone-host cluster after explicit GPS permission;
- routes between the current phone and declared geographic nodes;
- mapping XP and privacy-neutral location quests.

Device coordinates remain in browser memory only. **Forget GPS**, page refresh, or page closure clears them. No external tiles, analytics, geocoding, or location-upload service is used.

## Truth boundary

The games visualize evidence. They do not change what is true.

- XP is not money, authority, ownership, morality, or execution proof.
- Ready is a repository capability classification, not proof every service is currently running.
- A geographic marker is a declared location scope, not territorial ownership.
- Temporary phone anchoring is visual context, not a permanent repository declaration.
- Approvals and writes remain in the MoBox governance surface.


## Shared Living Geometry renderer

Tree and World embed one progressive Canvas renderer over their existing interaction surfaces. It consumes the canonical geometry contract and displays:

- deterministic Mandelbrot repository identity;
- three-dimensional Fibonacci-sphere position;
- sacred-form rotational symmetry;
- toroidal identity and coherence phases;
- evidence-derived brightness and status color;
- the Eternal Now intention-to-reflection cycle;
- a selected repository's evidence-derived restoration quest.

The animated presentation clock changes only visual rotation. It never changes canonical identity, evidence, coherence, quest state, or authority. Reduced-motion preferences freeze presentation rotation while preserving the complete information surface.

Selecting an existing repository node synchronizes the geometry details without replacing the game's native touch, pointer, keyboard, pan, pinch, zoom, GPS, or fullscreen behavior. The restoration doorway returns to MoBox inspection; it does not execute an action from the game.

## Chromatic Coherence palette

The shared visual substrate follows an interwoven neon spectrum over near-black:

| Token | Color | Meaning |
|---|---|---|
| Ink | `#050609` | Unmanifest field and maximum contrast |
| Cyan | `#00E5E5` | Connection, communication, and living pathways |
| Teal | `#00AEB5` | Stable relational structure |
| Electric blue | `#147DF5` | Knowledge, depth, and spatial recursion |
| Living green | `#00E676` | Evidence-backed readiness and regeneration |
| Coral | `#FF4F46` | Embodied action and transformational crossings |
| Hot pink | `#FF256E` | Blockage, urgency, and boundary visibility |
| Orange | `#FF7A35` | Detected potential and incomplete emergence |

Repository identity selects harmonic colors from the complete spectrum, while operational status retains an accessible invariant: green is ready, orange is detected, and pink is blocked. Color never serves as the only status signal; labels and evidence remain present.
