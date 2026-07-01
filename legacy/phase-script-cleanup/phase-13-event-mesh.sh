#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

export AIFT_ROOT="$ROOT"
export AIFT_OS_HOME="$OS"

cd "$OS" || exit 1

echo "== Phase 13: Federation Event Mesh =="

mkdir -p \
  internal/eventmesh \
  docs \
  tests \
  registry \
  reports \
  schemas \
  var/events \
  AI-Code-Training/scripts/phase-scripts \
  bin

cat > internal/eventmesh/eventmesh.go <<'GO'
package eventmesh

import (
"bufio"
"encoding/json"
"fmt"
"os"
"path/filepath"
"sort"
"strings"
"time"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Topic struct {
Name        string `json:"name"`
Status      string `json:"status"`
Description string `json:"description"`
Evidence    string `json:"evidence"`
}

type Subscriber struct {
Repo     string `json:"repo"`
Topic    string `json:"topic"`
Status   string `json:"status"`
Handler  string `json:"handler,omitempty"`
Evidence string `json:"evidence"`
}

type EventContract struct {
Repo        string       `json:"repo"`
Publishes   []Topic      `json:"publishes"`
Subscribes  []Subscriber `json:"subscribes"`
GeneratedAt string       `json:"generatedAt"`
}

type Registry struct {
GeneratedAt  string          `json:"generatedAt"`
Topics       []Topic         `json:"topics"`
Subscribers  []Subscriber    `json:"subscribers"`
Contracts    []EventContract `json:"contracts"`
}

func InitAll(cfg config.Config) error {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return err
}

for _, r := range repos {
if err := InitRepo(r.Name, r.Path); err != nil {
return err
}
}

return Scan(cfg)
}

func InitRepo(name, repoPath string) error {
dir := filepath.Join(repoPath, ".aift")
if err := os.MkdirAll(filepath.Join(dir, "events", "handlers"), 0755); err != nil {
return err
}

path := filepath.Join(dir, "events.json")
if _, err := os.Stat(path); err == nil {
return nil
}

contract := EventContract{
Repo: name,
Publishes: []Topic{
{Name: "repo.changed", Status: "planned", Description: "Repository content changed", Evidence: "default event contract"},
{Name: "capability.changed", Status: "planned", Description: "Capability status changed", Evidence: "default event contract"},
{Name: "manual.changed", Status: "planned", Description: "Manual source changed", Evidence: "default event contract"},
},
Subscribes: []Subscriber{},
GeneratedAt: time.Now().Format(time.RFC3339),
}

data, err := json.MarshalIndent(contract, "", "  ")
if err != nil {
return err
}

return os.WriteFile(path, append(data, '\n'), 0644)
}

func Scan(cfg config.Config) error {
repos, err := workspace.FindRepos(cfg)
if err != nil {
return err
}

topicMap := map[string]Topic{}
var subscribers []Subscriber
var contracts []EventContract

for _, r := range repos {
c, ok := readContract(r.Name, r.Path)
if !ok {
continue
}
contracts = append(contracts, c)

for _, t := range c.Publishes {
if t.Status == "" {
t.Status = "planned"
}
if t.Evidence == "" {
t.Evidence = ".aift/events.json"
}
topicMap[t.Name] = t
}

for _, s := range c.Subscribes {
if s.Status == "" {
s.Status = "planned"
}
if s.Evidence == "" {
s.Evidence = ".aift/events.json"
}
subscribers = append(subscribers, s)
}
}

topics := make([]Topic, 0, len(topicMap))
for _, t := range topicMap {
topics = append(topics, t)
}
sort.Slice(topics, func(i, j int) bool { return topics[i].Name < topics[j].Name })

sort.Slice(subscribers, func(i, j int) bool {
if subscribers[i].Topic == subscribers[j].Topic {
return subscribers[i].Repo < subscribers[j].Repo
}
return subscribers[i].Topic < subscribers[j].Topic
})

reg := Registry{
GeneratedAt: time.Now().Format(time.RFC3339),
Topics: topics,
Subscribers: subscribers,
Contracts: contracts,
}

if err := writeRegistry(cfg, reg); err != nil {
return err
}
if err := writeReport(cfg, reg); err != nil {
return err
}

return events.Emit(cfg, "eventmesh.scan", "eventmesh", "event mesh scanned", map[string]string{
"topics": fmt.Sprint(len(topics)),
"subscribers": fmt.Sprint(len(subscribers)),
})
}

