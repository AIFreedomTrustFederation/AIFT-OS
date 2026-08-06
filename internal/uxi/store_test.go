package uxi

import "testing"

func TestStoreSessionRoundTrip(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("Test")
	if err != nil {
		t.Fatal(err)
	}
	updated, err := store.AppendTurn(session.ID, Turn{Role: "user", Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Turns) != 1 || updated.Turns[0].Content != "hello" {
		t.Fatalf("unexpected turns: %#v", updated.Turns)
	}
	loaded, err := store.GetSession(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Title != "Test" {
		t.Fatalf("title = %q", loaded.Title)
	}
	events, err := store.ListEvents(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("events = %d", len(events))
	}
}

func TestStoreRejectsTraversalID(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetSession("../secret"); err != ErrNotFound {
		t.Fatalf("err = %v", err)
	}
}
