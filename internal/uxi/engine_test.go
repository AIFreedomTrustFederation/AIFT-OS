package uxi

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeCompleter struct{ text string }

func (f fakeCompleter) Complete(context.Context, string, []ChatMessage) (string, error) {
	return f.text, nil
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
		t.Fatalf("turns = %d", len(updated.Turns))
	}
	if !strings.Contains(updated.Turns[1].Content, "federation-kernel") {
		t.Fatalf("answer = %s", updated.Turns[1].Content)
	}
}

type captureCompleter struct{ messages []ChatMessage }

func (c *captureCompleter) Complete(_ context.Context, _ string, messages []ChatMessage) (string, error) {
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
	if len(completer.messages) != 3 {
		t.Fatalf("messages = %#v", completer.messages)
	}
	if completer.messages[0].Content != "first" || completer.messages[2].Content != "next" {
		t.Fatalf("messages = %#v", completer.messages)
	}
}
