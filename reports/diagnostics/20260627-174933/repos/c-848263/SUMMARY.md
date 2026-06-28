# c-848263

## Git
```text
```

## Detected Commands
```text
npm:dev=npm run dev
npm:build=npm run build
npm:build:dev=npm run build:dev
npm:lint=npm run lint
npm:preview=npm run preview
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

### npm-build.log
```text
REPO=c-848263
LABEL=npm-build
CMD=npm ci && npm run build
DATE=Sat Jun 27 17:56:35 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found

added 421 packages, and audited 422 packages in 33s

73 packages are looking for funding
  run `npm fund` for details

20 vulnerabilities (2 low, 8 moderate, 10 high)

To address all issues, run:
  npm audit fix

Run `npm audit` for details.
npm warn allow-scripts 3 packages have install scripts not yet covered by allowScripts:
npm warn allow-scripts   @swc/core@1.7.39 (postinstall: node postinstall.js)
npm warn allow-scripts   esbuild@0.21.5 (postinstall: node install.js)
npm warn allow-scripts   esbuild@0.25.0 (postinstall: node install.js)
npm warn allow-scripts
npm warn allow-scripts Run `npm approve-scripts --allow-scripts-pending` to review, or `npm approve-scripts <pkg>` to allow.

> vite_react_shadcn_ts@0.0.0 build
> vite build

failed to load config from /data/data/com.termux/files/home/AIFT/c-848263/vite.config.ts
error during build:
Error: Failed to load native binding
    at Object.<anonymous> (/data/data/com.termux/files/home/AIFT/c-848263/node_modules/@swc/core/binding.js:333:11)
    at Module._compile (node:internal/modules/cjs/loader:1873:14)
    at Module._extensions..js (node:internal/modules/cjs/loader:2013:10)
    at Module.load (node:internal/modules/cjs/loader:1596:32)
    at Module._load (node:internal/modules/cjs/loader:1398:12)
    at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
    at Module.require (node:internal/modules/cjs/loader:1619:12)
    at require (node:internal/modules/helpers:191:16)
    at Object.<anonymous> (/data/data/com.termux/files/home/AIFT/c-848263/node_modules/@swc/core/index.js:49:17)
    at Module._compile (node:internal/modules/cjs/loader:1873:14)
```
