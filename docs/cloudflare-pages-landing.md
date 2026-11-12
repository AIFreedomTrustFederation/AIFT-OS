# Cloudflare Pages: AIFT-OS landing frontend

This guide deploys **only the public Vite/React landing frontend**. It does not deploy the AIFT-OS Go control plane, Express API server, database, authenticated collectors, or federation operators. Do not expose secrets or private endpoints in `VITE_*` variables: Vite embeds them into public browser assets.

## Connect GitHub

1. In the Cloudflare dashboard, select **Workers & Pages → Create → Pages → Connect to Git** (wording may vary).
2. Authorize Cloudflare to access **AIFreedomTrustFederation/AIFT-OS**, granting only the repository permissions needed.
3. Select `main` as the production branch. Use preview deployments for pull requests.
4. Configure the build using the **repository root**, not the landing artifact directory.

| Setting | Value |
| --- | --- |
| Framework | Vite |
| Build command | `corepack enable && corepack pnpm install --frozen-lockfile && corepack pnpm --filter @workspace/aift-os-landing build` |
| Build output directory | `artifacts/aift-os-landing/dist/public` |
| Root directory | repository root |
| `NODE_VERSION` | `24` |
| `BASE_PATH` | `/` |

The landing Vite config already sets `base: process.env.BASE_PATH || "/"` and `build.outDir` to `dist/public`. Its package script is `vite build --config vite.config.ts`. The workspace pins `pnpm@11.25.0` using Corepack.

## Verify before public launch

- Build the landing frontend from a clean checkout with the frozen lockfile; do not publish the local Vite development server.
- Check that the resulting `index.html` and asset paths work under the assigned `https://<project>.pages.dev/` domain.
- Check direct navigation to every client-side route. If deep links return 404, add a Cloudflare Pages SPA fallback compatible with the app's routing and asset strategy.
- Confirm no browser bundle contains API keys, tokens, private configuration, or privileged federation capabilities.
- Confirm API calls are either disabled in the static preview or configured to reach an explicitly approved HTTPS backend with appropriate authentication, CORS, and access controls.
- Validate mobile rendering and preview deployment. Do not assume the public frontend is a live, authenticated federation operator console.
- Require explicit operator approval before production publishing or binding a custom domain.

## Scope and rollback

Cloudflare Pages Git integration can automatically deploy production commits after connection; do not connect production until approved. Draft PRs are for review, not publication. To roll back a deployed frontend, use the Cloudflare Pages deployment history to restore a verified prior deployment. Preserve the existing AIFT public portal and its domain; this setup does not replace it.
