package uxi

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	root string
	mu   sync.RWMutex
}

func NewStore(root string) (*Store, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("store root is required")
	}
	for _, dir := range []string{root, filepath.Join(root, "sessions"), filepath.Join(root, "events")} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("create store directory %s: %w", dir, err)
		}
	}
	return &Store{root: root}, nil
}

func (s *Store) Root() string { return s.root }

func (s *Store) CreateSession(title string) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if strings.TrimSpace(title) == "" {
		title = "New workspace"
	}
	session := Session{
		ID:        newID("ses"),
		Title:     strings.TrimSpace(title),
		CreatedAt: now,
		UpdatedAt: now,
		Context:   SessionContext{Mode: "inspect"},
		Turns:     []Turn{},
		Plans:     []Plan{},
		Actions:   []Action{},
		Approvals: []Approval{},
		Jobs:      []Job{},
		Artifacts: []Artifact{},
	}
	if err := s.writeSessionLocked(session); err != nil {
		return Session{}, err
	}
	_ = s.appendEventLocked(Event{
		ID:        newID("evt"),
		SessionID: session.ID,
		Kind:      "session.created",
		Status:    StatusSucceeded,
		Message:   "Session created",
		CreatedAt: now,
	})
	return session, nil
}

func (s *Store) ListSessions() ([]Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	matches, err := filepath.Glob(filepath.Join(s.root, "sessions", "*.json"))
	if err != nil {
		return nil, err
	}
	sessions := make([]Session, 0, len(matches))
	for _, path := range matches {
		session, err := readSession(path)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt) })
	return sessions, nil
}

func (s *Store) GetSession(id string) (Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getSessionLocked(id)
}

func (s *Store) AppendTurn(sessionID string, turn Turn) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return Session{}, err
	}
	if strings.TrimSpace(turn.Role) == "" || strings.TrimSpace(turn.Content) == "" {
		return Session{}, errors.New("turn role and content are required")
	}
	if turn.ID == "" {
		turn.ID = newID("turn")
	}
	if turn.CreatedAt.IsZero() {
		turn.CreatedAt = time.Now().UTC()
	}
	session.Turns = append(session.Turns, turn)
	session.UpdatedAt = turn.CreatedAt
	if err := s.writeSessionLocked(session); err != nil {
		return Session{}, err
	}
	_ = s.appendEventLocked(Event{
		ID:        newID("evt"),
		SessionID: session.ID,
		Kind:      "turn.appended",
		Status:    StatusSucceeded,
		Message:   turn.Role + " turn appended",
		Data:      map[string]any{"turn_id": turn.ID, "role": turn.Role},
		CreatedAt: turn.CreatedAt,
	})
	return session, nil
}

func (s *Store) ListEvents(limit int) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("decode event: %w", err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if limit > 0 && len(events) > limit {
		events = events[len(events)-limit:]
	}
	return events, nil
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
	return session, nil
}

func (s *Store) writeSessionLocked(session Session) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.root, "sessions", session.ID+".json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (s *Store) appendEventLocked(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	path := filepath.Join(s.root, "events", "events.jsonl")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(data, '\n'))
	return err
}

func newID(prefix string) string {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(raw[:])
}

func validID(id string) bool {
	if id == "" || strings.ContainsAny(id, `/\\`) || strings.Contains(id, "..") {
		return false
	}
	return true
}
