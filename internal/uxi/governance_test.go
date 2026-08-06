package uxi

import (
	"errors"
	"testing"
)

func TestGovernanceLifecycleDoesNotExecute(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("governance")
	if err != nil {
		t.Fatal(err)
	}
	session, err = store.AddPlan(session.ID, Plan{Objective: "Inspect Forge", Steps: []PlanStep{{Title: "Read mission evidence"}}})
	if err != nil {
		t.Fatal(err)
	}
	session, err = store.ProposeAction(session.ID, Action{Kind: "forge.patch.apply", Target: "AIFT-Forge", Risk: "moderate", ApprovalRequired: true})
	if err != nil {
		t.Fatal(err)
	}
	action := session.Actions[0]
	session, err = store.DecideAction(session.ID, action.ID, "approved", "human-local-operator")
	if err != nil {
		t.Fatal(err)
	}
	if session.Actions[0].Status != StatusApproved || len(session.Approvals) != 1 || len(session.Jobs) != 0 {
		t.Fatalf("session=%#v", session)
	}
}

func TestGovernanceRejectsRepeatDecision(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("decision")
	if err != nil {
		t.Fatal(err)
	}
	session, err = store.ProposeAction(session.ID, Action{Kind: "read.status", Target: "AIFT-OS"})
	if err != nil {
		t.Fatal(err)
	}
	action := session.Actions[0]
	if _, err := store.DecideAction(session.ID, action.ID, "rejected", "operator"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.DecideAction(session.ID, action.ID, "approved", "operator"); err == nil {
		t.Fatal("expected repeat decision to fail")
	}
}

func TestGovernanceRecoversPlanEventFailure(t *testing.T) {
	store, session := newGovernanceStore(t)
	injectOneEventFailure(store)
	updated, err := store.AddPlan(session.ID, Plan{Objective: "Plan once", Steps: []PlanStep{{Title: "Step"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Plans) != 1 {
		t.Fatalf("plans=%#v", updated.Plans)
	}
	assertEventCount(t, store, "plan.proposed", 1)
}

func TestGovernanceRecoversActionEventFailure(t *testing.T) {
	store, session := newGovernanceStore(t)
	injectOneEventFailure(store)
	updated, err := store.ProposeAction(session.ID, Action{Kind: "read.status", Target: "AIFT-OS"})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Actions) != 1 {
		t.Fatalf("actions=%#v", updated.Actions)
	}
	assertEventCount(t, store, "action.proposed", 1)
}

func TestGovernanceRecoversDecisionEventFailure(t *testing.T) {
	store, session := newGovernanceStore(t)
	session, err := store.ProposeAction(session.ID, Action{Kind: "read.status", Target: "AIFT-OS"})
	if err != nil {
		t.Fatal(err)
	}
	injectOneEventFailure(store)
	updated, err := store.DecideAction(session.ID, session.Actions[0].ID, "approved", "operator")
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Approvals) != 1 {
		t.Fatalf("approvals=%#v", updated.Approvals)
	}
	assertEventCount(t, store, "action.decision", 1)
}

func newGovernanceStore(t *testing.T) (*Store, Session) {
	t.Helper()
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession("governance")
	if err != nil {
		t.Fatal(err)
	}
	return store, session
}

func injectOneEventFailure(store *Store) {
	failed := false
	store.beforeEventAppend = func() error {
		if !failed {
			failed = true
			return errors.New("injected")
		}
		return nil
	}
}

func assertEventCount(t *testing.T, store *Store, kind string, want int) {
	t.Helper()
	events, err := store.ListEvents(0)
	if err != nil {
		t.Fatal(err)
	}
	got := 0
	for _, event := range events {
		if event.Kind == kind {
			got++
		}
	}
	if got != want {
		t.Fatalf("%s events=%d", kind, got)
	}
}
