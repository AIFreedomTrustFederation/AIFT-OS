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
