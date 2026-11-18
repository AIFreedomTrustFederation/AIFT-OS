# Runtime and Forge producer integration

The integration runner executes the actual Runtime and Forge verification
exporters, which run their canonical source QA gates, and feeds both emitted
records into the existing offline OS consumer. It verifies explicit commit
identities and clean checkouts before either producer runs and again afterward.
It does not change production execution, collector authorization or deployment.

The dedicated [workflow](../.github/workflows/producer-consumer-integration.yml)
checks out fixed producer revisions with read-only permissions and no persisted
credentials. Forge dependencies are provisioned through locked npm installation
with lifecycle scripts disabled. Runtime LFS downloads remain disabled; source
QA checks tracked pointer text rather than loading models. Only the initial
checkout and dependency provisioning require network access; the integration
runner makes no network calls itself, while the inspected QA implementations
remain responsible for their own behavior.

From OS root, with those clean producer revisions already available:

```sh
python3 -B scripts/verify_producer_integration.py \
  --runtime-root ../AIFT-Runtime \
  --runtime-sha a713fd7807da1e308889572aca34987c98c1b6bb \
  --forge-root ../AIFT-Forge \
  --forge-sha 584781172a48c2d8ed1aaff70ae543a70f7822d6
```

The runner needs Python 3.10+, Bash, Git, Node/npm and installed Forge
dependencies. It deliberately runs repository source QA; use only reviewed
checkout revisions. It neither installs dependencies nor syncs repositories.
It stages unsigned evidence outside the repositories in a temporary directory
and removes that directory afterward. Logs go to stderr; the aggregate JSON
report goes to stdout. Redirect stdout to a chosen local evidence path if needed.
Ignored dependency/build caches can change; clean-tree checks cover tracked and
untracked non-ignored source, not all host state. Each subprocess has a ten-minute
limit, and the hosted job has a fifteen-minute limit.

Exit 0 means both producers and structural consumption completed without failed
checks. The consumer's `partial` status and unsigned provenance must remain
visible: this is no security audit, dependency provenance assessment, live model
inference, authenticated Actions-artifact collection or deployment test. Exit 1
means failed QA or rejected evidence. Exit 2 means preflight, execution,
publication, timeout or post-run checkout validation blocked completion. No
success report is printed after a blocked execution.

The five [integration regressions](../integration/test_producer_consumer.py)
copy the real exporter source into disposable Git repositories and run it as a
subprocess with small synthetic QA gates. They cover source pass with retained
gaps, either producer failing independently, stale SHA, dirty tree and a missing
exporter. They do not replace the workflow's separate full-source-QA integration.

```sh
AIFT_TEST_RUNTIME_ROOT="$PWD/../AIFT-Runtime" \
AIFT_TEST_FORGE_ROOT="$PWD/../AIFT-Forge" \
python3 -B integration/test_producer_consumer.py -v
```

Local real-source validation at the pinned revisions passed both producers:
Runtime's 135 tests and Forge's active QA profile completed; the consumer
reported `partial` and `unsigned_supplied_json_not_authenticated`. This does not
validate open Runtime #15 or Forge #15. Those command-contract PRs and OS #76
remain independent; this PR is based on merged preparation work only. Future
producer changes need deliberate pin updates and reruns, rather than silently
following mutable branches.
