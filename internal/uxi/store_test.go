package uxi

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestAppendExchangeIsAtomic(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("exchange")
	if err != nil {
		t.Fatal(err)
	}
	updated, err := store.AppendExchange(session.ID, Turn{Role: "user", Content: "hello"}, Turn{Role: "assistant", Content: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Turns) != 2 || updated.Turns[0].Role != "user" || updated.Turns[1].Role != "assistant" {
		t.Fatalf("turns=%#v", updated.Turns)
	}
}

func TestListSessionsSkipsCorruptFileAndAudits(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateSession("healthy"); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(store.Root(), "sessions", "ses_bad.json")
	if err := os.WriteFile(bad, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 {
		t.Fatalf("sessions=%#v", sessions)
	}
	events, err := store.ListEvents(0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, event := range events {
		if event.Kind == "store.session_corrupt" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing corrupt-session audit event")
	}
}

func TestListEventsSupportsLargeRecords(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	event := Event{ID: newID("evt"), Kind: "large", Status: StatusSucceeded, Message: strings.Repeat("x", 128<<10)}
	store.mu.Lock()
	err = store.appendEventLocked(event)
	store.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	events, err := store.ListEvents(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || len(events[0].Message) != len(event.Message) {
		t.Fatalf("events=%d", len(events))
	}
}

func TestStoreRecoversEventFailureWithoutDuplicate(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("recovery")
	if err != nil {
		t.Fatal(err)
	}
	failed := false
	store.beforeEventAppend = func() error {
		if !failed {
			failed = true
			return errors.New("injected")
		}
		return nil
	}
	updated, err := store.AppendTurn(session.ID, Turn{Role: "user", Content: "once"})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Turns) != 1 {
		t.Fatalf("turns=%d", len(updated.Turns))
	}
	events, err := store.ListEvents(0)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, event := range events {
		if event.Kind == "turn.appended" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("turn events=%d", count)
	}
}

func TestOptionalTimesAreOmitted(t *testing.T) {
	data, err := json.Marshal(struct {
		Approval Approval `json:"approval"`
		Job      Job      `json:"job"`
	}{})
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, "expires_at") || strings.Contains(text, "started_at") || strings.Contains(text, "ended_at") {
		t.Fatalf("json=%s", text)
	}
}
