package uxi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ErrNotFound is returned when an allowlisted local identifier cannot be found.
var ErrNotFound = errors.New("not found")

type pendingCommit struct {
	ID           string    `json:"id"`
	OperationKey string    `json:"operation_key,omitempty"`
	Session      Session   `json:"session"`
	Event        Event     `json:"event"`
	CreatedAt    time.Time `json:"created_at"`
}

// Store persists private UXI sessions, audit events, and recoverable commits.
type Store struct {
	root string
	mu   sync.RWMutex

	// Test-only failure injection hooks. Production callers leave these nil.
	beforeSessionWrite func() error
	beforeEventAppend  func() error
	recovered          map[string]struct{}
	appliedEventIDs    map[string]struct{}
}

// NewStore creates a private local store and recovers unfinished commits.
func NewStore(root string) (*Store, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("store root is required")
	}
	for _, dir := range []string{root, filepath.Join(root, "sessions"), filepath.Join(root, "events"), filepath.Join(root, "transactions")} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("create store directory %s: %w", dir, err)
		}
	}
	store := &Store{
		root: root,
		recovered: map[string]struct{}{},
		appliedEventIDs: map[string]struct{}{},
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	events, err := store.listEventsLocked(0)
	if err != nil {
		return nil, fmt.Errorf("load UXI event index: %w", err)
	}
	for _, event := range events {
		store.appliedEventIDs[event.ID] = struct{}{}
	}
	if _, err := store.recoverPendingLocked(); err != nil {
		return nil, fmt.Errorf("recover UXI commits: %w", err)
	}
	return store, nil
}

// Root returns the configured private data directory.
func (s *Store) Root() string { return s.root }

// CreateSession creates a new workspace and its required audit event together.
func (s *Store) CreateSession(title string) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.recoverPendingLocked(); err != nil {
		return Session{}, err
	}

	now := time.Now().UTC()
	if strings.TrimSpace(title) == "" {
		title = "New workspace"
	}
	session := Session{
		ID: newID("ses"), Title: strings.TrimSpace(title), CreatedAt: now, UpdatedAt: now,
		Context: SessionContext{Mode: "inspect"}, Turns: []Turn{}, Plans: []Plan{}, Actions: []Action{},
		Approvals: []Approval{}, Jobs: []Job{}, Artifacts: []Artifact{},
	}
	event := Event{ID: newID("evt"), SessionID: session.ID, Kind: "session.created", Status: StatusSucceeded, Message: "Session created", CreatedAt: now}
	if err := s.commitSessionEventLocked(session, event, "session.created:"+session.ID); err != nil {
		return Session{}, err
	}
	return session, nil
}

// ListSessions returns healthy sessions newest first and records corrupt files as audit events.
func (s *Store) ListSessions() ([]Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.recoverPendingLocked(); err != nil {
		return nil, err
	}

	matches, err := filepath.Glob(filepath.Join(s.root, "sessions", "*.json"))
	if err != nil {
		return nil, err
	}
	sessions := make([]Session, 0, len(matches))
	for _, path := range matches {
		session, readErr := readSession(path)
		if readErr != nil {
			event := Event{
				ID: stableID("evt_corrupt", path+":"+readErr.Error()), Kind: "store.session_corrupt", Status: StatusBlocked,
				Message: "Unreadable session file skipped", Data: map[string]any{"path": path, "error": readErr.Error()}, CreatedAt: time.Now().UTC(),
			}
			if _, exists := s.appliedEventIDs[event.ID]; !exists {
				_ = s.appendEventLocked(event)
			}
			continue
		}
		sessions = append(sessions, session)
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt) })
	return sessions, nil
}

// GetSession loads one allowlisted session identifier.
func (s *Store) GetSession(id string) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.recoverPendingLocked(); err != nil {
		return Session{}, err
	}
	return s.getSessionLocked(id)
}

