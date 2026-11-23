package uxihttp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/uxi"
)

const messageJobTimeout = 100 * time.Second

type messageJob struct {
	ID             string       `json:"id"`
	SessionID      string       `json:"session_id"`
	Status         string       `json:"status"`
	Error          string       `json:"error,omitempty"`
	Result         *uxi.Session `json:"session,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	FinishedAt     time.Time    `json:"finished_at,omitempty"`
	IdempotencyKey string       `json:"idempotency_key,omitempty"`
	RequestHash    string       `json:"request_hash,omitempty"`
}

type messageJobs struct {
	mu     sync.Mutex
	jobs   map[string]*messageJob
	active map[string]string
	slots  chan struct{}
	store  *messageJobStore
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	closed bool
}

func newMessageJobs() *messageJobs {
	ctx, cancel := context.WithCancel(context.Background())
	return &messageJobs{jobs: make(map[string]*messageJob), active: make(map[string]string), slots: make(chan struct{}, 2), ctx: ctx, cancel: cancel}
}

func (s *Server) handleAsyncMessage(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content string `json:"content"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	input.Content = strings.TrimSpace(input.Content)
	if input.Content == "" {
		writeError(w, http.StatusBadRequest, errors.New("message content is required"))
		return
	}
	if len(input.Content) > 4096 {
		writeError(w, http.StatusRequestEntityTooLarge, errors.New("message exceeds 4096 characters"))
		return
	}
	id := r.PathValue("id")
	if _, err := s.Engine.Store.GetSession(id); err != nil {
		if errors.Is(err, uxi.ErrNotFound) {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	q := s.messageJobs
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		writeError(w, http.StatusServiceUnavailable, errors.New("server shutting down"))
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if len(key) > 128 || strings.ContainsAny(key, "\r\n\t ") {
		q.mu.Unlock()
		writeError(w, http.StatusBadRequest, errors.New("invalid idempotency key"))
		return
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(input.Content)))
	if key != "" {
		for _, prior := range q.jobs {
			if prior.SessionID == id && prior.IdempotencyKey == key {
				snapshot := *prior
				q.mu.Unlock()
				if snapshot.RequestHash != hash {
					writeError(w, http.StatusConflict, errors.New("idempotency key reused with different content"))
					return
				}
				writeJSON(w, http.StatusAccepted, publicMessageJob(snapshot))
				return
			}
		}
	}
	now := time.Now()
	if existing := q.active[id]; existing != "" {
		q.mu.Unlock()
		writeError(w, http.StatusConflict,
			errors.New("a message is already running for this session"))
		return
	}
	// Reserve capacity before accepting work. No unbounded queue.
	select {
	case q.slots <- struct{}{}:
	default:
		q.mu.Unlock()
		writeError(w, http.StatusTooManyRequests, errors.New("local inference capacity reached; retry shortly"))
		return
	}
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		<-q.slots
		q.mu.Unlock()
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	job := &messageJob{ID: hex.EncodeToString(token), SessionID: id, Status: "running", CreatedAt: now.UTC(), IdempotencyKey: key, RequestHash: hash}
	if err := q.store.save(*job); err != nil {
		<-q.slots
		q.mu.Unlock()
		writeError(w, http.StatusInternalServerError, errors.New("unable to persist job"))
		return
	}
	q.jobs[job.ID] = job
	q.active[id] = job.ID
	q.wg.Add(1)
	q.mu.Unlock()
	go func() {
		defer q.wg.Done()
		defer func() { <-q.slots }()
		ctx, cancel := context.WithTimeout(q.ctx, messageJobTimeout)
		defer cancel()
		session, err := s.Engine.HandleMessageForJob(ctx, id, input.Content, job.ID)
		q.mu.Lock()
		defer q.mu.Unlock()
		if err != nil {
			job.Status = "failed"
			job.Error = "inference failed"
		} else {
			job.Status = "completed"
			job.Result = &session
		}
		job.FinishedAt = time.Now().UTC()
		if err := q.store.save(*job); err != nil {
			// Keep the original accepted record on disk and block new work on this session.
			job.Status = "interrupted"
			job.Error = "result persistence unavailable; manual reconciliation required"
			job.Result = nil
			job.FinishedAt = time.Time{}
			return
		}
		delete(q.active, id)
	}()
	writeJSON(w, http.StatusAccepted, publicMessageJob(*job))
}

func (s *Server) handleMessageJob(w http.ResponseWriter, r *http.Request) {
	q := s.messageJobs
	q.mu.Lock()
	job := q.jobs[r.PathValue("jobID")]
	if job == nil || job.SessionID != r.PathValue("id") {
		q.mu.Unlock()
		writeError(w, http.StatusNotFound, uxi.ErrNotFound)
		return
	}
	snapshot := *job
	q.mu.Unlock()
	writeJSON(w, http.StatusOK, publicMessageJob(snapshot))
}

// Shutdown stops accepting work, cancels active inference and waits for workers.
func (s *Server) Shutdown(ctx context.Context) error {
	q := s.messageJobs
	q.mu.Lock()
	q.closed = true
	q.cancel()
	q.mu.Unlock()
	done := make(chan struct{})
	go func() { q.wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// publicMessageJob excludes private retry metadata while preserving the durable record.
func publicMessageJob(job messageJob) messageJob {
 job.IdempotencyKey = ""
 job.RequestHash = ""
 return job
}
