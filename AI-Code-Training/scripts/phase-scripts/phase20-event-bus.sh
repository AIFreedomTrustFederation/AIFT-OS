#!/usr/bin/env bash
set -Eeuo pipefail

cd ~/AIFT/AIFT-OS

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
BIN="bin/aiftd"
REPORT="reports/event-bus-$STAMP.md"

mkdir -p internal/eventbus docs tests schemas reports registry var/events AI-Code-Training/scripts/phase-scripts bin

echo "== AIFT Event Bus =="

TMP_GO="$(mktemp)"

cat > "$TMP_GO" <<'GO'
package eventbus

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
)

type Event struct {
ID          string            `json:"id"`
Schema      string            `json:"schema"`
Topic       string            `json:"topic"`
Kind        string            `json:"kind"`
Source      string            `json:"source"`
Message     string            `json:"message"`
Status      string            `json:"status"`
Payload     map[string]string `json:"payload"`
Evidence    []Evidence        `json:"evidence"`
GeneratedAt string            `json:"generatedAt"`
}

type Evidence struct {
Kind        string `json:"kind"`
Path        string `json:"path"`
Description string `json:"description"`
ObservedAt  string `json:"observedAt"`
}

type Summary struct {
SchemaVersion string   `json:"schemaVersion"`
GeneratedAt   string   `json:"generatedAt"`
LogPath       string   `json:"logPath"`
Count         int      `json:"count"`
Topics        []string `json:"topics"`
Sources       []string `json:"sources"`
}

func Publish(cfg config.Config, topic string, kind string, source string, message string, payload map[string]string) error {
if strings.TrimSpace(topic) == "" {
return fmt.Errorf("event topic is required")
}
if strings.TrimSpace(kind) == "" {
kind = "event"
}
if strings.TrimSpace(source) == "" {
source = "aiftd"
}
if payload == nil {
payload = map[string]string{}
}

now := time.Now().UTC().Format(time.RFC3339Nano)
event := Event{
ID:      eventID(now, source, topic),
Schema:  "aift.event.v1",
Topic:   topic,
Kind:    kind,
Source:  source,
Message: message,
Status:  "detected",
Payload: payload,
Evidence: []Evidence{
{
Kind:        "command",
Path:        "aiftd event-bus publish",
Description: "Event was published through the local AIFT event bus command.",
ObservedAt:  now,
},
},
GeneratedAt: now,
}

return appendEvent(cfg, event)
}

func List(cfg config.Config) error {
events, err := Load(cfg)
if err != nil {
return err
}

fmt.Printf("%-40s %-28s %-20s %s\n", "ID", "TOPIC", "SOURCE", "MESSAGE")
for _, event := range events {
fmt.Printf("%-40s %-28s %-20s %s\n", event.ID, event.Topic, event.Source, event.Message)
}
return nil
}

func Replay(cfg config.Config, topic string) error {
events, err := Load(cfg)
if err != nil {
return err
}

for _, event := range events {
if topic != "" && event.Topic != topic {
continue
}
data, err := json.MarshalIndent(event, "", "  ")
if err != nil {
return err
}
fmt.Println(string(data))
}
return nil
}

func Report(cfg config.Config) error {
events, err := Load(cfg)
if err != nil {
return err
}
if err := WriteReport(cfg, events); err != nil {
return err
}

data, err := os.ReadFile(filepath.Join(cfg.OSHome, "reports", "event-bus.md"))
if err != nil {
return err
}
fmt.Print(string(data))
return nil
}

func Snapshot(cfg config.Config) (Summary, error) {
events, err := Load(cfg)
if err != nil {
return Summary{}, err
}

topics := map[string]bool{}
sources := map[string]bool{}
for _, event := range events {
if event.Topic != "" {
topics[event.Topic] = true
}
if event.Source != "" {
sources[event.Source] = true
}
}

return Summary{
SchemaVersion: "aift.eventbus.summary.v1",
GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
LogPath:       logPath(cfg),
Count:         len(events),
Topics:        sortedKeys(topics),
Sources:       sortedKeys(sources),
}, nil
}

