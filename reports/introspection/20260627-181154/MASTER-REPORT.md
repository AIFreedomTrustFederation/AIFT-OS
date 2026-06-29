# AIFT Federation Discovery & Introspection Kernel Report

Generated: Sat Jun 27 18:18:47 PDT 2026

## Environment
```text
DATE=Sat Jun 27 18:11:54 PDT 2026
ROOT=/data/data/com.termux/files/home/AIFT
OS=/data/data/com.termux/files/home/AIFT/AIFT-OS
SHELL=/data/data/com.termux/files/usr/bin/bash
PATH=/data/data/com.termux/files/home/.local/bin:/data/data/com.termux/files/home/.local/bin:/data/data/com.termux/files/home/.local/bin:/data/data/com.termux/files/home/.local/bin:/data/data/com.termux/files/usr/bin:/product/bin:/apex/com.android.runtime/bin:/apex/com.android.art/bin:/system_ext/bin:/system/bin:/system/xbin:/odm/bin:/vendor/bin:/vendor/xbin
NODE=/data/data/com.termux/files/usr/bin/node v26.3.1
NPM=/data/data/com.termux/files/usr/bin/npm 11.17.0
GO=/data/data/com.termux/files/usr/bin/go go version go1.26.4 android/arm64
GIT=git version 2.54.0
UNAME=Linux localhost 4.19.113-27223811 #1 SMP PREEMPT Fri Sep 26 20:41:10 KST 2025 aarch64 Android
```

## Discovered Repositories
- .github
- AI-Freedom-Trust
- AIFT-Forge
- AIFT-Genesis
- AIFT-OS
- Aether_Coin_biozonecurrency
- AetherianGovernance
- BookSmith-Federation-OS
- OpenMontage
- TheMindofAll
- VPS
- aifreedomtrustfederation.github.io
- biozone-harmony-boost
- booksmith-ai
- c-848263
- capital-city-provisions
- chktex
- mobox
- repo-brainstorm-backend-forge
- tastycutz
- www.aifreedomtrust.com

## Repository Summaries

---
# .github

## Git Status
```text
?? .aift/
?? docs/
```

## Detected Commands
```text
```

## Logs

---
# AI-Freedom-Trust

## Git Status
```text
```

## Detected Commands
```text
```

## Logs

---
# AIFT-Forge

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
DATE=Sat Jun 27 18:11:57 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found

added 126 packages, and audited 127 packages in 6s

35 packages are looking for funding
  run `npm fund` for details

5 vulnerabilities (3 moderate, 1 high, 1 critical)

To address all issues (including breaking changes), run:
  npm audit fix --force

Run `npm audit` for details.
npm warn allow-scripts 1 package has install scripts not yet covered by allowScripts:
npm warn allow-scripts   esbuild@0.21.5 (postinstall: node install.js)
npm warn allow-scripts
npm warn allow-scripts Run `npm approve-scripts --allow-scripts-pending` to review, or `npm approve-scripts <pkg>` to allow.

> aift-root-apps@0.1.0 build
> node scripts/aift-build-check.mjs

✅ Android/Termux build profile passed.
```

---
# AIFT-Genesis

## Git Status
```text
```

## Detected Commands
```text
```

## Logs

---
# AIFT-OS

## Git Status
```text
?? AI-Code-Training/scripts/phase-scripts/phase-16-introspection-kernel.sh
?? ci-truth-repair-report.sh
?? fix-introspection-and-push-federation.sh
?? phase-16-diagnostics.sh
?? phase-16-introspection-kernel.sh
?? reports/diagnostics-20260627-174933.tar.gz
?? reports/diagnostics/
?? reports/introspection/
```

## Detected Commands
```text
go:test=go test ./...
go:build=go build ./...
```

## Logs

### go-build.log
```text
LABEL=go-build
CMD=go build ./...
DATE=Sat Jun 27 18:12:07 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found
```

### go-test.log
```text
LABEL=go-test
CMD=go test ./...
DATE=Sat Jun 27 18:12:05 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found
?   	github.com/AIFreedomTrustFederation/AIFT-OS/cmd/aift	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/api	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/capabilities	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/config	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/daemon	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/doctor	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/eventmesh	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/events	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/federation	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/gitx	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/graph	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/intelligence	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/jobs	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernel	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/manual	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/planner	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/plugins	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/providers	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/repo	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/runtime	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/scheduler	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/servicecontracts	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/services	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/state	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/supervisor	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/sync	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/version	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/workflow	[no test files]
?   	github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace	[no test files]
```

---
# Aether_Coin_biozonecurrency

## Git Status
```text
```

## Detected Commands
```text

















