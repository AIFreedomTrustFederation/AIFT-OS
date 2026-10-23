# AIFT-OS — The Truthful Federation Control Plane

**The operating-system layer of the AI Freedom Trust Federation: federation control plane, discovery, runtime registry, intelligence, dashboards, scheduling, evidence, and orchestration.**

| Federation metadata | Value |
| --- | --- |
| Layer | `operating-system` |
| Role | federation control plane, runtime, registry, intelligence, dashboards, and orchestration |
| Primary implementation | Go with supporting shell/runtime tooling |
| Core principle | truth before automation |
| Relationship to repositories | discovers and coordinates sovereign repositories; does not assume or absorb them |

AIFT-OS exists because federation-wide automation becomes dangerous when the control plane begins with assumptions. Before it schedules, patches, reports, or coordinates anything, it must discover what actually exists: which repositories are present, what metadata they declare, which runtimes and commands are available, what state is verified, what is blocked, and what evidence supports the decision.

The constitutional reason for that design comes from the [One Eternal Scroll of ALO'ha](https://aifreedomtrustfederation.github.io/AI-Freedom-Trust/docs/pdf/one-eternal-scroll-of-aloha.pdf). Its operational discipline is [SOP-ALOHA-001](https://github.com/AIFreedomTrustFederation/AI-Freedom-Trust/blob/main/SOP-ALOHA-001.md).

---

## Book I — Truth Before Automation

AIFT-OS treats truthfulness as an operating-system feature. A capability that is planned is not reported as ready. A repository that cannot be found is not invented. A failed command does not become a green status because the desired state was obvious. The operating system begins with evidence and builds its runtime model from that evidence.

This is the technical expression of ALO'ha inside the control plane: relationship begins by seeing what is actually present. A sovereign repository is not merely a row in a central database. It has its own Git history, metadata, commands, dependencies, maintainers, risks, and purpose. The OS may coordinate that repository because it can inspect and name those facts, not because it has erased the repository into a generic federation object.

### Illuminated passage — the circuit of discovery and return

![Harmonic Krystal Torus](https://raw.githubusercontent.com/AIFreedomTrustFederation/AI-Freedom-Trust/main/docs/images/aetherion/harmonic-krystal-torus.png)

For AIFT-OS the torus is an execution diagram: observation moves inward to a runtime model; planning moves outward toward action; verification closes the circuit; reporting returns the result to human authority. The control plane is incomplete if the loop does not return.

---

## Book II — What the Operating System Owns

The operating system owns **federation awareness and orchestration**, not the domain logic of every project. Its architecture is built around discovery, registry, intelligence, readiness, scheduling, execution, evidence, graphing, and reports.

A typical path is:

```text
Federation
  ↓
Discovery
  ↓
Repository and capability registries
  ↓
Runtime intelligence
  ↓
Readiness model
  ↓
Scheduler / plan
  ↓
Approved execution
  ↓
Verification and evidence
  ↓
Report and return
```

Its integration boundaries follow the current repository metadata:

- **AIFT-Genesis → OS:** Genesis defines trust identity, constitutional genome, schemas, and inheritance patterns. The OS discovers instantiated systems; it does not invent their identity.
- **AIFT-Forge → OS:** Forge provides reusable coordination, package, build, and agent patterns. The OS determines which patterns are actually present and executable in a repository.
- **AIFT-Runtime ↔ OS:** Runtime supplies local execution, intelligence, registry, graph, status, pull/push, doctor, and verification behavior. AIFT-OS is the higher control plane that can reason across repositories.
- **VPS ↔ OS:** VPS owns nodes, deployment, relay, and server infrastructure. AIFT-OS can discover infrastructure state and schedule governed work without becoming the infrastructure provider itself.
- **Aetherion ↔ OS:** the economy layer may expose governed capabilities to the OS, but wallet, custody, transaction, and value authority remain in Aetherion and the human consent model.
- **BookSmith / TheMindofAll / portals ↔ OS:** each declares its own federation role and commands. The OS coordinates only what metadata and evidence make available.

`aift.repo.json` is therefore not decoration. It is part of the discovery grammar through which a repository tells the control plane what it is before the control plane decides what may be done with it.

---

## Book III — SOP-ALOHA-001 in AIFT-OS

The shared loop becomes a concrete operating-system lifecycle:

```text
Receive → Inspect → Name → Propose → Consent → Act → Verify → Record → Return
```

**Receive** accepts a repository, command, mission, event, or desired state. **Inspect** runs discovery rather than assumption. **Name** classifies repositories, capabilities, runtimes, dependencies, status, risk, and readiness. **Propose** builds a plan and assigns an explicit next action. **Consent** gates work that exceeds delegated authority or touches high-risk domains. **Act** invokes the real repository or runtime command rather than a simulated success. **Verify** checks the resulting state and architectural invariants. **Record** preserves evidence, reports, registry state, and learning. **Return** reports what happened, what remains blocked, and who owns the next step.

The current Go CLI exposes a smaller **implemented** command set than the full architecture envisions. From the repository root, after building `./aift` as shown below, use:

```bash
./aift help             # list actual commands and readiness labels
./aift status           # local status (doctor is an alias)
./aift apps list        # discover app manifests
./aift registry         # inspect the command registry
./aift bootstrap        # inspect local bootstrap discovery
./aift verify           # local bootstrap checks, not full make verify
make verify             # broader repository verification
go run ./tools/architecture --ci
```

**Not yet exposed by this CLI:** `aift introspect scan`, `aift capabilities scan`, and `aift runtime scan`. They describe intended control-plane capabilities, not commands implemented in `cmd/aift/main.go`. The `federation`, `repo`, and `workflow` command groups are registered as **planned** and should not be treated as operational. For supported commands and their current status, inspect `./aift registry`.

Installation follows the Go implementation:

```bash
git clone https://github.com/AIFreedomTrustFederation/AIFT-OS
cd AIFT-OS
go mod download
go build -o ./aift ./cmd/aift
./aift doctor
```

The optional `aift-os.sh` launcher resolves its checkout from the script location by default, supports `AIFT_ROOT` and `AIFT_OS_HOME` overrides (the latter takes precedence), and builds **`bin/aift` from `./cmd/aift`** on first use. It reuses that CLI executable on subsequent runs and does not detect source changes automatically. After updating CLI source, rebuild it explicitly:

```bash
mkdir -p bin
go build -o bin/aift ./cmd/aift
```

Alternatively, `make build` builds both commands into distinct paths: **`bin/aift` is the CLI**, while **`bin/aiftd` is the separate loopback MoBox UXI HTTP server**. The launcher never executes `bin/aiftd`, even when it exists.

### MoBox UXI HTTP exposure boundary

The separate `bin/aiftd` service defaults to `127.0.0.1:8787` and exposes local federation metadata, sessions, artifacts, and registered read-only adapter invocation. Its HTTP handlers do not authenticate callers; binding to loopback is a deployment boundary, not per-user authorization. In the current implementation, `requireLoopback` accepts an empty hostname, so **do not** set `AIFT_UXI_ADDR=:8787` (or use another empty-host listen address): Go's HTTP listener can then bind to non-loopback interfaces. Keep the explicit loopback address and restrict access to trusted local processes. Action decisions are recorded separately from invocation; the adapter executor rejects mutating adapters, but this does not make the HTTP API safe for network exposure. See [`cmd/aiftd/main.go`](cmd/aiftd/main.go) and [`internal/uxihttp/server.go`](internal/uxihttp/server.go). This is a documented current limitation, not an implemented authentication or network-isolation guarantee.
 The isolated launcher regression test checks this separation, first-run CLI creation, binary reuse, and environment overrides. That mock-based test does not replace a real Go build or full repository verification.

The build creates a local `./aift` executable; it does **not** automatically add `aift` to your shell's `PATH`. For the `aift ...` commands shown above, either invoke `./aift ...` from the repository root or install the executable in a directory on your `PATH`. Run `./aift --help` to inspect supported commands for the checked-out revision before assuming every documented subcommand is available.

### Termux Ubuntu PRoot verification limitation

The current [`scripts/dev.sh`](scripts/dev.sh) uses Bash process substitution (`done < <(find ... -print0)`) in `check_shell`. On some Termux `proot-distro` Ubuntu installations, the `/dev/fd` path required by process substitution is unavailable, causing `make verify` to stop at the shell-syntax check before running the remaining tests. This is an environment-specific limitation, not evidence that the Go tests failed. The hosted [CI workflow](.github/workflows/ci.yml) uses Ubuntu 24.04 GitHub-hosted runners rather than Android PRoot. For local PRoot verification, use a checkout with a separately validated compatible shell-check implementation; do not report a full `make verify` pass if the command stops at this stage. See [the verification orchestrator](scripts/dev.sh) for the current implementation.

### Workspace verification and CI prerequisites

This repository also contains a **pnpm-managed JavaScript/TypeScript workspace**, in addition to its Go implementation. The root [`package.json`](package.json) declares `pnpm@11.25.0`, rejects non-pnpm installs in its `preinstall` script, and defines `typecheck`, `lint`, `test`, and `build` scripts. Do not substitute `npm install` for the documented pnpm workflow.

**Installer guard:** The `preinstall` hook checks the package manager *before* removing alternate lockfiles. Rejected npm/Yarn attempts preserve any existing `package-lock.json` and `yarn.lock`; accepted pnpm installs retain the existing cleanup behavior. Verify this with `node --test scripts/preinstall.test.mjs`. CI runs the same regression independently of lifecycle scripts (its dependency installation uses `--ignore-scripts`).

The [CI workflow](.github/workflows/ci.yml) runs on pushes to `main`, pull requests, and manual dispatch. It uses Go from `go.mod` and Node.js 24, installs dependencies with `pnpm install --frozen-lockfile --ignore-scripts`, then runs workspace typechecking followed by the canonical `make verify` target. A local reproduction of its principal checks, with Go, Node.js 24, Corepack, and the pinned pnpm version available, is:

```bash
corepack enable
pnpm install --frozen-lockfile --ignore-scripts
pnpm run typecheck
make verify
```

`corepack enable` makes the `pnpm` executable available to package scripts that invoke `pnpm` directly. If the default shim directory is not writable, use the CI approach: `mkdir -p "$HOME/.local/bin" && corepack enable --install-directory "$HOME/.local/bin" && export PATH="$HOME/.local/bin:$PATH"` before running `pnpm`. Simply prefixing the top-level commands with `corepack` is not sufficient when nested scripts call `pnpm`.

The CLI launcher also has an SDK-free, isolated regression test that checks first-run creation of `bin/` and reuse of the built executable. Run it directly with:

```bash
bash -n aift-os.sh tests/test-aift-launcher.sh
bash tests/test-aift-launcher.sh
```

`make verify-fast` and `make verify` invoke this test through `scripts/dev.sh`. This test uses a mock `go` executable; it does not replace a real `go build` or full CI run.

CI also checks for unexpected working-tree changes and uploads coverage and architecture reports when available. **The existence of this workflow does not mean the current branch has passed CI**: consult the relevant run's actual job results and artifacts before claiming a verified build. The commands above reproduce the main verification steps, not every CI environment or checkout-cleanliness check.

### Cross-repository verification

For a best-effort check of the local AIFT workspace, run:

```bash
bash scripts/verify-all.sh "$HOME/AIFT"
```

This helper inspects Git worktrees immediately under the selected workspace and attempts available npm lint/typecheck/test/build, Cargo check, and Go tests. It continues across repositories but now returns a **nonzero exit status** if an attempted check fails, a required tool for a detected manifest is missing, or no Git repositories are found. A green result is not proof that every repository has tests: npm commands use `--if-present`, so absent scripts are skipped; other languages and nested worktrees are outside this helper's coverage. Use each repository's own documented verification command and inspect CI results before declaring a federation-wide release verified.

The truthfulness contract remains strict: no invented commands, fabricated repositories, silent failures, fake readiness, or capability claims unsupported by evidence. A planned object may be reported as planned. A missing object must remain missing. A blocked object remains blocked until the evidence changes.

---

## Book IV — A Federation That Can See Itself

AIFT-OS becomes useful when the Federation grows beyond what any person can reliably hold in working memory. The operating system maps repositories without making them dependent on hard-coded names. It gives the human operator a graph of what exists, a readiness model of what can act, and an evidence trail for why an action was proposed or performed.

That makes the OS a constitutional mechanism as much as a technical one. Sovereignty without discovery becomes isolation. Coordination without sovereignty becomes centralization. AIFT-OS is designed for the middle condition: repositories remain themselves, while the Federation becomes capable of knowing how they relate.

### The Return of the Word

In AIFT-OS, the Word returns as verified state. A request becomes inspection, inspection becomes a named model, the model becomes an approved plan, the plan becomes action, and action returns as evidence rather than assertion. The operating system speaks truthfully because it has learned to return what it can prove.
