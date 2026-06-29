# booksmith-ai

## Git Status
```text
```

## Detected Commands
```text

















































































































































```

## Logs

### npm-build.log
```text
LABEL=npm-build
CMD=npm ci && npm run build
DATE=Sat Jun 27 18:15:41 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found

added 357 packages, and audited 358 packages in 41s

144 packages are looking for funding
  run `npm fund` for details

2 moderate severity vulnerabilities

To address all issues (including breaking changes), run:
  npm audit fix --force

Run `npm audit` for details.
npm warn allow-scripts 3 packages have install scripts not yet covered by allowScripts:
npm warn allow-scripts   esbuild@0.28.1 (postinstall: node install.js)
npm warn allow-scripts   sharp@0.34.5 (install: node install/check.js || npm run build)
npm warn allow-scripts   unrs-resolver@1.12.2 (postinstall: node postinstall.js)
npm warn allow-scripts
npm warn allow-scripts Run `npm approve-scripts --allow-scripts-pending` to review, or `npm approve-scripts <pkg>` to allow.

> booksmith-ai@0.1.0 build
> next build

  Using cached swc package @next/swc-wasm-nodejs...
  Skipping creating a lockfile at /data/data/com.termux/files/home/AIFT/booksmith-ai/.next/lock because we're using WASM bindings
▲ Next.js 16.2.9 (Turbopack)

  Creating an optimized production build ...

> Build error occurred
Error: Turbopack is not supported on this platform (android/arm64) because native bindings are not available. Only WebAssembly (WASM) bindings were loaded, and Turbopack requires native bindings.

To build on this platform, use Webpack instead:
  next build --webpack

For more information, see: https://nextjs.org/docs/app/api-reference/turbopack#supported-platforms
    at ignore-listed frames
```
