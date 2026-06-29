# AIFT Federation Diagnostic Bundle

Generated: Sat Jun 27 17:58:40 PDT 2026

## Environment
```text
DATE=Sat Jun 27 17:49:33 PDT 2026
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

## Repositories
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

## Repo Summaries

---
# .github

## Git
```text
?? .aift/
?? docs/
```

## Detected Commands
```text
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

---
# AI-Freedom-Trust

## Git
```text
```

## Detected Commands
```text
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

---
# AIFT-Forge

## Git
```text
```

## Detected Commands
```text
npm:lint=npm run lint
npm:test=npm run test
npm:typecheck=npm run typecheck
npm:build=npm run build
npm:deps:doctor=npm run deps:doctor
npm:provider:smoke=npm run provider:smoke
npm:pipeline=npm run pipeline
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

### npm-build.log
```text
REPO=AIFT-Forge
LABEL=npm-build
CMD=npm ci && npm run build
DATE=Sat Jun 27 17:49:46 PDT 2026

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

### npm-test.log
```text
REPO=AIFT-Forge
LABEL=npm-test
CMD=npm ci && npm test
DATE=Sat Jun 27 17:49:36 PDT 2026

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

> aift-root-apps@0.1.0 test
> vitest run --passWithNoTests


 RUN  v2.1.9 /data/data/com.termux/files/home/AIFT/AIFT-Forge

✔ wallet-core exports foundation object (6.632083ms)
 ✓ packages/wallet-core/tests/index.test.mjs (0 test) 2ms
 ✓ tests/git-smart-http-access.test.mjs (4 tests) 44ms

 Test Files  2 passed (2)
      Tests  4 passed (4)
   Start at  17:49:44
   Duration  1.73s (transform 276ms, setup 0ms, collect 417ms, tests 46ms, environment 2ms, prepare 583ms)

```

---
# AIFT-Genesis

## Git
```text
```

## Detected Commands
```text
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

---
# AIFT-OS

## Git
```text
?? ci-truth-repair-report.sh
?? phase-16-diagnostics.sh
?? reports/diagnostics/
```

## Detected Commands
```text
go:test=go test ./...
go:build=go build ./...
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

### go-build.log
```text
REPO=AIFT-OS
LABEL=go-build
CMD=go build ./...
DATE=Sat Jun 27 17:49:55 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found
```

### go-test.log
```text
REPO=AIFT-OS
LABEL=go-test
CMD=go test ./...
DATE=Sat Jun 27 17:49:54 PDT 2026

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

## Git
```text
```

## Detected Commands
```text
npm:test=npm run test
npm:lint=npm run lint
npm:qa:local=npm run qa:local
npm:verify:structure=npm run verify:structure
npm:dev=npm run dev
npm:build=npm run build
npm:start=npm run start
npm:check=npm run check
npm:db:push=npm run db:push
npm:security:audit=npm run security:audit
npm:security:deps=npm run security:deps
npm:security:fix=npm run security:fix
npm:security:api-keys=npm run security:api-keys
npm:security:install-hooks=npm run security:install-hooks
npm:eval=npm run eval
npm:eval:custom=npm run eval:custom
npm:eval:ai=npm run eval:ai
aift:status.sh=sh .aift/commands/status.sh
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
npm error A complete log of this run can be found in: /data/data/com.termux/files/home/.npm/_logs/2026-06-28T00_52_57_121Z-debug-0.log
```

### npm-test.log
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
npm error A complete log of this run can be found in: /data/data/com.termux/files/home/.npm/_logs/2026-06-28T00_50_00_492Z-debug-0.log
```

---
# AetherianGovernance

## Git
```text
```

## Detected Commands
```text
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

---
# BookSmith-Federation-OS

## Git
```text
```

## Detected Commands
```text
npm:dev=npm run dev
npm:build=npm run build
npm:typecheck=npm run typecheck
npm:lint=npm run lint
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

### npm-build.log
```text
REPO=BookSmith-Federation-OS
LABEL=npm-build
CMD=npm ci && npm run build
DATE=Sat Jun 27 17:55:03 PDT 2026

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
npm error A complete log of this run can be found in: /data/data/com.termux/files/home/.npm/_logs/2026-06-28T00_55_04_421Z-debug-0.log
```

---
# OpenMontage

## Git
```text
```

## Detected Commands
```text
python:pytest=python -m pytest
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

---
# TheMindofAll

## Git
```text
```

