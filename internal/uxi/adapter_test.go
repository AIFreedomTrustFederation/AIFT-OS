package uxi

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultAdaptersAreReadOnly(t *testing.T) {
	registry, err := NewDefaultAdapterRegistry(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	descriptors := registry.List()
	if len(descriptors) != 3 {
		t.Fatalf("descriptors = %#v", descriptors)
	}
	for _, descriptor := range descriptors {
		if descriptor.Mutating {
			t.Fatalf("default adapter must be read-only: %#v", descriptor)
		}
	}
}

func TestInvokeRepositoryInspectCreatesArtifact(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "AIFT-OS", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(store, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("invoke")
	if err != nil {
		t.Fatal(err)
	}
	session, err = store.ProposeAction(session.ID, Action{Kind: "repository.inspect", Target: "AIFT-OS", Risk: "low"})
	if err != nil {
		t.Fatal(err)
	}
	session, err = engine.InvokeAction(context.Background(), session.ID, session.Actions[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if session.Actions[0].Status != StatusSucceeded || len(session.Jobs) != 1 || len(session.Artifacts) != 1 {
		t.Fatalf("session = %#v", session)
	}
	if _, err := store.ReadArtifact(session.Artifacts[0].ID); err != nil {
		t.Fatal(err)
	}
}

func TestInvokeRequiresApprovalWhenDeclared(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "AIFT-OS", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	store, _ := NewStore(t.TempDir())
	engine, _ := NewEngine(store, root, nil)
	session, _ := store.CreateSession("approval")
	session, _ = store.ProposeAction(session.ID, Action{Kind: "repository.inspect", Target: "AIFT-OS", Risk: "low", ApprovalRequired: true})
	actionID := session.Actions[0].ID
	if _, err := engine.InvokeAction(context.Background(), session.ID, actionID); err == nil {
		t.Fatal("expected approval requirement")
	}
	if _, err := store.DecideAction(session.ID, actionID, "approved", "operator"); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.InvokeAction(context.Background(), session.ID, actionID); err != nil {
		t.Fatal(err)
	}
}

type mutatingTestAdapter struct{}

func (mutatingTestAdapter) Descriptor() AdapterDescriptor {
	return AdapterDescriptor{Kind: "test.mutate", Version: "v1", Mutating: true, Risk: "high"}
}
func (mutatingTestAdapter) Invoke(context.Context, AdapterRequest) (AdapterResult, error) {
	return AdapterResult{}, nil
}

func TestMutatingAdapterCannotRun(t *testing.T) {
	store, _ := NewStore(t.TempDir())
	engine, _ := NewEngine(store, t.TempDir(), nil)
	registry, err := NewAdapterRegistry(mutatingTestAdapter{})
	if err != nil {
		t.Fatal(err)
	}
	engine.Adapters = registry
	session, _ := store.CreateSession("blocked")
	session, _ = store.ProposeAction(session.ID, Action{Kind: "test.mutate", Target: "anything", Risk: "high"})
	_, err = engine.InvokeAction(context.Background(), session.ID, session.Actions[0].ID)
	if !errors.Is(err, ErrMutationDisabled) {
		t.Fatalf("err = %v", err)
	}
	loaded, _ := store.GetSession(session.ID)
	if len(loaded.Jobs) != 0 {
		t.Fatalf("mutating adapter created job: %#v", loaded.Jobs)
	}
}

type failingTestAdapter struct{}

func (failingTestAdapter) Descriptor() AdapterDescriptor {
	return AdapterDescriptor{Kind: "test.fail", Version: "v1", Mutating: false, Risk: "low"}
}
func (failingTestAdapter) Invoke(context.Context, AdapterRequest) (AdapterResult, error) {
	return AdapterResult{}, errors.New("expected failure")
}

func TestAdapterFailureIsPersisted(t *testing.T) {
	store, _ := NewStore(t.TempDir())
	engine, _ := NewEngine(store, t.TempDir(), nil)
	registry, _ := NewAdapterRegistry(failingTestAdapter{})
	engine.Adapters = registry
	session, _ := store.CreateSession("failure")
	session, _ = store.ProposeAction(session.ID, Action{Kind: "test.fail", Target: "anything", Risk: "low"})
	updated, err := engine.InvokeAction(context.Background(), session.ID, session.Actions[0].ID)
	if err == nil {
		t.Fatal("expected adapter error")
	}
	if updated.Actions[0].Status != StatusFailed || updated.Jobs[0].Status != StatusFailed {
		t.Fatalf("updated = %#v", updated)
	}
}
