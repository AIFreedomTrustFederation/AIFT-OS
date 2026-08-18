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

Core commands reflect that lifecycle:

```bash
# environment and system health
aift doctor

# inspect repository/federation reality
aift introspect scan

# discover available capabilities
aift capabilities scan

# evaluate runtime readiness
aift runtime scan

# verify the operating system
aift verify

# validate architecture
go run ./tools/architecture --ci
```

Installation follows the Go implementation:

```bash
git clone https://github.com/AIFreedomTrustFederation/AIFT-OS
cd AIFT-OS
go mod download
go build ./cmd/aift
```

The truthfulness contract remains strict: no invented commands, fabricated repositories, silent failures, fake readiness, or capability claims unsupported by evidence. A planned object may be reported as planned. A missing object must remain missing. A blocked object remains blocked until the evidence changes.

---

## Book IV — A Federation That Can See Itself

AIFT-OS becomes useful when the Federation grows beyond what any person can reliably hold in working memory. The operating system maps repositories without making them dependent on hard-coded names. It gives the human operator a graph of what exists, a readiness model of what can act, and an evidence trail for why an action was proposed or performed.

That makes the OS a constitutional mechanism as much as a technical one. Sovereignty without discovery becomes isolation. Coordination without sovereignty becomes centralization. AIFT-OS is designed for the middle condition: repositories remain themselves, while the Federation becomes capable of knowing how they relate.

### The Return of the Word

In AIFT-OS, the Word returns as verified state. A request becomes inspection, inspection becomes a named model, the model becomes an approved plan, the plan becomes action, and action returns as evidence rather than assertion. The operating system speaks truthfully because it has learned to return what it can prove.