## Detected Commands
```text
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

---
# VPS

## Git
```text
```

## Detected Commands
```text
npm:test=npm run test
npm:lint=npm run lint
npm:lint:structure=npm run lint:structure
npm:qa:local=npm run qa:local
npm:verify:structure=npm run verify:structure
npm:verify:installer-registry=npm run verify:installer-registry
npm:dashboard:dev=npm run dashboard:dev
npm:dashboard:build=npm run dashboard:build
npm:dashboard:start=npm run dashboard:start
npm:desktop:dev=npm run desktop:dev
npm:desktop:build:win=npm run desktop:build:win
npm:node-agent:build=npm run node-agent:build
npm:android:sync=npm run android:sync
npm:android:build=npm run android:build
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

### npm-test.log
```text
REPO=VPS
LABEL=npm-test
CMD=npm ci && npm test
DATE=Sat Jun 27 17:55:08 PDT 2026

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
npm error A complete log of this run can be found in: /data/data/com.termux/files/home/.npm/_logs/2026-06-28T00_55_08_458Z-debug-0.log
```

---
# aifreedomtrustfederation.github.io

## Git
```text
```

## Detected Commands
```text
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

---
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

---
# booksmith-ai

## Git
```text
```

## Detected Commands
```text
npm:dev=npm run dev
npm:build=npm run build
npm:start=npm run start
npm:lint=npm run lint
npm:ensure:book-folders=npm run ensure:book-folders
npm:validate:library=npm run validate:library
npm:latex:sample=npm run latex:sample
npm:latex:sample:lua=npm run latex:sample:lua
npm:fhqcm:images=npm run fhqcm:images
npm:fhqcm:polish=npm run fhqcm:polish
npm:latex:sample:tectonic=npm run latex:sample:tectonic
npm:render:latex=npm run render:latex
npm:render:latex:fhqcm=npm run render:latex:fhqcm
npm:proof:report=npm run proof:report
npm:proof:report:fhqcm=npm run proof:report:fhqcm
npm:validate:latex-integrity=npm run validate:latex-integrity
npm:validate:latex-integrity:fhqcm=npm run validate:latex-integrity:fhqcm
npm:quality:gate=npm run quality:gate
npm:quality:gate:fhqcm=npm run quality:gate:fhqcm
npm:packet:build=npm run packet:build
npm:packet:build:fhqcm=npm run packet:build:fhqcm
npm:phase1:pipeline=npm run phase1:pipeline
npm:phase1:pipeline:fhqcm=npm run phase1:pipeline:fhqcm
npm:proof:autofix=npm run proof:autofix
npm:proof:autofix:fhqcm=npm run proof:autofix:fhqcm
npm:proof:limit=npm run proof:limit
npm:proof:limit:fhqcm=npm run proof:limit:fhqcm
npm:phase1:strict-proof=npm run phase1:strict-proof
npm:phase1:strict-proof:fhqcm=npm run phase1:strict-proof:fhqcm
npm:zip:latest=npm run zip:latest
npm:zip:latest:fhqcm=npm run zip:latest:fhqcm
npm:repair:proof=npm run repair:proof
npm:repair:proof:fhqcm=npm run repair:proof:fhqcm
npm:repair:queue=npm run repair:queue
npm:repair:queue:fhqcm=npm run repair:queue:fhqcm
npm:proof:v2=npm run proof:v2
npm:proof:v2:fhqcm=npm run proof:v2:fhqcm
npm:proof:repair:v2=npm run proof:repair:v2
npm:proof:repair:v2:fhqcm=npm run proof:repair:v2:fhqcm
npm:phase1:repair:v2=npm run phase1:repair:v2
npm:phase1:repair:v2:fhqcm=npm run phase1:repair:v2:fhqcm
npm:json:repair=npm run json:repair
npm:json:validate=npm run json:validate
npm:latex:structure=npm run latex:structure
npm:latex:structure:fhqcm=npm run latex:structure:fhqcm
npm:typography:v3=npm run typography:v3
npm:typography:v3:fhqcm=npm run typography:v3:fhqcm
npm:phase1:typography:v3=npm run phase1:typography:v3
npm:phase1:typography:v3:fhqcm=npm run phase1:typography:v3:fhqcm
npm:proof:visual=npm run proof:visual
npm:proof:visual:fhqcm=npm run proof:visual:fhqcm
npm:phase1:visual-proof=npm run phase1:visual-proof
npm:phase1:visual-proof:fhqcm=npm run phase1:visual-proof:fhqcm
npm:source-map:latex=npm run source-map:latex
npm:source-map:latex:fhqcm=npm run source-map:latex:fhqcm
npm:source-map:diagnostics=npm run source-map:diagnostics
npm:source-map:diagnostics:fhqcm=npm run source-map:diagnostics:fhqcm
npm:phase1:source-mapped-proof=npm run phase1:source-mapped-proof
npm:phase1:source-mapped-proof:fhqcm=npm run phase1:source-mapped-proof:fhqcm
npm:proof:v3:context=npm run proof:v3:context
npm:proof:v3:context:fhqcm=npm run proof:v3:context:fhqcm
npm:proof:v3:plan=npm run proof:v3:plan
npm:proof:v3:plan:fhqcm=npm run proof:v3:plan:fhqcm
npm:proof:v3:apply-one=npm run proof:v3:apply-one
npm:proof:v3=npm run proof:v3
npm:proof:v3:fhqcm=npm run proof:v3:fhqcm
npm:typography:diagnostics:v4=npm run typography:diagnostics:v4
npm:typography:diagnostics:v4:fhqcm=npm run typography:diagnostics:v4:fhqcm
npm:phase1:typography-diagnostics:v4=npm run phase1:typography-diagnostics:v4
npm:phase1:typography-diagnostics:v4:fhqcm=npm run phase1:typography-diagnostics:v4:fhqcm
npm:proof:context=npm run proof:context
npm:proof:context:fhqcm=npm run proof:context:fhqcm
npm:proof:architecture-check=npm run proof:architecture-check
npm:proof:architecture-check:fhqcm=npm run proof:architecture-check:fhqcm
npm:phase1:architecture:v4=npm run phase1:architecture:v4
npm:phase1:architecture:v4:fhqcm=npm run phase1:architecture:v4:fhqcm
npm:proof:semantic-context:v4=npm run proof:semantic-context:v4
npm:proof:semantic-context:v4:fhqcm=npm run proof:semantic-context:v4:fhqcm
npm:proof:semantic-plan:v4=npm run proof:semantic-plan:v4
npm:proof:semantic-plan:v4:fhqcm=npm run proof:semantic-plan:v4:fhqcm
npm:proof:semantic-apply:v4=npm run proof:semantic-apply:v4
npm:phase1:semantic:v4=npm run phase1:semantic:v4
npm:phase1:semantic:v4:fhqcm=npm run phase1:semantic:v4:fhqcm
npm:semantic:blocks=npm run semantic:blocks
npm:semantic:blocks:fhqcm=npm run semantic:blocks:fhqcm
npm:phase1:semantic-blocks:v1=npm run phase1:semantic-blocks:v1
npm:phase1:semantic-blocks:v1:fhqcm=npm run phase1:semantic-blocks:v1:fhqcm
npm:figure:audit=npm run figure:audit
npm:figure:audit:fhqcm=npm run figure:audit:fhqcm
npm:phase1:figure-audit:v1=npm run phase1:figure-audit:v1
npm:phase1:figure-audit:v1:fhqcm=npm run phase1:figure-audit:v1:fhqcm
npm:figures:registry=npm run figures:registry
npm:figures:registry:fhqcm=npm run figures:registry:fhqcm
npm:registry:build=npm run registry:build
npm:validate:library:v2=npm run validate:library:v2
npm:publication:gate:v2=npm run publication:gate:v2
npm:publication:gate:v2:fhqcm=npm run publication:gate:v2:fhqcm
npm:booksmith:stabilize-submit=npm run booksmith:stabilize-submit
npm:booksmith:stabilize-submit:fhqcm=npm run booksmith:stabilize-submit:fhqcm
npm:bibliography:audit:v2=npm run bibliography:audit:v2
npm:bibliography:audit:v2:fhqcm=npm run bibliography:audit:v2:fhqcm
npm:bibliography:stub:v2=npm run bibliography:stub:v2
npm:bibliography:stub:v2:fhqcm=npm run bibliography:stub:v2:fhqcm
npm:bibliography:pipeline:v2=npm run bibliography:pipeline:v2
npm:bibliography:pipeline:v2:fhqcm=npm run bibliography:pipeline:v2:fhqcm
npm:reference:intelligence:v1=npm run reference:intelligence:v1
npm:reference:intelligence:v1:fhqcm=npm run reference:intelligence:v1:fhqcm
npm:figure:intelligence:v1=npm run figure:intelligence:v1
npm:figure:intelligence:v1:fhqcm=npm run figure:intelligence:v1:fhqcm
npm:publication:engine:v1=npm run publication:engine:v1
npm:publication:engine:v1:fhqcm=npm run publication:engine:v1:fhqcm
npm:booksmith:publish=npm run booksmith:publish
npm:booksmith:publish:fhqcm=npm run booksmith:publish:fhqcm
npm:artifacts:manage:v1=npm run artifacts:manage:v1
npm:artifacts:manage:v1:fhqcm=npm run artifacts:manage:v1:fhqcm
npm:booksmith:publish-managed=npm run booksmith:publish-managed
npm:booksmith:publish-managed:fhqcm=npm run booksmith:publish-managed:fhqcm
npm:booksmith:os=npm run booksmith:os
npm:booksmith:os:fhqcm=npm run booksmith:os:fhqcm
npm:booksmith:os:submit:fhqcm=npm run booksmith:os:submit:fhqcm
npm:figure:engine=npm run figure:engine
npm:figure:queue=npm run figure:queue
npm:figure:import=npm run figure:import
npm:figure:approve=npm run figure:approve
npm:figure:reject=npm run figure:reject
npm:figure:pipeline:v1=npm run figure:pipeline:v1
npm:figure:pipeline:v1:fhqcm=npm run figure:pipeline:v1:fhqcm
npm:figure:studio=npm run figure:studio
npm:figure:studio:fhqcm=npm run figure:studio:fhqcm
npm:figure:revise=npm run figure:revise
npm:figure:studio:pipeline=npm run figure:studio:pipeline
npm:figure:studio:pipeline:fhqcm=npm run figure:studio:pipeline:fhqcm
npm:figure:art-director=npm run figure:art-director
npm:figure:art-director:fhqcm=npm run figure:art-director:fhqcm
npm:figure:art-director:pipeline=npm run figure:art-director:pipeline
npm:figure:art-director:pipeline:fhqcm=npm run figure:art-director:pipeline:fhqcm
npm:studio:figures:prepare=npm run studio:figures:prepare
npm:studio:dev=npm run studio:dev
npm:system:health=npm run system:health
npm:studio=npm run studio
npm:bootstrap=npm run bootstrap
npm:server=npm run server
npm:install:termux=npm run install:termux
npm:server:fhqcm=npm run server:fhqcm
npm:install:missing-tools=npm run install:missing-tools
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