// AppendTurn appends one validated turn and its audit event atomically.
func (s *Store) AppendTurn(sessionID string, turn Turn) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.recoverPendingLocked(); err != nil {
		return Session{}, err
	}

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return Session{}, err
	}
	turn, err = normalizeTurn(turn)
	if err != nil {
		return Session{}, err
	}
	session.Turns = append(session.Turns, turn)
	session.UpdatedAt = turn.CreatedAt
	event := Event{
		ID: newID("evt"), SessionID: session.ID, Kind: "turn.appended", Status: StatusSucceeded,
		Message: turn.Role + " turn appended", Data: map[string]any{"turn_id": turn.ID, "role": turn.Role}, CreatedAt: turn.CreatedAt,
	}
	if err := s.commitSessionEventLocked(session, event, "turn.appended:"+turn.ID); err != nil {
		return Session{}, err
	}
	return session, nil
}

// AppendExchange persists a user turn and its assistant response in one session commit.
func (s *Store) AppendExchange(sessionID string, userTurn, assistantTurn Turn) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.recoverPendingLocked(); err != nil {
		return Session{}, err
	}

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return Session{}, err
	}
	userTurn, err = normalizeTurn(userTurn)
	if err != nil {
		return Session{}, err
	}
	assistantTurn, err = normalizeTurn(assistantTurn)
	if err != nil {
		return Session{}, err
	}
	if userTurn.Role != "user" || assistantTurn.Role != "assistant" {
		return Session{}, errors.New("exchange requires user then assistant roles")
	}
	if assistantTurn.CreatedAt.Before(userTurn.CreatedAt) {
		assistantTurn.CreatedAt = userTurn.CreatedAt
	}
	session.Turns = append(session.Turns, userTurn, assistantTurn)
	session.UpdatedAt = assistantTurn.CreatedAt
	event := Event{
		ID: newID("evt"), SessionID: session.ID, Kind: "conversation.exchange", Status: StatusSucceeded,
		Message: "User and assistant turns committed", Data: map[string]any{"user_turn_id": userTurn.ID, "assistant_turn_id": assistantTurn.ID}, CreatedAt: assistantTurn.CreatedAt,
	}
	if err := s.commitSessionEventLocked(session, event, "conversation.exchange:"+userTurn.ID); err != nil {
		return Session{}, err
	}
	return session, nil
}

// ListEvents streams append-only JSON values without a line-size limit.
func (s *Store) ListEvents(limit int) ([]Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.recoverPendingLocked(); err != nil {
		return nil, err
	}
	return s.listEventsLocked(limit)
}

func (s *Store) listEventsLocked(limit int) ([]Event, error) {
	path := filepath.Join(s.root, "events", "events.jsonl")
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return []Event{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []Event
	decoder := json.NewDecoder(file)
	for {
		var event Event
		if err := decoder.Decode(&event); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("decode event: %w", err)
		}
		events = append(events, event)
	}
	if limit > 0 && len(events) > limit {
		events = events[len(events)-limit:]
	}
	return events, nil
}

func normalizeTurn(turn Turn) (Turn, error) {
	turn.Role = strings.TrimSpace(turn.Role)
	turn.Content = strings.TrimSpace(turn.Content)
	if turn.Role == "" || turn.Content == "" {
		return Turn{}, errors.New("turn role and content are required")
	}
	if turn.Role != "user" && turn.Role != "assistant" && turn.Role != "system" && turn.Role != "tool" {
		return Turn{}, fmt.Errorf("unsupported turn role %q", turn.Role)
	}
	if turn.ID == "" {
		turn.ID = newID("turn")
	}
	if turn.CreatedAt.IsZero() {
		turn.CreatedAt = time.Now().UTC()
	} else {
		turn.CreatedAt = turn.CreatedAt.UTC()
	}
	return turn, nil
}

func (s *Store) getSessionLocked(id string) (Session, error) {
	if !validID(id) {
		return Session{}, ErrNotFound
	}
	path := filepath.Join(s.root, "sessions", id+".json")
	session, err := readSession(path)
	if errors.Is(err, os.ErrNotExist) {
		return Session{}, ErrNotFound
	}
	return session, err
}

func readSession(path string) (Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Session{}, err
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return Session{}, fmt.Errorf("decode session %s: %w", path, err)
	}
	if !validID(session.ID) {
		return Session{}, fmt.Errorf("decode session %s: invalid session id", path)
	}
	return session, nil
}

