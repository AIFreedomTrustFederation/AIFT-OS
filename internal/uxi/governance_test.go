package uxi

import "testing"

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
	if session.Plans[0].Status != StatusProposed {
		t.Fatalf("plan = %#v", session.Plans[0])
	}
	session, err = store.ProposeAction(session.ID, Action{Kind: "forge.patch.apply", Target: "AIFT-Forge", Risk: "moderate", ApprovalRequired: true})
	if err != nil {
		t.Fatal(err)
	}
	action := session.Actions[0]
	if action.Status != StatusAwaitingApproval {
		t.Fatalf("action = %#v", action)
	}
	session, err = store.DecideAction(session.ID, action.ID, "approved", "human-local-operator")
	if err != nil {
		t.Fatal(err)
	}
	if session.Actions[0].Status != StatusApproved || len(session.Approvals) != 1 {
		t.Fatalf("session = %#v", session)
	}
	if len(session.Jobs) != 0 {
		t.Fatalf("approval must not execute a job: %#v", session.Jobs)
	}
}

func TestGovernanceRejectsRepeatDecision(t *testing.T) {
	store, _ := NewStore(t.TempDir())
	session, _ := store.CreateSession("decision")
	session, _ = store.ProposeAction(session.ID, Action{Kind: "read.status", Target: "AIFT-OS"})
	action := session.Actions[0]
	if _, err := store.DecideAction(session.ID, action.ID, "rejected", "operator"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.DecideAction(session.ID, action.ID, "approved", "operator"); err == nil {
		t.Fatal("expected repeat decision to fail")
	}
}
