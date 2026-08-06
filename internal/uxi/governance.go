package uxi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// AddPlan validates and records a proposed plan without executing it.
func (s *Store) AddPlan(sessionID string, plan Plan) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	plan.Objective = strings.TrimSpace(plan.Objective)
	if plan.Objective == "" {
		return Session{}, errors.New("plan objective is required")
	}
	if len(plan.Steps) == 0 {
		return Session{}, errors.New("plan requires at least one step")
	}
	for i := range plan.Steps {
		plan.Steps[i].Title = strings.TrimSpace(plan.Steps[i].Title)
		if plan.Steps[i].Title == "" {
			return Session{}, fmt.Errorf("plan step %d title is required", i+1)
		}
	}
	opKey := governanceOperationKey("plan.proposed", sessionID, plan)
	if _, err := s.recoverPendingLocked(); err != nil {
		return Session{}, err
	}
	if s.consumeRecoveredLocked(opKey) {
		return s.getSessionLocked(sessionID)
	}

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return Session{}, err
	}
	for i := range plan.Steps {
		if plan.Steps[i].ID == "" {
			plan.Steps[i].ID = newID("step")
		}
		if plan.Steps[i].Status == "" {
			plan.Steps[i].Status = StatusProposed
		}
	}
	now := time.Now().UTC()
	plan.ID = newID("plan")
	plan.Status = StatusProposed
	plan.CreatedAt = now
	plan.UpdatedAt = now
	session.Plans = append(session.Plans, plan)
	session.UpdatedAt = now
	event := Event{
		ID: newID("evt"), SessionID: session.ID, Kind: "plan.proposed", Status: StatusProposed,
		Message: "Plan proposed", Data: map[string]any{"plan_id": plan.ID, "objective": plan.Objective, "operation_key": opKey}, CreatedAt: now,
	}
	if err := s.commitSessionEventLocked(session, event, opKey); err != nil {
		return Session{}, err
	}
	return session, nil
}

// ProposeAction validates and records an action proposal without invoking it.
func (s *Store) ProposeAction(sessionID string, action Action) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	action.Kind = strings.TrimSpace(action.Kind)
	action.Target = strings.TrimSpace(action.Target)
	if action.Kind == "" || action.Target == "" {
		return Session{}, errors.New("action kind and target are required")
	}
	if action.Risk == "" {
		action.Risk = "low"
	}
	action.Risk = strings.ToLower(strings.TrimSpace(action.Risk))
	if !validRisk(action.Risk) {
		return Session{}, fmt.Errorf("invalid action risk %q", action.Risk)
	}
	opKey := governanceOperationKey("action.proposed", sessionID, action)
	if _, err := s.recoverPendingLocked(); err != nil {
		return Session{}, err
	}
	if s.consumeRecoveredLocked(opKey) {
		return s.getSessionLocked(sessionID)
	}

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return Session{}, err
	}
	now := time.Now().UTC()
	action.ID = newID("act")
	if action.ApprovalRequired {
		action.Status = StatusAwaitingApproval
	} else {
		action.Status = StatusProposed
	}
	action.CreatedAt = now
	action.UpdatedAt = now
	session.Actions = append(session.Actions, action)
	session.UpdatedAt = now
	event := Event{
		ID: newID("evt"), SessionID: session.ID, Kind: "action.proposed", Status: action.Status,
		Message: "Action proposed", Data: map[string]any{"action_id": action.ID, "kind": action.Kind, "target": action.Target, "operation_key": opKey}, CreatedAt: now,
	}
	if err := s.commitSessionEventLocked(session, event, opKey); err != nil {
		return Session{}, err
	}
	return session, nil
}

// DecideAction records a human approval or rejection and never creates a job.
func (s *Store) DecideAction(sessionID, actionID, decision, actor string) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	decision = strings.ToLower(strings.TrimSpace(decision))
	actor = strings.TrimSpace(actor)
	if decision != "approved" && decision != "rejected" {
		return Session{}, errors.New("decision must be approved or rejected")
	}
	if actor == "" {
		return Session{}, errors.New("approval actor is required")
	}
	opKey := governanceOperationKey("action.decision", sessionID, map[string]string{"action_id": actionID, "decision": decision, "actor": actor})
	if _, err := s.recoverPendingLocked(); err != nil {
		return Session{}, err
	}
	if s.consumeRecoveredLocked(opKey) {
		return s.getSessionLocked(sessionID)
	}

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return Session{}, err
	}
	index := -1
	for i := range session.Actions {
		if session.Actions[i].ID == actionID {
			index = i
			break
		}
	}
	if index < 0 {
		return Session{}, ErrNotFound
	}
	action := &session.Actions[index]
	if action.Status != StatusProposed && action.Status != StatusAwaitingApproval {
		return Session{}, fmt.Errorf("action %s cannot be decided from status %s", action.ID, action.Status)
	}
	now := time.Now().UTC()
	if decision == "approved" {
		action.Status = StatusApproved
	} else {
		action.Status = StatusCancelled
	}
	action.UpdatedAt = now
	approval := Approval{ID: newID("apr"), ActionID: action.ID, Decision: decision, Actor: actor, CreatedAt: now}
	session.Approvals = append(session.Approvals, approval)
	session.UpdatedAt = now
	event := Event{
		ID: newID("evt"), SessionID: session.ID, Kind: "action.decision", Status: action.Status,
		Message: "Action decision recorded", Data: map[string]any{"action_id": action.ID, "decision": decision, "actor": actor, "operation_key": opKey}, CreatedAt: now,
	}
	if err := s.commitSessionEventLocked(session, event, opKey); err != nil {
		return Session{}, err
	}
	return session, nil
}

func governanceOperationKey(kind, sessionID string, value any) string {
	data, _ := json.Marshal(value)
	return stableID("op", kind+":"+sessionID+":"+string(data))
}

func validRisk(risk string) bool {
	switch strings.ToLower(strings.TrimSpace(risk)) {
	case "low", "moderate", "high", "critical":
		return true
	default:
		return false
	}
}