func (s *Store) commitSessionEventLocked(session Session, event Event, operationKey string) error {
	pending := pendingCommit{ID: event.ID, OperationKey: operationKey, Session: session, Event: event, CreatedAt: time.Now().UTC()}
	if err := s.writePendingLocked(pending); err != nil {
		return err
	}
	if err := s.applyPendingLocked(pending); err != nil {
		// Retry once immediately. One-shot failures recover without exposing a false failure.
		if retryErr := s.applyPendingLocked(pending); retryErr == nil {
			return nil
		} else {
			pendingPath := filepath.Join(s.root, "transactions", pending.ID+".json")
			if removeErr := os.Remove(pendingPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				return errors.Join(err, retryErr, removeErr)
			}
			return errors.Join(err, retryErr)
		}
	}
	return nil
}

func (s *Store) writePendingLocked(pending pendingCommit) error {
	return writeAtomicJSON(filepath.Join(s.root, "transactions", pending.ID+".json"), pending, 0o600)
}

func (s *Store) applyPendingLocked(pending pendingCommit) error {
	if err := s.writeSessionLocked(pending.Session); err != nil {
		return err
	}
	if _, exists := s.appliedEventIDs[pending.Event.ID]; !exists {
		if err := s.appendEventLocked(pending.Event); err != nil {
			return err
		}
	}
	if err := os.Remove(filepath.Join(s.root, "transactions", pending.ID+".json")); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *Store) recoverPendingLocked() ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(s.root, "transactions", "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	var recovered []string
	for _, path := range matches {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			s.auditCorruptTransactionLocked(path, readErr)
			continue
		}
		var pending pendingCommit
		if err := json.Unmarshal(data, &pending); err != nil {
			s.auditCorruptTransactionLocked(path, err)
			continue
		}
		if err := s.applyPendingLocked(pending); err != nil {
			return recovered, err
		}
		recovered = append(recovered, pending.OperationKey)
		if pending.OperationKey != "" {
			s.recovered[pending.OperationKey] = struct{}{}
		}
	}
	return recovered, nil
}

func (s *Store) auditCorruptTransactionLocked(path string, cause error) {
	event := Event{
		ID: stableID("evt_corrupt_tx", path+":"+cause.Error()), Kind: "store.transaction_corrupt", Status: StatusBlocked,
		Message: "Unreadable transaction file skipped", Data: map[string]any{"path": path, "error": cause.Error()}, CreatedAt: time.Now().UTC(),
	}
	if _, exists := s.appliedEventIDs[event.ID]; !exists {
		_ = s.appendEventLocked(event)
	}
}

func (s *Store) consumeRecoveredLocked(operationKey string) bool {
	if operationKey == "" {
		return false
	}
	if _, ok := s.recovered[operationKey]; !ok {
		return false
	}
	delete(s.recovered, operationKey)
	return true
}

func (s *Store) writeSessionLocked(session Session) error {
	if s.beforeSessionWrite != nil {
		if err := s.beforeSessionWrite(); err != nil {
			return err
		}
	}
	session.UpdatedAt = session.UpdatedAt.UTC()
	return writeAtomicJSON(filepath.Join(s.root, "sessions", session.ID+".json"), session, 0o600)
}

func writeAtomicJSON(path string, value any, mode os.FileMode) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		_ = file.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (s *Store) appendEventLocked(event Event) error {
	if s.beforeEventAppend != nil {
		if err := s.beforeEventAppend(); err != nil {
			return err
		}
	}
	path := filepath.Join(s.root, "events", "events.jsonl")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(event); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	s.appliedEventIDs[event.ID] = struct{}{}
	return nil
}

func (s *Store) eventExistsLocked(id string) (bool, error) {
	_, exists := s.appliedEventIDs[id]
	return exists, nil
}

func newID(prefix string) string {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(raw[:])
}

func stableID(prefix, value string) string {
	digest := sha256.Sum256([]byte(value))
	return prefix + "_" + hex.EncodeToString(digest[:8])
}

func validID(id string) bool {
	if id == "" || len(id) > 128 || strings.ContainsAny(id, `/\\`) || strings.Contains(id, "..") {
		return false
	}
	for _, r := range id {
		if !(r == '_' || r == '-' || r == '.' || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z') {
			return false
		}
	}
	return true
}
