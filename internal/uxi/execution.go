package uxi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (e *Engine) InvokeAction(ctx context.Context, sessionID, actionID string) (Session, error) {
	if e.Adapters == nil {
		return Session{}, errors.New("adapter registry is unavailable")
	}
	session, err := e.Store.GetSession(sessionID)
	if err != nil {
		return Session{}, err
	}
	action, err := findAction(session.Actions, actionID)
	if err != nil {
		return Session{}, err
	}
	adapter, ok := e.Adapters.Lookup(action.Kind)
	if !ok {
		return Session{}, fmt.Errorf("no adapter registered for action kind %q", action.Kind)
	}
	descriptor := adapter.Descriptor()
	if descriptor.Mutating {
		return Session{}, ErrMutationDisabled
	}
	if action.ApprovalRequired {
		if action.Status != StatusApproved {
			return Session{}, fmt.Errorf("action %s requires approval", action.ID)
		}
	} else if action.Status != StatusProposed && action.Status != StatusApproved {
		return Session{}, fmt.Errorf("action %s cannot be invoked from status %s", action.ID, action.Status)
	}

	session, job, err := e.Store.startJob(sessionID, actionID, descriptor.Kind)
	if err != nil {
		return Session{}, err
	}
	result, invokeErr := adapter.Invoke(ctx, AdapterRequest{
		SessionID: sessionID, ActionID: actionID, Target: action.Target, Parameters: action.Parameters,
	})
	if invokeErr != nil {
		return e.Store.finishJob(sessionID, actionID, job.ID, "", invokeErr)
	}
	_, artifact, artifactErr := e.Store.writeArtifact(sessionID, "adapter-result", descriptor.Kind+" result", result)
	if artifactErr != nil {
		return e.Store.finishJob(sessionID, actionID, job.ID, "", artifactErr)
	}
	return e.Store.finishJob(sessionID, actionID, job.ID, artifact.ID, nil)
}

func findAction(actions []Action, actionID string) (Action, error) {
	for _, action := range actions {
		if action.ID == actionID {
			return action, nil
		}
	}
	return Action{}, ErrNotFound
}

func (s *Store) startJob(sessionID, actionID, adapterKind string) (Session, Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return Session{}, Job{}, err
	}
	index := -1
	for i := range session.Actions {
		if session.Actions[i].ID == actionID {
			index = i
			break
		}
	}
	if index < 0 {
		return Session{}, Job{}, ErrNotFound
	}
	now := time.Now().UTC()
	session.Actions[index].Status = StatusRunning
	session.Actions[index].UpdatedAt = now
	job := Job{ID: newID("job"), ActionID: actionID, AdapterKind: adapterKind, Status: StatusRunning, StartedAt: &now}
	session.Jobs = append(session.Jobs, job)
	session.UpdatedAt = now
	event := Event{ID: newID("evt"), SessionID: sessionID, Kind: "job.started", Status: StatusRunning, Message: "Read-only adapter job started", Data: map[string]any{"job_id": job.ID, "action_id": actionID, "adapter": adapterKind}, CreatedAt: now}
	if err := s.commitSessionEventLocked(session, event, "job.started:"+job.ID); err != nil {
		return Session{}, Job{}, err
	}
	return session, job, nil
}

func (s *Store) finishJob(sessionID, actionID, jobID, artifactID string, jobErr error) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return Session{}, err
	}
	actionIndex, jobIndex := -1, -1
	for i := range session.Actions {
		if session.Actions[i].ID == actionID {
			actionIndex = i
			break
		}
	}
	for i := range session.Jobs {
		if session.Jobs[i].ID == jobID {
			jobIndex = i
			break
		}
	}
	if actionIndex < 0 || jobIndex < 0 {
		return Session{}, ErrNotFound
	}
	now := time.Now().UTC()
	status := StatusSucceeded
	if jobErr != nil {
		status = StatusFailed
	}
	session.Actions[actionIndex].Status = status
	session.Actions[actionIndex].UpdatedAt = now
	session.Jobs[jobIndex].Status = status
	session.Jobs[jobIndex].EndedAt = &now
	session.Jobs[jobIndex].ResultArtifactID = artifactID
	if jobErr != nil {
		session.Jobs[jobIndex].Error = jobErr.Error()
	}
	session.UpdatedAt = now
	data := map[string]any{"job_id": jobID, "action_id": actionID}
	if artifactID != "" {
		data["artifact_id"] = artifactID
	}
	if jobErr != nil {
		data["error"] = jobErr.Error()
	}
	event := Event{ID: newID("evt"), SessionID: sessionID, Kind: "job.finished", Status: status, Message: "Read-only adapter job finished", Data: data, CreatedAt: now}
	if err := s.commitSessionEventLocked(session, event, "job.finished:"+jobID); err != nil {
		return Session{}, errors.Join(jobErr, err)
	}
	return session, jobErr
}

func (s *Store) writeArtifact(sessionID, kind, name string, value any) (Session, Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return Session{}, Artifact{}, err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return Session{}, Artifact{}, err
	}
	id := newID("art")
	relativePath := filepath.Join("artifacts", id+".json")
	path := filepath.Join(s.root, relativePath)
	tmp := path + ".tmp"
	payload := append(data, '\n')
	if err := os.WriteFile(tmp, payload, 0o600); err != nil {
		return Session{}, Artifact{}, err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return Session{}, Artifact{}, err
	}
	digest := sha256.Sum256(payload)
	now := time.Now().UTC()
	artifact := Artifact{ID: id, Kind: strings.TrimSpace(kind), Name: strings.TrimSpace(name), Path: filepath.ToSlash(relativePath), Digest: hex.EncodeToString(digest[:]), CreatedAt: now}
	session.Artifacts = append(session.Artifacts, artifact)
	session.UpdatedAt = now
	if err := s.writeSessionLocked(session); err != nil {
		_ = os.Remove(path)
		return Session{}, Artifact{}, err
	}
	_ = s.appendEventLocked(Event{ID: newID("evt"), SessionID: sessionID, Kind: "artifact.created", Status: StatusSucceeded, Message: "Adapter result artifact created", Data: map[string]any{"artifact_id": id, "path": artifact.Path}, CreatedAt: now})
	return session, artifact, nil
}

func (s *Store) ReadArtifact(id string) (json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !validID(id) {
		return nil, ErrNotFound
	}
	data, err := os.ReadFile(filepath.Join(s.root, "artifacts", id+".json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !json.Valid(data) {
		return nil, errors.New("artifact JSON is invalid")
	}
	return json.RawMessage(data), nil
}
