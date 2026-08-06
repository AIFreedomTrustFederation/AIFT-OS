package uxi

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type fakeCompleter struct {
	text string
	err  error
}

func (f fakeCompleter) Complete(context.Context, string, []ChatMessage) (string, error) {
	return f.text, f.err
}

func TestEngineDeterministicInspect(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "AIFT-OS")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("inspect")
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(store, root, fakeCompleter{text: "unused"})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := engine.HandleMessage(context.Background(), session.ID, "/inspect AIFT-OS")
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Turns) != 2 {
		t.Fatalf("turns=%d", len(updated.Turns))
	}
	if !strings.Contains(updated.Turns[1].Content, "federation-kernel") {
		t.Fatalf("answer=%s", updated.Turns[1].Content)
	}
}

type captureCompleter struct {
	mu       sync.Mutex
	messages []ChatMessage
}

func (c *captureCompleter) Complete(_ context.Context, _ string, messages []ChatMessage) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append([]ChatMessage(nil), messages...)
	return "continued", nil
}

func TestEngineIncludesConversationHistory(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "AIFT-OS", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("history")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AppendTurn(session.ID, Turn{Role: "user", Content: "first"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AppendTurn(session.ID, Turn{Role: "assistant", Content: "reply"}); err != nil {
		t.Fatal(err)
	}
	completer := &captureCompleter{}
	engine, err := NewEngine(store, root, completer)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.HandleMessage(context.Background(), session.ID, "next"); err != nil {
		t.Fatal(err)
	}
	if len(completer.messages) != 3 || completer.messages[0].Content != "first" || completer.messages[2].Content != "next" {
		t.Fatalf("messages=%#v", completer.messages)
	}
}

func TestEngineForgeInspection(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "AIFT-OS", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("forge")
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(store, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := engine.HandleMessage(context.Background(), session.ID, "/forge")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(updated.Turns[len(updated.Turns)-1].Content, "No persisted Forge mission") {
		t.Fatalf("turn=%#v", updated.Turns)
	}
}

func TestEngineForgeFailureLeavesNoDanglingTurn(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "AIFT-Forge", ".forge", "mission.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("forge")
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(store, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.HandleMessage(context.Background(), session.ID, "/forge"); err == nil {
		t.Fatal("expected forge read error")
	}
	loaded, err := store.GetSession(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Turns) != 0 {
		t.Fatalf("turns=%#v", loaded.Turns)
	}
}

func TestEngineTruthfulNoCompleterReason(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "AIFT-OS", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("degraded")
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(store, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := engine.HandleMessage(context.Background(), session.ID, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(updated.Turns[1].Content, "No local model endpoint is configured") {
		t.Fatalf("answer=%q", updated.Turns[1].Content)
	}
}

func TestEngineSerializesConcurrentExchanges(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "AIFT-OS", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("concurrent")
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(store, root, fakeCompleter{text: "ok"})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, message := range []string{"one", "two"} {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			_, err := engine.HandleMessage(context.Background(), session.ID, m)
			errs <- err
		}(message)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	loaded, err := store.GetSession(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Turns) != 4 {
		t.Fatalf("turns=%#v", loaded.Turns)
	}
	for i := 0; i < len(loaded.Turns); i += 2 {
		if loaded.Turns[i].Role != "user" || loaded.Turns[i+1].Role != "assistant" {
			t.Fatalf("turn order=%#v", loaded.Turns)
		}
	}
}

func TestEngineInferenceFailureDegrades(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "AIFT-OS", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("offline")
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(store, root, fakeCompleter{err: errors.New("offline")})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := engine.HandleMessage(context.Background(), session.ID, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(updated.Turns[1].Content, "endpoint is not available") {
		t.Fatalf("answer=%q", updated.Turns[1].Content)
	}
}