### npm-build.log
```text
REPO=booksmith-ai
LABEL=npm-build
CMD=npm ci && npm run build
DATE=Sat Jun 27 17:55:44 PDT 2026

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

  Downloading swc package @next/swc-wasm-nodejs... to /data/data/com.termux/files/home/.cache/next-swc
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

---
# capital-city-provisions

## Git
```text
```

## Detected Commands
```text
npm:dev=npm run dev
npm:build=npm run build
npm:start=npm run start
npm:lint=npm run lint
npm:typecheck=npm run typecheck
npm:license:audit=npm run license:audit
npm:verify=npm run verify
npm:vercel-build=npm run vercel-build
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

### npm-verify.log
```text
REPO=capital-city-provisions
LABEL=npm-verify
CMD=npm ci && npm run verify
DATE=Sat Jun 27 17:57:12 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found

added 406 packages in 37s
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

## Git
```text
```

## Detected Commands
```text
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

---
# mobox

## Git
```text
```

## Detected Commands
```text
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

---
# repo-brainstorm-backend-forge

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
REPO=repo-brainstorm-backend-forge
LABEL=npm-build
CMD=npm ci && npm run build
DATE=Sat Jun 27 17:58:07 PDT 2026

/data/data/com.termux/files/usr/bin/sh: 3: source: not found

added 395 packages, and audited 396 packages in 28s

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

## Git
```text
```

## Detected Commands
```text
npm:dev=npm run dev
npm:build=npm run build
npm:start=npm run start
npm:lint=npm run lint
npm:typegen=npm run typegen
npm:typecheck=npm run typecheck
npm:test=npm run test
npm:test:e2e=npm run test:e2e
npm:knowledge:index=npm run knowledge:index
npm:verify=npm run verify
aift:status.sh=sh .aift/commands/status.sh
```

## Logs

### npm-verify.log
```text
REPO=tastycutz
LABEL=npm-verify
CMD=npm ci && npm run verify
DATE=Sat Jun 27 17:58:37 PDT 2026

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
npm error A complete log of this run can be found in: /data/data/com.termux/files/home/.npm/_logs/2026-06-28T00_58_38_105Z-debug-0.log
```

---
# www.aifreedomtrust.com

## Git
```text
```

## Detected Commands
```text
aift:status.sh=sh .aift/commands/status.sh
```

## Logs
