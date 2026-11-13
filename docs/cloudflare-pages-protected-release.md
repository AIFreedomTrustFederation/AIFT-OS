# Cloudflare Pages protected release runbook

## Architecture and scope

The workflow `.github/workflows/cloudflare-pages-release.yml` deploys the **static AIFT-OS landing frontend only**. It does not host the Go control plane, Express API, databases, collectors, secrets, or authenticated operator endpoints. It does not alter the existing federation portal or domains.

The workflow runs **credential-free verification on relevant pull requests**. Production deployment remains **manual only** (`workflow_dispatch`), requires an exact `DEPLOY` confirmation, and runs only from `main`. A successful verification job packages the frontend as an Actions artifact. The deploy job consumes only that artifact, and is associated with the `cloudflare-production` GitHub environment. No credentials are available to the build job.

## One-time operator setup (required before deployment)

1. Create a Cloudflare Pages project named `aift-os` using Direct Upload / Wrangler. The project must exist before the workflow can deploy. This workflow intentionally does **not** automatically create an account, Pages project, or custom domain. Avoid attaching a Git repository to the same Pages project if you intend to deploy solely through Wrangler.
2. In Cloudflare, create a **scoped API token** with Cloudflare Pages **Edit** permission limited to the intended account. Do not use a global API key or commit a token to the repository. Record the Cloudflare account ID from the dashboard.
3. In GitHub repository **Settings → Environments**, create `cloudflare-production`. Configure **required reviewers**, **prevent self-review** if available, and restrict deployment branches to `main`. GitHub environment protection options and enforcement depend on repository visibility and plan. **If required reviewers cannot be enforced, do not run the release workflow until an equivalent external approval gate is established.** Creating an environment alone does not guarantee reviewer enforcement.
4. Add these **environment secrets** (not repository variables): `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_ACCOUNT_ID`. Add environment variable `CLOUDFLARE_PAGES_PROJECT=aift-os`. Never paste token values into chat, issues, PRs, logs, or terminal screenshots.
5. Configure branch protection/rulesets for `main`, require the relevant CI checks and reviews, and restrict who may push. The deployment workflow does not substitute for branch protection.
6. Review the Pages project visibility and ensure the static assets contain no privileged credentials. `VITE_*` values are **public** in browser bundles.

## Release procedure

1. Merge reviewed changes into `main` only after exact-head CI and any required approvals pass.
2. Open GitHub **Actions → Cloudflare Pages protected release → Run workflow**. Select `main` and type `DEPLOY`.
3. The verification job performs a frozen pnpm install, typechecks, JavaScript regression tests, federation evidence tests, canonical verification, builds the Vite landing, and checks output.
4. The deployment job waits at the `cloudflare-production` environment gate **only if the environment is configured to enforce it**. An authorized reviewer approves the release.
5. Wrangler uploads the verified artifact and the workflow checks that `https://aift-os.pages.dev/` returns HTML. Confirm the page renders and its static assets load in a browser; HTTP verification alone does not prove functional correctness.
6. Inspect the run summary, deployment history, and Cloudflare project. Record the release commit and URL. If verification fails, do not report a successful deployment.

## Security and limitations

- A Cloudflare token should be restricted to the intended account and Pages edit scope; rotate it if exposed.
- The workflow does not accept arbitrary shell input, deploy from PRs, or expose credentials to untrusted PR jobs.
- GitHub-hosted Linux x64 is the release build target. Keep the Linux ARM64 native-dependency compatibility repair separately reviewed.
- Pages preview URLs and custom domains may have different routing behavior. Test any client-side deep links and SPA fallback before announcing release readiness.
- A successful deployment does not make the AIFT-OS backend available. Do not claim federation operations are live without separate authenticated backend evidence.
- For rollback, select and restore a previously verified Pages deployment using Cloudflare's deployment history, subject to Cloudflare's supported rollback options.
- Cloudflare and GitHub free-plan quotas and environment approval features can change; verify current entitlements before the first release.

## Pull-request validation

Changes to the release workflow, landing frontend, workspace lockfile, or package manifest trigger the `Verify and package landing` job on pull requests. This checks the build without Cloudflare credentials or publishing. The artifact upload and `deploy-production` job are restricted to an explicitly confirmed manual dispatch from `main`. A passing PR validation is necessary but not sufficient for a production release; the protected environment and Cloudflare project must still be configured and approved.