func readContract(name, repoPath string) (EventContract, bool) {
path := filepath.Join(repoPath, ".aift", "events.json")
data, err := os.ReadFile(path)
if err != nil {
return EventContract{}, false
}

var c EventContract
if json.Unmarshal(data, &c) != nil {
return EventContract{}, false
}
if c.Repo == "" {
c.Repo = name
}
return c, true
}

func Publish(cfg config.Config, topic string, source string, message string) error {
if topic == "" {
return fmt.Errorf("topic is required")
}
if source == "" {
source = "manual"
}
if message == "" {
message = topic
}

return events.Emit(cfg, topic, source, message, map[string]string{
"eventMesh": "true",
})
}

func Tail(cfg config.Config, n int) error {
return events.Tail(cfg, n)
}

func Topics(cfg config.Config) error {
reg, err := loadOrScan(cfg)
if err != nil {
return err
}

fmt.Printf("%-32s %-12s %s\n", "TOPIC", "STATUS", "DESCRIPTION")
for _, t := range reg.Topics {
fmt.Printf("%-32s %-12s %s\n", t.Name, t.Status, t.Description)
}
return nil
}

func Subscribers(cfg config.Config) error {
reg, err := loadOrScan(cfg)
if err != nil {
return err
}

fmt.Printf("%-32s %-28s %-12s %s\n", "TOPIC", "REPO", "STATUS", "HANDLER")
for _, s := range reg.Subscribers {
fmt.Printf("%-32s %-28s %-12s %s\n", s.Topic, s.Repo, s.Status, s.Handler)
}
return nil
}

func Replay(cfg config.Config, topic string) error {
path := filepath.Join(cfg.OSHome, "var", "events", "events.jsonl")
file, err := os.Open(path)
if err != nil {
return err
}
defer file.Close()

scanner := bufio.NewScanner(file)
for scanner.Scan() {
line := scanner.Text()
if topic == "" || strings.Contains(line, `"type":"`+topic+`"`) || strings.Contains(line, `"topic":"`+topic+`"`) {
fmt.Println(line)
}
}
return scanner.Err()
}

func Report(cfg config.Config) error {
path := filepath.Join(cfg.OSHome, "reports", "event-mesh.md")
data, err := os.ReadFile(path)
if err != nil {
if err := Scan(cfg); err != nil {
return err
}
data, err = os.ReadFile(path)
if err != nil {
return err
}
}
fmt.Print(string(data))
return nil
}

func loadOrScan(cfg config.Config) (Registry, error) {
path := filepath.Join(cfg.OSHome, "registry", "event-mesh.json")
data, err := os.ReadFile(path)
if err != nil {
if err := Scan(cfg); err != nil {
return Registry{}, err
}
data, err = os.ReadFile(path)
if err != nil {
return Registry{}, err
}
}

var reg Registry
if err := json.Unmarshal(data, &reg); err != nil {
return Registry{}, err
}
return reg, nil
}

func writeRegistry(cfg config.Config, reg Registry) error {
out := filepath.Join(cfg.OSHome, "registry", "event-mesh.json")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}
data, err := json.MarshalIndent(reg, "", "  ")
if err != nil {
return err
}
if err := os.WriteFile(out, append(data, '\n'), 0644); err != nil {
return err
}
fmt.Println("Wrote", out)
return nil
}

