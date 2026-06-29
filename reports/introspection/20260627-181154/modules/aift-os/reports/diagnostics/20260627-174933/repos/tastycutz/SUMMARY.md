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
