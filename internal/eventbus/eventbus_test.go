package eventbus

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestEventID(t *testing.T) {
	tests := []struct {
		now    string
		source string
		topic  string
	}{
		{"2024-01-01T00:00:00Z", "aiftd", "test.topic"},
		{"2024-01-01T00:00:00Z", "my-source", "my/topic"},
		{"2024-01-01T00:00:00+05:00", "src", "topic"},
	}
	for _, tt := range tests {
		id := eventID(tt.now, tt.source, tt.topic)
		if id == "" {
			t.Error("eventID should not be empty")
		}
		if len(id) > 96 {
			t.Errorf("eventID length %d exceeds 96", len(id))
		}
		if strings.HasPrefix(id, ".") || strings.HasPrefix(id, "-") {
			t.Errorf("eventID %q should not start with . or -", id)
		}
		if strings.HasSuffix(id, ".") || strings.HasSuffix(id, "-") {
			t.Errorf("eventID %q should not end with . or -", id)
		}
	}
}

func TestEventIDMaxLength(t *testing.T) {
	long := strings.Repeat("a", 200)
	id := eventID("2024-01-01T00:00:00Z", long, long)
	if len(id) > 96 {
		t.Errorf("eventID length %d should be capped at 96", len(id))
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]bool{"cherry": true, "apple": true, "banana": true}
	got := sortedKeys(m)
	want := []string{"apple", "banana", "cherry"}
	if len(got) != len(want) {
		t.Fatalf("sortedKeys returned %d keys, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("sortedKeys[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSortedKeysEmpty(t *testing.T) {
	got := sortedKeys(map[string]bool{})
	if len(got) != 0 {
		t.Errorf("sortedKeys(empty) returned %d keys, want 0", len(got))
	}
}

func TestEscapeTable(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"a | b", "a \\| b"},
		{"line1\nline2", "line1 line2"},
		{"a | b\nc | d", "a \\| b c \\| d"},
		{"", ""},
	}
	for _, tt := range tests {
		got := escapeTable(tt.input)
		if got != tt.want {
			t.Errorf("escapeTable(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func testConfig(t *testing.T) config.Config {
	t.Helper()
	dir := t.TempDir()
	return config.Config{
		Root:   dir,
		OSHome: dir,
	}
}

func TestPublishAndLoad(t *testing.T) {
	cfg := testConfig(t)

	err := Publish(cfg, "test.topic", "event", "test-source", "hello world", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	events, err := Load(cfg)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	e := events[0]
	if e.Topic != "test.topic" {
		t.Errorf("topic = %q, want test.topic", e.Topic)
	}
	if e.Source != "test-source" {
		t.Errorf("source = %q, want test-source", e.Source)
	}
	if e.Message != "hello world" {
		t.Errorf("message = %q, want hello world", e.Message)
	}
	if e.Payload["key"] != "val" {
		t.Errorf("payload[key] = %q, want val", e.Payload["key"])
	}
	if e.Schema != "aift.event.v1" {
		t.Errorf("schema = %q, want aift.event.v1", e.Schema)
	}
	if e.Status != "detected" {
		t.Errorf("status = %q, want detected", e.Status)
	}
}

func TestPublishEmptyTopic(t *testing.T) {
	cfg := testConfig(t)
	err := Publish(cfg, "", "event", "source", "msg", nil)
	if err == nil {
		t.Error("Publish with empty topic should fail")
	}
}

func TestPublishWhitespaceTopic(t *testing.T) {
	cfg := testConfig(t)
	err := Publish(cfg, "   ", "event", "source", "msg", nil)
	if err == nil {
		t.Error("Publish with whitespace-only topic should fail")
	}
}

func TestPublishDefaultKind(t *testing.T) {
	cfg := testConfig(t)
	Publish(cfg, "test.topic", "", "source", "msg", nil)

	events, _ := Load(cfg)
	if len(events) != 1 {
		t.Fatal("expected 1 event")
	}
	if events[0].Kind != "event" {
		t.Errorf("default kind = %q, want event", events[0].Kind)
	}
}

func TestPublishDefaultSource(t *testing.T) {
	cfg := testConfig(t)
	Publish(cfg, "test.topic", "event", "", "msg", nil)

	events, _ := Load(cfg)
	if len(events) != 1 {
		t.Fatal("expected 1 event")
	}
	if events[0].Source != "aiftd" {
		t.Errorf("default source = %q, want aiftd", events[0].Source)
	}
}

func TestPublishNilPayload(t *testing.T) {
	cfg := testConfig(t)
	err := Publish(cfg, "test.topic", "event", "source", "msg", nil)
	if err != nil {
		t.Fatalf("Publish with nil payload should not fail: %v", err)
	}

	events, _ := Load(cfg)
	if events[0].Payload == nil {
		t.Error("nil payload should be normalized to empty map")
	}
}

func TestLoadEmptyLog(t *testing.T) {
	cfg := testConfig(t)
	events, err := Load(cfg)
	if err != nil {
		t.Fatalf("Load with no log file should not fail: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	cfg := testConfig(t)
	logDir := filepath.Join(cfg.OSHome, "var", "events")
	os.MkdirAll(logDir, 0755)
	os.WriteFile(filepath.Join(logDir, "event-bus.jsonl"), []byte("not json\n"), 0644)

	_, err := Load(cfg)
	if err == nil {
		t.Error("Load with invalid JSON should fail")
	}
}

func TestLoadMultipleEvents(t *testing.T) {
	cfg := testConfig(t)

	for i := 0; i < 3; i++ {
		Publish(cfg, "topic."+string(rune('a'+i)), "event", "source", "msg", nil)
	}

	events, err := Load(cfg)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}
}

func TestLoadSkipsBlankLines(t *testing.T) {
	cfg := testConfig(t)
	logDir := filepath.Join(cfg.OSHome, "var", "events")
	os.MkdirAll(logDir, 0755)

	e := Event{
		ID: "test", Schema: "aift.event.v1", Topic: "t", Kind: "event",
		Source: "s", Message: "m", Status: "detected", Payload: map[string]string{},
		Evidence: []Evidence{}, GeneratedAt: "2024-01-01T00:00:00Z",
	}
	data, _ := json.Marshal(e)
	content := string(data) + "\n\n" + string(data) + "\n"
	os.WriteFile(filepath.Join(logDir, "event-bus.jsonl"), []byte(content), 0644)

	events, err := Load(cfg)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events (blank lines skipped), got %d", len(events))
	}
}

func TestSnapshot(t *testing.T) {
	cfg := testConfig(t)
	Publish(cfg, "topic.a", "event", "src1", "msg1", nil)
	Publish(cfg, "topic.b", "event", "src2", "msg2", nil)
	Publish(cfg, "topic.a", "event", "src1", "msg3", nil)

	summary, err := Snapshot(cfg)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}
	if summary.Count != 3 {
		t.Errorf("count = %d, want 3", summary.Count)
	}
	if len(summary.Topics) != 2 {
		t.Errorf("topics = %d, want 2", len(summary.Topics))
	}
	if len(summary.Sources) != 2 {
		t.Errorf("sources = %d, want 2", len(summary.Sources))
	}
	if summary.SchemaVersion != "aift.eventbus.summary.v1" {
		t.Errorf("schema = %q, want aift.eventbus.summary.v1", summary.SchemaVersion)
	}
}

func TestLogPath(t *testing.T) {
	cfg := config.Config{OSHome: "/tmp/test-os"}
	path := logPath(cfg)
	if path != "/tmp/test-os/var/events/event-bus.jsonl" {
		t.Errorf("logPath = %q", path)
	}
}