```

## Logs

### npm-build.log
```text
npm error Missing: @esbuild/win32-ia32@0.28.1 from lock file
npm error Missing: @esbuild/win32-x64@0.28.1 from lock file
npm error Missing: uuid@9.0.1 from lock file
npm error Missing: @walletconnect/sign-client@2.21.0 from lock file
npm error Missing: @walletconnect/types@2.21.0 from lock file
npm error Missing: @walletconnect/utils@2.21.0 from lock file
npm error Missing: @walletconnect/core@2.21.0 from lock file
npm error Missing: viem@2.23.2 from lock file
npm error Missing: @walletconnect/sign-client@2.21.0 from lock file
npm error Missing: @walletconnect/types@2.21.0 from lock file
npm error Missing: @walletconnect/utils@2.21.0 from lock file
npm error Missing: @walletconnect/core@2.21.0 from lock file
npm error Missing: viem@2.23.2 from lock file
npm error Missing: @walletconnect/sign-client@2.21.0 from lock file
npm error Missing: @walletconnect/utils@2.21.0 from lock file
npm error Missing: @walletconnect/core@2.21.0 from lock file
npm error Missing: viem@2.23.2 from lock file
npm error Missing: @scure/bip32@1.6.2 from lock file
npm error Missing: @scure/bip39@1.5.4 from lock file
npm error Missing: abitype@1.0.8 from lock file
npm error Missing: isows@1.0.6 from lock file
npm error Missing: ox@0.6.7 from lock file
npm error Missing: ws@8.18.0 from lock file
npm error Missing: uint8arrays@6.1.1 from lock file
npm error Missing: interface-store@8.0.0 from lock file
npm error Missing: uint8arrays@6.1.1 from lock file
npm error Missing: uint8arraylist@3.0.2 from lock file
npm error
npm error Clean install a project
npm error
npm error Usage:
npm error npm ci
npm error
npm error Options:
npm error [--install-strategy <hoisted|nested|shallow|linked>] [--legacy-bundling]
npm error [--global-style] [--omit <dev|optional|peer> [--omit <dev|optional|peer> ...]]
npm error [--include <prod|dev|optional|peer> [--include <prod|dev|optional|peer> ...]]
npm error [--strict-peer-deps] [--foreground-scripts] [--ignore-scripts]
npm error [--allow-directory <all|none|root>] [--allow-file <all|none|root>]
npm error [--allow-git <all|none|root>] [--allow-remote <all|none|root>]
npm error [--allow-scripts <package-list> [--allow-scripts <package-list> ...]]
npm error [--strict-allow-scripts] [--dangerously-allow-all-scripts] [--no-audit]
npm error [--no-bin-links] [--no-fund] [--dry-run]
npm error [-w|--workspace <workspace-name> [-w|--workspace <workspace-name> ...]]
npm error [--workspaces] [--include-workspace-root] [--install-links]
npm error
npm error   --install-strategy
npm error     Sets the strategy for installing packages in node_modules.
npm error
npm error   --legacy-bundling
npm error     Instead of hoisting package installs in `node_modules`, install packages
npm error
npm error   --global-style
npm error     Only install direct dependencies in the top level `node_modules`,
npm error
npm error   --omit
npm error     Dependency types to omit from the installation tree on disk.
npm error
npm error   --include
npm error     Option that allows for defining which types of dependencies to install.
npm error
npm error   --strict-peer-deps
npm error     If set to `true`, and `--legacy-peer-deps` is not set, then _any_
npm error
npm error   --foreground-scripts
npm error     Run all build scripts (ie, `preinstall`, `install`, and
npm error
npm error   --ignore-scripts
npm error     If true, npm does not run scripts specified in package.json files.
npm error
npm error   --allow-directory
npm error     Limits the ability for npm to install dependencies from directories.
npm error
npm error   --allow-file
npm error     Limits the ability for npm to install dependencies from tarball files.
npm error
npm error   --allow-git
npm error     Limits the ability for npm to fetch dependencies from git references.
npm error
npm error   --allow-remote
npm error     Limits the ability for npm to fetch dependencies from urls.
npm error
npm error   --allow-scripts
npm error     Comma-separated list of packages whose install-time lifecycle scripts
npm error
npm error   --strict-allow-scripts
npm error     If `true`, turn the install-script policy from a warning into a hard
npm error
npm error   --dangerously-allow-all-scripts
npm error     If `true`, bypass the `allowScripts` policy entirely and run every
npm error
npm error   --audit
npm error     When "true" submit audit reports alongside the current npm command to the
npm error
npm error   --bin-links
npm error     Tells npm to create symlinks (or `.cmd` shims on Windows) for package
npm error
npm error   --fund
npm error     When "true" displays the message at the end of each `npm install`
npm error
npm error   --dry-run
npm error     Indicates that you don't want npm to make any changes and that it should
npm error
npm error   -w|--workspace
npm error     Enable running a command in the context of the configured workspaces of the
npm error
npm error   --workspaces
npm error     Set to true to run the command in the context of **all** configured
npm error
npm error   --include-workspace-root
npm error     Include the workspace root when workspaces are enabled for a command.
npm error
npm error   --install-links
npm error     When set file: protocol dependencies will be packed and installed as
npm error
npm error
npm error aliases: clean-install, ic, install-clean, isntall-clean
npm error
npm error Run "npm help ci" for more info
npm error A complete log of this run can be found in: /data/data/com.termux/files/home/.npm/_logs/2026-06-28T01_12_11_587Z-debug-0.log
```

---
# AetherianGovernance

## Git Status
```text
```

## Detected Commands
```text
```

## Logs

---
# BookSmith-Federation-OS

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
DATE=Sat Jun 27 18:15:00 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found
npm error code EUSAGE
npm error
npm error The `npm ci` command can only install with an existing package-lock.json or
npm error npm-shrinkwrap.json with lockfileVersion >= 1. Run an install with npm@5 or
npm error later to generate a package-lock.json file, then try again.
npm error
npm error Clean install a project
npm error
npm error Usage:
npm error npm ci
npm error
npm error Options:
npm error [--install-strategy <hoisted|nested|shallow|linked>] [--legacy-bundling]
npm error [--global-style] [--omit <dev|optional|peer> [--omit <dev|optional|peer> ...]]
npm error [--include <prod|dev|optional|peer> [--include <prod|dev|optional|peer> ...]]
npm error [--strict-peer-deps] [--foreground-scripts] [--ignore-scripts]
npm error [--allow-directory <all|none|root>] [--allow-file <all|none|root>]
npm error [--allow-git <all|none|root>] [--allow-remote <all|none|root>]
npm error [--allow-scripts <package-list> [--allow-scripts <package-list> ...]]
npm error [--strict-allow-scripts] [--dangerously-allow-all-scripts] [--no-audit]
npm error [--no-bin-links] [--no-fund] [--dry-run]
npm error [-w|--workspace <workspace-name> [-w|--workspace <workspace-name> ...]]
npm error [--workspaces] [--include-workspace-root] [--install-links]
npm error
npm error   --install-strategy
npm error     Sets the strategy for installing packages in node_modules.
npm error
npm error   --legacy-bundling
npm error     Instead of hoisting package installs in `node_modules`, install packages
npm error
npm error   --global-style
npm error     Only install direct dependencies in the top level `node_modules`,
npm error
npm error   --omit
npm error     Dependency types to omit from the installation tree on disk.
npm error
npm error   --include
npm error     Option that allows for defining which types of dependencies to install.
npm error
npm error   --strict-peer-deps
npm error     If set to `true`, and `--legacy-peer-deps` is not set, then _any_
npm error
npm error   --foreground-scripts
npm error     Run all build scripts (ie, `preinstall`, `install`, and
npm error
npm error   --ignore-scripts
npm error     If true, npm does not run scripts specified in package.json files.
npm error
npm error   --allow-directory
npm error     Limits the ability for npm to install dependencies from directories.
npm error
npm error   --allow-file
npm error     Limits the ability for npm to install dependencies from tarball files.
npm error
npm error   --allow-git
npm error     Limits the ability for npm to fetch dependencies from git references.
npm error
npm error   --allow-remote
npm error     Limits the ability for npm to fetch dependencies from urls.
npm error
npm error   --allow-scripts
npm error     Comma-separated list of packages whose install-time lifecycle scripts
npm error
npm error   --strict-allow-scripts
npm error     If `true`, turn the install-script policy from a warning into a hard
npm error
npm error   --dangerously-allow-all-scripts
npm error     If `true`, bypass the `allowScripts` policy entirely and run every
npm error
npm error   --audit
npm error     When "true" submit audit reports alongside the current npm command to the
npm error
npm error   --bin-links
npm error     Tells npm to create symlinks (or `.cmd` shims on Windows) for package
npm error
npm error   --fund
npm error     When "true" displays the message at the end of each `npm install`
npm error
npm error   --dry-run
npm error     Indicates that you don't want npm to make any changes and that it should
npm error
npm error   -w|--workspace
npm error     Enable running a command in the context of the configured workspaces of the
npm error
npm error   --workspaces
npm error     Set to true to run the command in the context of **all** configured
npm error
npm error   --include-workspace-root
npm error     Include the workspace root when workspaces are enabled for a command.
npm error
npm error   --install-links
npm error     When set file: protocol dependencies will be packed and installed as
npm error
npm error
npm error aliases: clean-install, ic, install-clean, isntall-clean
npm error
npm error Run "npm help ci" for more info
npm error A complete log of this run can be found in: /data/data/com.termux/files/home/.npm/_logs/2026-06-28T01_15_01_356Z-debug-0.log
```

---
# OpenMontage

## Git Status
```text
```

## Detected Commands
```text
```

## Logs

---
# TheMindofAll

## Git Status
```text
```

## Detected Commands
```text
```

## Logs

---
# VPS

## Git Status
```text
```

## Detected Commands
```text














```

## Logs

---
# aifreedomtrustfederation.github.io

## Git Status
```text
```

## Detected Commands
```text
```

## Logs

---
# biozone-harmony-boost

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
DATE=Sat Jun 27 18:15:06 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found
npm warn deprecated glob@10.5.0: Old versions of glob are not supported, and contain widely publicized security vulnerabilities, which have been fixed in the current version. Please update. Support for old versions may be purchased (at exorbitant rates) by contacting i@izs.me

added 453 packages, and audited 454 packages in 30s

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

---
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

---
# c-848263

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
DATE=Sat Jun 27 18:16:29 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found

added 421 packages, and audited 422 packages in 36s

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

---
# capital-city-provisions

## Git Status
```text
```

## Detected Commands
```text








```

## Logs

### npm-verify.log
```text
LABEL=npm-verify
CMD=npm ci && npm run verify
DATE=Sat Jun 27 18:17:09 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found

added 406 packages in 43s
npm warn allow-scripts 2 packages have install scripts not yet covered by allowScripts:
npm warn allow-scripts   sharp@0.34.5 (install: node install/check.js || npm run build)
npm warn allow-scripts   unrs-resolver@1.12.2 (postinstall: node postinstall.js)
npm warn allow-scripts
npm warn allow-scripts Run `npm approve-scripts --allow-scripts-pending` to review, or `npm approve-scripts <pkg>` to allow.

> capital-city-provisions@0.1.0 verify
> npm run typecheck && npm run lint && npm run license:audit && npm run build


> capital-city-provisions@0.1.0 typecheck
> tsc --noEmit --incremental false

app/premium-meats/page.tsx(3,20): error TS2307: Cannot find module './PremiumMeatsHome.module.css' or its corresponding type declarations.
```

---
# chktex

## Git Status
```text
```

## Detected Commands
```text
```

## Logs

---
# mobox

## Git Status
```text
```

## Detected Commands
```text
```

## Logs

---
# repo-brainstorm-backend-forge

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
DATE=Sat Jun 27 18:18:09 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found

added 395 packages, and audited 396 packages in 31s

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

failed to load config from /data/data/com.termux/files/home/AIFT/repo-brainstorm-backend-forge/vite.config.ts
error during build:
Error: Failed to load native binding
    at Object.<anonymous> (/data/data/com.termux/files/home/AIFT/repo-brainstorm-backend-forge/node_modules/@swc/core/binding.js:333:11)
    at Module._compile (node:internal/modules/cjs/loader:1873:14)
    at Module._extensions..js (node:internal/modules/cjs/loader:2013:10)
    at Module.load (node:internal/modules/cjs/loader:1596:32)
    at Module._load (node:internal/modules/cjs/loader:1398:12)
    at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
    at Module.require (node:internal/modules/cjs/loader:1619:12)
    at require (node:internal/modules/helpers:191:16)
    at Object.<anonymous> (/data/data/com.termux/files/home/AIFT/repo-brainstorm-backend-forge/node_modules/@swc/core/index.js:49:17)
    at Module._compile (node:internal/modules/cjs/loader:1873:14)
```

---
# tastycutz

## Git Status
```text
```

## Detected Commands
```text










```

## Logs

### npm-verify.log
```text
LABEL=npm-verify
CMD=npm ci && npm run verify
DATE=Sat Jun 27 18:18:44 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found
npm warn EBADENGINE Unsupported engine {
npm warn EBADENGINE   package: '@renovatebot/pep440@4.2.1',
npm warn EBADENGINE   required: { node: '^20.9.0 || ^22.11.0 || ^24', pnpm: '^10.0.0' },
npm warn EBADENGINE   current: { node: 'v26.3.1', npm: '11.17.0' }
npm warn EBADENGINE }
npm error code EBADPLATFORM
npm error notsup Unsupported platform for onnxruntime-node@1.24.3: wanted {"os":"win32,darwin,linux"} (current: {"os":"android"})
npm error notsup Valid os:  win32,darwin,linux
npm error notsup Actual os: android
npm error A complete log of this run can be found in: /data/data/com.termux/files/home/.npm/_logs/2026-06-28T01_18_44_760Z-debug-0.log
```

---
# www.aifreedomtrust.com

## Git Status
```text
```

## Detected Commands
```text
```

## Logs
