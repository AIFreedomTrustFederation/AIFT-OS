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

The Tree game consumes `GET /v1/federation/tree` and renders:

- one federation root;
- seven living layers;
- Tree of Life and Tree of Knowledge branches;
- repository leaves;
- readiness glow, growth, level, XP, capabilities, and quests;
- filters for life, knowledge, ready, and blocked repositories.

Selecting a repository opens an evidence panel and can return to MoBox for `/inspect <repository>`. The game has no action-execution endpoint.

## World game

The World game consumes `GET /v1/federation/world` and renders:

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
