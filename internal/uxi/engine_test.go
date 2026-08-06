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