func writeReport(cfg config.Config, reg Registry) error {
out := filepath.Join(cfg.OSHome, "reports", "event-mesh.md")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}

var b strings.Builder
b.WriteString("# Federation Event Mesh\n\n")
b.WriteString("The event mesh describes asynchronous federation coordination.\n\n")
b.WriteString("## Topics\n\n")
b.WriteString("| Topic | Status | Description | Evidence |\n")
b.WriteString("|---|---|---|---|\n")
for _, t := range reg.Topics {
b.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s |\n", t.Name, t.Status, t.Description, t.Evidence))
}

b.WriteString("\n## Subscribers\n\n")
b.WriteString("| Topic | Repository | Status | Handler | Evidence |\n")
b.WriteString("|---|---|---|---|---|\n")
for _, s := range reg.Subscribers {
b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | %s |\n", s.Topic, s.Repo, s.Status, s.Handler, s.Evidence))
}

return os.WriteFile(out, []byte(b.String()), 0644)
}
GO

python - <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text()

if '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/eventmesh"' not in s:
    s = s.replace(
        '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"',
        '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"\n\t"github.com/AIFreedomTrustFederation/AIFT-OS/internal/eventmesh"'
    )

if 'fmt.Println("  mesh init-all|scan|topics|subscribers|publish|replay|tail|report")' not in s:
    s = s.replace(
        'fmt.Println("  graph [summary|repo|type|status]")',
        'fmt.Println("  graph [summary|repo|type|status]")\n\tfmt.Println("  mesh init-all|scan|topics|subscribers|publish|replay|tail|report")'
    )

case_block = r'''case "mesh":
err = runMesh(cfg, args)
'''

if 'case "mesh":' not in s:
    s = s.replace('case "verify":\n\t\terr = verify(cfg)', case_block + '\tcase "verify":\n\t\terr = verify(cfg)')

func_block = r'''
func runMesh(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "scan" {
return eventmesh.Scan(cfg)
}
switch args[0] {
case "init-all":
return eventmesh.InitAll(cfg)
case "topics":
return eventmesh.Topics(cfg)
case "subscribers":
return eventmesh.Subscribers(cfg)
case "publish":
if len(args) < 2 {
return fmt.Errorf("usage: aift mesh publish <topic> [source] [message]")
}
source := "manual"
message := args[1]
if len(args) > 2 {
source = args[2]
}
if len(args) > 3 {
message = args[3]
}
return eventmesh.Publish(cfg, args[1], source, message)
case "replay":
topic := ""
if len(args) > 1 {
topic = args[1]
}
return eventmesh.Replay(cfg, topic)
case "tail":
return eventmesh.Tail(cfg, 25)
case "report":
return eventmesh.Report(cfg)
default:
return fmt.Errorf("usage: aift mesh init-all|scan|topics|subscribers|publish|replay|tail|report")
}
}
'''

if 'func runMesh(' not in s:
    s = s.replace('func verify(cfg config.Config) error {', func_block + '\nfunc verify(cfg config.Config) error {')

if 'if err := eventmesh.Scan(cfg); err != nil {' not in s:
    s = s.replace(
        'if err := graph.Build(cfg); err != nil {\n\t\treturn err\n\t}',
        'if err := graph.Build(cfg); err != nil {\n\t\treturn err\n\t}\n\tif err := eventmesh.Scan(cfg); err != nil {\n\t\treturn err\n\t}'
    )

p.write_text(s)
PY

cat > schemas/event-mesh.schema.json <<'JSON'
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "AIFT Event Mesh Registry",
  "type": "object",
  "required": ["generatedAt", "topics", "subscribers", "contracts"],
  "properties": {
    "generatedAt": { "type": "string" },
    "topics": { "type": "array" },
    "subscribers": { "type": "array" },
    "contracts": { "type": "array" }
  }
}
JSON

