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
