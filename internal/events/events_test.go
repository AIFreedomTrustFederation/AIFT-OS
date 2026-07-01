package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestEmit(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}

	err := Emit(cfg, "test.event", "unit-test", "hello", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	path := filepath.Join(dir, "var", "events", "events.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("events.jsonl not created: %v", err)
	}

	var ev Event
	if err := json.Unmarshal(data, &ev); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if ev.Type != "test.event" {
		t.Errorf("Type = %q, want test.event", ev.Type)
	}
	if ev.Source != "unit-test" {
		t.Errorf("Source = %q, want unit-test", ev.Source)
	}
	if ev.Message != "hello" {
		t.Errorf("Message = %q, want hello", ev.Message)
	}
	if ev.Data["key"] != "value" {
		t.Errorf("Data[key] = %q, want value", ev.Data["key"])
	}
	if ev.Time == "" {
		t.Error("Time should not be empty")
	}
}

func TestEmitNilData(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}

	err := Emit(cfg, "test.event", "src", "msg", nil)
	if err != nil {
		t.Fatalf("Emit with nil data failed: %v", err)
	}

	path := filepath.Join(dir, "var", "events", "events.jsonl")
	data, _ := os.ReadFile(path)
	var ev Event
	json.Unmarshal(data, &ev)
	if ev.Type != "test.event" {
		t.Errorf("Type = %q", ev.Type)
	}
}

func TestEmitMultiple(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}

	Emit(cfg, "e1", "s", "m1", nil)
	Emit(cfg, "e2", "s", "m2", nil)
	Emit(cfg, "e3", "s", "m3", nil)

	path := filepath.Join(dir, "var", "events", "events.jsonl")
	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 events, got %d", len(lines))
	}
}

func TestTailNoEvents(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}

	err := Tail(cfg, 10)
	if err != nil {
		t.Errorf("Tail with no events should not error: %v", err)
	}
}

func TestTailAfterEmit(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}

	Emit(cfg, "e1", "s", "m1", nil)
	Emit(cfg, "e2", "s", "m2", nil)
	Emit(cfg, "e3", "s", "m3", nil)

	err := Tail(cfg, 2)
	if err != nil {
		t.Errorf("Tail failed: %v", err)
	}
}

func TestTailZeroLimit(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}

	Emit(cfg, "e1", "s", "m1", nil)

	err := Tail(cfg, 0)
	if err != nil {
		t.Errorf("Tail with limit=0 failed: %v", err)
	}
}

func TestTailEmptyFile(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}

	evDir := filepath.Join(dir, "var", "events")
	os.MkdirAll(evDir, 0755)
	os.WriteFile(filepath.Join(evDir, "events.jsonl"), []byte(""), 0644)

	err := Tail(cfg, 10)
	if err != nil {
		t.Errorf("Tail with empty file should not error: %v", err)
	}
}

func TestEventJSON(t *testing.T) {
	ev := Event{
		Time:    "2024-01-01T00:00:00Z",
		Type:    "test",
		Source:  "src",
		Message: "msg",
		Data:    map[string]string{"k": "v"},
	}
	data, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Type != "test" || decoded.Data["k"] != "v" {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestEventDataOmitEmpty(t *testing.T) {
	ev := Event{Time: "now", Type: "t", Source: "s", Message: "m"}
	data, _ := json.Marshal(ev)
	if strings.Contains(string(data), "data") {
		t.Error("nil data should be omitted from JSON")
	}
}
