# AIFT-OS

A living, federated operating system governed by evidence, not authority — the public-facing landing page and architectural reference for the AI Freedom Trust Federation.

## Run & Operate

- `pnpm --filter @workspace/api-server run dev` — run the API server (port 5000)
- `pnpm run typecheck` — full typecheck across all packages
- `pnpm run build` — typecheck + build all packages
- `pnpm --filter @workspace/api-spec run codegen` — regenerate API hooks and Zod schemas from the OpenAPI spec
- `pnpm --filter @workspace/db run push` — push DB schema changes (dev only)
- Required env: `DATABASE_URL` — Postgres connection string

## Stack

- pnpm workspaces, Node.js 24, TypeScript 5.9
- API: Express 5
- DB: PostgreSQL + Drizzle ORM
- Validation: Zod (`zod/v4`), `drizzle-zod`
- API codegen: Orval (from OpenAPI spec)
- Build: esbuild (CJS bundle)

## Where things live

- Landing page: `artifacts/aift-os-landing/src/` — React + Vite, no backend
- Theme: `artifacts/aift-os-landing/src/index.css` — dark palette, Syne + Space Mono fonts
- Architecture sections: `artifacts/aift-os-landing/src/components/` — ArchitectureRings, Trees, LivingLayers, LivingSpectrum, DiscoveryLifecycle, Hero, Footer
- API server: `artifacts/api-server/src/` — Express 5, health endpoint only currently

## Architecture decisions

- Landing page is presentation-only (no backend). All architecture content is static/animated in the frontend.
- Canvas tree renderer (`Trees.tsx`) uses synchronous recursive `drawBranch` — no `setTimeout` fan-out to avoid unbounded timer growth and frame jank.
- Governance model: 12 Validator Orders are constitutional domains (not people), 24 Elders are a role-based council (not a fixed list), Constitutional Core is the federation's ground truth.
- Seven Living Layers are permanent architectural domains; repositories are implementations of layers, not the architecture itself.
- Tree of Life (runtime/teal) and Tree of Knowledge (knowledge/amber) share the same underlying data registries — two views, one root.

## Product

The AIFT-OS landing page explains the complete federation architecture through 7 interactive sections: governance wheel-within-wheels diagram, seven living layers, dual tree visualizations, living spectrum state system, discovery lifecycle, core philosophy, and sovereign vision. Scroll-driven with Framer Motion animations throughout.

## User preferences

_Populate as you build — explicit user instructions worth remembering across sessions._

## Gotchas

_Populate as you build — sharp edges, "always run X before Y" rules._

## Pointers

- See the `pnpm-workspace` skill for workspace structure, TypeScript setup, and package details
