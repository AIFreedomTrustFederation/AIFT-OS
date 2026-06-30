package eventmesh

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestReadContract(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "events.json"), []byte(`{
		"repo": "test-repo",
		"publishes": [
			{"name": "repo.changed", "status": "planned", "description": "Repo changed", "evidence": "test"}
		],
		"subscribes": [
			{"repo": "test-repo", "topic": "build.done", "status": "planned", "evidence": "test"}
		]
	}`), 0644)

	c, ok := readContract("test-repo", dir)
	if !ok {
		t.Fatal("readContract should succeed")
	}
	if c.Repo != "test-repo" {
		t.Errorf("Repo = %q, want test-repo", c.Repo)
	}
	if len(c.Publishes) != 1 {
		t.Fatalf("expected 1 publish, got %d", len(c.Publishes))
	}
	if c.Publishes[0].Name != "repo.changed" {
		t.Errorf("publish name = %q, want repo.changed", c.Publishes[0].Name)
	}
	if len(c.Subscribes) != 1 {
		t.Fatalf("expected 1 subscriber, got %d", len(c.Subscribes))
	}
	if c.Subscribes[0].Topic != "build.done" {
		t.Errorf("subscriber topic = %q, want build.done", c.Subscribes[0].Topic)
	}
}

func TestReadContractMissing(t *testing.T) {
	dir := t.TempDir()
	_, ok := readContract("test-repo", dir)
	if ok {
		t.Error("readContract should fail for missing file")
	}
}

func TestReadContractInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "events.json"), []byte("not json"), 0644)

	_, ok := readContract("test-repo", dir)
	if ok {
		t.Error("readContract should fail for invalid JSON")
	}
}

func TestReadContractDefaultsRepo(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "events.json"), []byte(`{
		"publishes": [],
		"subscribes": []
	}`), 0644)

	c, ok := readContract("fallback-name", dir)
	if !ok {
		t.Fatal("readContract should succeed")
	}
	if c.Repo != "fallback-name" {
		t.Errorf("Repo = %q, want fallback-name (defaulted from param)", c.Repo)
	}
}

func TestTopicJSON(t *testing.T) {
	topic := Topic{
		Name:        "repo.changed",
		Status:      "planned",
		Description: "Repository content changed",
		Evidence:    "default event contract",
	}

	data, err := json.Marshal(topic)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Topic
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Name != "repo.changed" {
		t.Errorf("Name = %q", decoded.Name)
	}
	if decoded.Status != "planned" {
		t.Errorf("Status = %q", decoded.Status)
	}
}

func TestSubscriberJSON(t *testing.T) {
	sub := Subscriber{
		Repo:     "test-repo",
		Topic:    "build.done",
		Status:   "planned",
		Handler:  "scripts/on-build.sh",
		Evidence: "test",
	}

	data, err := json.Marshal(sub)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Subscriber
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Handler != "scripts/on-build.sh" {
		t.Errorf("Handler = %q", decoded.Handler)
	}
}

func TestSubscriberJSONOmitEmptyHandler(t *testing.T) {
	sub := Subscriber{
		Repo:     "test-repo",
		Topic:    "build.done",
		Status:   "planned",
		Evidence: "test",
	}

	data, err := json.Marshal(sub)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	raw := string(data)
	if contains(raw, "handler") {
		t.Error("empty handler should be omitted from JSON")
	}
}

func TestEventContractJSON(t *testing.T) {
	ec := EventContract{
		Repo: "test",
		Publishes: []Topic{
			{Name: "repo.changed", Status: "planned", Description: "test", Evidence: "test"},
		},
		Subscribes:  []Subscriber{},
		GeneratedAt: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(ec)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded EventContract
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Repo != "test" {
		t.Errorf("Repo = %q", decoded.Repo)
	}
	if len(decoded.Publishes) != 1 {
		t.Errorf("Publishes count = %d, want 1", len(decoded.Publishes))
	}
}

func TestRegistryJSON(t *testing.T) {
	reg := Registry{
		GeneratedAt: "2024-01-01T00:00:00Z",
		Topics: []Topic{
			{Name: "repo.changed", Status: "planned", Description: "test", Evidence: "test"},
			{Name: "build.done", Status: "planned", Description: "test", Evidence: "test"},
		},
		Subscribers: []Subscriber{
			{Repo: "a", Topic: "repo.changed", Status: "planned", Evidence: "test"},
		},
		Contracts: []EventContract{},
	}

	data, err := json.Marshal(reg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Registry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(decoded.Topics) != 2 {
		t.Errorf("Topics count = %d, want 2", len(decoded.Topics))
	}
	if len(decoded.Subscribers) != 1 {
		t.Errorf("Subscribers count = %d, want 1", len(decoded.Subscribers))
	}
}

func TestPublishValidation(t *testing.T) {
	// Publish requires a non-empty topic
	cfg := testCfg(t)
	err := Publish(cfg, "", "source", "msg")
	if err == nil {
		t.Error("Publish with empty topic should fail")
	}
}

func TestPublishDefaultSource(t *testing.T) {
	cfg := testCfg(t)

	// Should succeed with empty source (defaults to "manual")
	err := Publish(cfg, "test.topic", "", "msg")
	if err != nil {
		t.Errorf("Publish with empty source should succeed: %v", err)
	}
}

func TestPublishDefaultMessage(t *testing.T) {
	cfg := testCfg(t)

	// Should succeed with empty message (defaults to topic name)
	err := Publish(cfg, "test.topic", "source", "")
	if err != nil {
		t.Errorf("Publish with empty message should succeed: %v", err)
	}
}

func TestInitRepoCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	err := InitRepo("test-repo", dir)
	if err != nil {
		t.Fatalf("InitRepo failed: %v", err)
	}

	eventsPath := filepath.Join(dir, ".aift", "events.json")
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("events.json not created: %v", err)
	}

	var c EventContract
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("invalid events.json: %v", err)
	}
	if c.Repo != "test-repo" {
		t.Errorf("Repo = %q, want test-repo", c.Repo)
	}
	if len(c.Publishes) != 3 {
		t.Errorf("expected 3 default publish topics, got %d", len(c.Publishes))
	}

	handlersDir := filepath.Join(dir, ".aift", "events", "handlers")
	info, err := os.Stat(handlersDir)
	if err != nil || !info.IsDir() {
		t.Error("handlers directory should be created")
	}
}

func TestInitRepoIdempotent(t *testing.T) {
	dir := t.TempDir()
	InitRepo("test-repo", dir)

	// Modify the file
	eventsPath := filepath.Join(dir, ".aift", "events.json")
	original, _ := os.ReadFile(eventsPath)
	os.WriteFile(eventsPath, []byte(`{"repo":"modified"}`), 0644)

	// InitRepo should not overwrite
	InitRepo("test-repo", dir)
	data, _ := os.ReadFile(eventsPath)
	if string(data) != `{"repo":"modified"}` {
		t.Errorf("InitRepo overwrote existing file. Before: %s, After: %s", original, data)
	}
}

func testCfg(t *testing.T) config.Load()
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