func Load(cfg config.Config) ([]Event, error) {
path := logPath(cfg)
file, err := os.Open(path)
if err != nil {
if os.IsNotExist(err) {
return []Event{}, nil
}
return nil, err
}
defer file.Close()

out := []Event{}
scanner := bufio.NewScanner(file)
scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

lineNumber := 0
for scanner.Scan() {
lineNumber++
line := strings.TrimSpace(scanner.Text())
if line == "" {
continue
}

var event Event
if err := json.Unmarshal([]byte(line), &event); err != nil {
return nil, fmt.Errorf("invalid event log json at line %d: %w", lineNumber, err)
}
out = append(out, event)
}

if err := scanner.Err(); err != nil {
return nil, err
}

return out, nil
}

func WriteReport(cfg config.Config, events []Event) error {
out := filepath.Join(cfg.OSHome, "reports", "event-bus.md")
if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
return err
}

summary, err := Snapshot(cfg)
if err != nil {
return err
}

var b strings.Builder
b.WriteString("# AIFT Event Bus\n\n")
b.WriteString("The event bus is append-only runtime state. It records observed actions without treating them as source truth.\n\n")
b.WriteString("## Summary\n\n")
b.WriteString(fmt.Sprintf("- Events: %d\n", summary.Count))
b.WriteString(fmt.Sprintf("- Topics: %s\n", strings.Join(summary.Topics, ", ")))
b.WriteString(fmt.Sprintf("- Sources: %s\n", strings.Join(summary.Sources, ", ")))
b.WriteString(fmt.Sprintf("- Log: `%s`\n\n", summary.LogPath))

b.WriteString("## Events\n\n")
b.WriteString("| ID | Topic | Source | Status | Message |\n")
b.WriteString("|---|---|---|---|---|\n")
for _, event := range events {
b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | %s |\n",
event.ID,
event.Topic,
event.Source,
event.Status,
escapeTable(event.Message),
))
}

return os.WriteFile(out, []byte(b.String()), 0644)
}

func appendEvent(cfg config.Config, event Event) error {
path := logPath(cfg)
if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
return err
}

data, err := json.Marshal(event)
if err != nil {
return err
}

file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
if err != nil {
return err
}
defer file.Close()

if _, err := file.Write(append(data, '\n')); err != nil {
return err
}
return nil
}

func logPath(cfg config.Config) string {
return filepath.Join(cfg.OSHome, "var", "events", "event-bus.jsonl")
}

func eventID(now string, source string, topic string) string {
raw := now + "." + source + "." + topic
raw = strings.ToLower(raw)
replacer := strings.NewReplacer(":", "-", "/", ".", " ", "-", "_", "-", "+", "-", "z", "z")
raw = replacer.Replace(raw)
raw = strings.Trim(raw, ".-")
if len(raw) > 96 {
raw = raw[:96]
}
return raw
}

func sortedKeys(values map[string]bool) []string {
out := []string{}
for value := range values {
out = append(out, value)
}
sort.Strings(out)
return out
}

func escapeTable(value string) string {
value = strings.ReplaceAll(value, "|", "\\|")
value = strings.ReplaceAll(value, "\n", " ")
return value
}
GO

echo "== Validate generated Go before installing =="
gofmt -w "$TMP_GO"
cp "$TMP_GO" internal/eventbus/eventbus.go
rm -f "$TMP_GO"

echo "== Patch CLI =="
python3 - <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text()

imp = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/eventbus"'
if imp not in s:
    marker = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/discoveryengine"'
    if marker not in s:
        marker = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"'
    s = s.replace(marker, marker + "\n\t" + imp, 1)

help_line = 'fmt.Println("  event-bus publish|list|replay|report")'
if help_line not in s:
    marker = 'fmt.Println("  discovery scan|list|object|report")'
    if marker in s:
        s = s.replace(marker, marker + "\n\t" + help_line, 1)

if 'case "event-bus":' not in s:
    marker = 'case "verify":\n\t\terr = verify(cfg)'
    if marker not in s:
        raise SystemExit("missing verify case marker")
    s = s.replace(marker, 'case "event-bus":\n\t\terr = runEventBus(cfg, args)\n\t' + marker, 1)

