# biozone-harmony-boost

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
REPO=biozone-harmony-boost
LABEL=npm-build
CMD=npm ci && npm run build
DATE=Sat Jun 27 17:55:11 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found
npm warn deprecated glob@10.5.0: Old versions of glob are not supported, and contain widely publicized security vulnerabilities, which have been fixed in the current version. Please update. Support for old versions may be purchased (at exorbitant rates) by contacting i@izs.me

added 453 packages, and audited 454 packages in 29s

93 packages are looking for funding
  run `npm fund` for details

found 0 vulnerabilities
npm warn allow-scripts 1 package has install scripts not yet covered by allowScripts:
npm warn allow-scripts   @swc/core@1.15.41 (postinstall: node postinstall.js)
npm warn allow-scripts
npm warn allow-scripts Run `npm approve-scripts --allow-scripts-pending` to review, or `npm approve-scripts <pkg>` to allow.

> biozone-harmony-boost@0.0.0 build
> vite build

failed to load config from /data/data/com.termux/files/home/AIFT/biozone-harmony-boost/vite.config.ts
error during build:
Error: Failed to load native binding
    at Object.<anonymous> (/data/data/com.termux/files/home/AIFT/biozone-harmony-boost/node_modules/@swc/core/binding.js:345:11)
    at Module._compile (node:internal/modules/cjs/loader:1873:14)
    at Module._extensions..js (node:internal/modules/cjs/loader:2013:10)
    at Module.load (node:internal/modules/cjs/loader:1596:32)
    at Module._load (node:internal/modules/cjs/loader:1398:12)
    at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
    at Module.require (node:internal/modules/cjs/loader:1619:12)
    at require (node:internal/modules/helpers:191:16)
    at Object.<anonymous> (/data/data/com.termux/files/home/AIFT/biozone-harmony-boost/node_modules/@swc/core/index.js:66:17)
    at Module._compile (node:internal/modules/cjs/loader:1873:14) {
  [cause]: [
    Error: Cannot find module './swc.android-arm64.node'
    Require stack:
    - /data/data/com.termux/files/home/AIFT/biozone-harmony-boost/node_modules/@swc/core/binding.js
    - /data/data/com.termux/files/home/AIFT/biozone-harmony-boost/node_modules/@swc/core/index.js
        at Module._resolveFilename (node:internal/modules/cjs/loader:1519:15)
        at wrapResolveFilename (node:internal/modules/cjs/loader:1073:27)
        at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1097:10)
        at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1124:12)
        at Module._load (node:internal/modules/cjs/loader:1296:5)
        at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
        at Module.require (node:internal/modules/cjs/loader:1619:12)
        at require (node:internal/modules/helpers:191:16)
        at requireNative (/data/data/com.termux/files/home/AIFT/biozone-harmony-boost/node_modules/@swc/core/binding.js:63:16)
        at Object.<anonymous> (/data/data/com.termux/files/home/AIFT/biozone-harmony-boost/node_modules/@swc/core/binding.js:318:17) {
      code: 'MODULE_NOT_FOUND',
      requireStack: [Array]
    },
    Error: Cannot find module '@swc/core-android-arm64'
    Require stack:
    - /data/data/com.termux/files/home/AIFT/biozone-harmony-boost/node_modules/@swc/core/binding.js
    - /data/data/com.termux/files/home/AIFT/biozone-harmony-boost/node_modules/@swc/core/index.js
        at Module._resolveFilename (node:internal/modules/cjs/loader:1519:15)
        at wrapResolveFilename (node:internal/modules/cjs/loader:1073:27)
        at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1097:10)
        at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1124:12)
        at Module._load (node:internal/modules/cjs/loader:1296:5)
        at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
        at Module.require (node:internal/modules/cjs/loader:1619:12)
        at require (node:internal/modules/helpers:191:16)
        at requireNative (/data/data/com.termux/files/home/AIFT/biozone-harmony-boost/node_modules/@swc/core/binding.js:68:16)
        at Object.<anonymous> (/data/data/com.termux/files/home/AIFT/biozone-harmony-boost/node_modules/@swc/core/binding.js:318:17) {
      code: 'MODULE_NOT_FOUND',
      requireStack: [Array]
    }
  ]
}
```
