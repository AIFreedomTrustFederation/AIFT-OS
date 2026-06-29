package eventbus

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/sliceutil"
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
		Topics:        sliceutil.SortedBoolMapKeys(topics),
		Sources:       sliceutil.SortedBoolMapKeys(sources),
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

func escapeTable(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}