if 'func runEventBus(' not in s:
    marker = 'func verify(cfg config.Config) error {'
    if marker not in s:
        raise SystemExit("missing verify function marker")
    block = r'''
func runEventBus(cfg config.Config, args []string) error {
if len(args) == 0 || args[0] == "list" {
return eventbus.List(cfg)
}

switch args[0] {
case "publish":
if len(args) < 3 {
return fmt.Errorf("usage: aift event-bus publish <topic> <message> [key=value...]")
}
payload := map[string]string{}
for _, item := range args[3:] {
parts := strings.SplitN(item, "=", 2)
if len(parts) == 2 {
payload[parts[0]] = parts[1]
}
}
return eventbus.Publish(cfg, args[1], "manual", "aiftd", args[2], payload)
case "list":
return eventbus.List(cfg)
case "replay":
topic := ""
if len(args) > 1 {
topic = args[1]
}
return eventbus.Replay(cfg, topic)
case "report":
return eventbus.Report(cfg)
default:
return fmt.Errorf("usage: aift event-bus publish|list|replay|report")
}
}

'''
    s = s.replace(marker, block + marker, 1)

if '"strings"' not in s:
    s = s.replace('"fmt"', '"fmt"\n\t"strings"', 1)

p.write_text(s)
PY

echo "== Write docs, schema, smoke test =="
cat > tests/event-bus-smoke.sh <<'SH'
#!/usr/bin/env sh
set -eu

go test ./...
go build -o bin/aiftd ./cmd/aift

bin/aiftd event-bus publish system.test "event bus smoke" source=smoke >/dev/null
bin/aiftd event-bus list >/dev/null
bin/aiftd event-bus replay system.test >/dev/null
bin/aiftd event-bus report >/dev/null

test -f var/events/event-bus.jsonl
test -f reports/event-bus.md

echo "OK: event bus smoke passed"
SH
chmod +x tests/event-bus-smoke.sh

cat > docs/EVENT-BUS.md <<'DOC'
# AIFT Event Bus

The Event Bus is the nervous system of AIFT-OS.

It records meaningful operating-system activity as append-only runtime evidence.

## Principles

- Events are runtime state, not source truth.
- Events should be append-only.
- Events should support replay.
- Events should be generated from observable actions.
- Events should not pretend a capability is active unless validation proves it.

## Commands

- `aiftd event-bus publish <topic> <message> [key=value...]`
- `aiftd event-bus list`
- `aiftd event-bus replay [topic]`
- `aiftd event-bus report`

## Runtime artifacts

- `var/events/event-bus.jsonl`
- `reports/event-bus.md`

These are intentionally ignored generated runtime artifacts.
DOC

cat > schemas/event-bus.schema.json <<'JSON'
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "AIFT Event Bus Event",
  "type": "object",
  "required": ["id", "schema", "topic", "kind", "source", "message", "status", "payload", "evidence", "generatedAt"],
  "properties": {
    "id": { "type": "string" },
    "schema": { "type": "string" },
    "topic": { "type": "string" },
    "kind": { "type": "string" },
    "source": { "type": "string" },
    "message": { "type": "string" },
    "status": {
      "type": "string",
      "enum": ["planned", "detected", "ready", "active", "deprecated", "removed"]
    },
    "payload": { "type": "object" },
    "evidence": { "type": "array" },
    "generatedAt": { "type": "string" }
  }
}
JSON

echo "== Verify =="
gofmt -w internal/eventbus/eventbus.go cmd/aift/main.go
go test ./...
go build -o "$BIN" ./cmd/aift
chmod +x "$BIN"
sh tests/event-bus-smoke.sh

cat > "$REPORT" <<EOF
# Event Bus Implementation Report

Generated: $STAMP

Passed:

- gofmt
- go test ./...
- go build ./cmd/aift
- tests/event-bus-smoke.sh

Generated runtime artifacts are intentionally ignored:

- var/events/event-bus.jsonl
- reports/event-bus.md
EOF

cp "$0" AI-Code-Training/scripts/phase-scripts/phase20-event-bus.sh 2>/dev/null || true

echo "== Stage source files only =="
git add \
  internal/eventbus/eventbus.go \
  cmd/aift/main.go \
  tests/event-bus-smoke.sh \
  docs/EVENT-BUS.md \
  schemas/event-bus.schema.json \
  AI-Code-Training/scripts/phase-scripts/phase20-event-bus.sh

echo "== Commit and push =="
if git diff --cached --quiet; then
  echo "Nothing staged."
else
  git commit -m "Implement event bus foundation"
  git push origin main
fi

echo "DONE"
