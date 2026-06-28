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