cat > schemas/events-contract.schema.json <<'JSON'
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "AIFT Repo Event Contract",
  "type": "object",
  "required": ["repo", "publishes", "subscribes"],
  "properties": {
    "repo": { "type": "string" },
    "publishes": { "type": "array" },
    "subscribes": { "type": "array" },
    "generatedAt": { "type": "string" }
  }
}
JSON

cat > tests/event-mesh-smoke.sh <<'SH'
#!/usr/bin/env sh
set -eu

ROOT="${AIFT_ROOT:-$HOME/AIFT}"
OS="${AIFT_OS_HOME:-$ROOT/AIFT-OS}"
BIN="$OS/bin/aiftd"

cd "$OS" || exit 1

go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" mesh init-all >/dev/null
"$ROOT/aift" mesh scan >/dev/null
"$ROOT/aift" mesh topics >/dev/null
"$ROOT/aift" mesh subscribers >/dev/null
"$ROOT/aift" mesh publish phase13.test tests "phase 13 smoke event" >/dev/null
"$ROOT/aift" mesh replay phase13.test >/dev/null
"$ROOT/aift" mesh tail >/dev/null
"$ROOT/aift" mesh report >/dev/null
"$ROOT/aift" verify >/dev/null

test -f registry/event-mesh.json
test -f reports/event-mesh.md
test -f .aift/events.json
test -d .aift/events/handlers

echo "OK: event mesh smoke passed"
SH

chmod +x tests/event-mesh-smoke.sh

cat > docs/PHASE-13-EVENT-MESH.md <<'DOC'
# Phase 13: Federation Event Mesh

AIFT-OS now has the foundation for asynchronous federation coordination.

## Principle

Repositories are not chained through hard-coded if/then workflows.

They publish events, subscribe to events, and declare handlers through `.aift/events.json`.

## Commands

- `aift mesh init-all`
- `aift mesh scan`
- `aift mesh topics`
- `aift mesh subscribers`
- `aift mesh publish <topic> [source] [message]`
- `aift mesh replay [topic]`
- `aift mesh tail`
- `aift mesh report`

## Per Repo

- `.aift/events.json`
- `.aift/events/handlers/`

## Generated

- `registry/event-mesh.json`
- `reports/event-mesh.md`

## Truth Rule

A topic or subscriber can be planned, detected, ready, v1, broken, deprecated, or disabled.

AIFT-OS records event contracts but does not pretend unverified handlers are running.
DOC

cp phase-13-event-mesh.sh AI-Code-Training/scripts/phase-scripts/phase-13-event-mesh.sh 2>/dev/null || true

echo "== Build/test =="
go clean -cache
gofmt -w cmd internal
go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"

"$ROOT/aift" mesh init-all
"$ROOT/aift" mesh scan
"$ROOT/aift" mesh topics
"$ROOT/aift" mesh publish phase13.created AIFT-OS "Phase 13 event mesh created"
sh tests/event-mesh-smoke.sh

echo "== Commit AIFT-OS =="
git add .
if git diff --cached --quiet; then
  echo "AIFT-OS: nothing staged."
else
  git commit -m "Add federation event mesh"
fi
git push origin main

echo "== Commit event contracts in federation repos =="
for repo in "$ROOT"/*; do
  [ -d "$repo/.git" ] || continue
  name="$(basename "$repo")"
  cd "$repo" || continue

  git add .aift/events.json .aift/events 2>/dev/null || true

  if git diff --cached --quiet; then
    echo "$name: no event contract changes."
  else
    git commit -m "Add AIFT event mesh contract"
    branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo main)"
    if git remote get-url origin >/dev/null 2>&1; then
      git push origin "$branch" || true
    fi
  fi
done

cd "$OS" || exit 1

echo
echo "DONE."
echo "Try:"
echo "  ~/AIFT/aift mesh topics"
echo "  ~/AIFT/aift mesh publish repo.changed AIFT-OS 'manual test event'"
echo "  ~/AIFT/aift mesh replay repo.changed"
echo "  ~/AIFT/aift mesh report"
