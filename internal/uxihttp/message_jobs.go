package uxihttp

import (
 "context"
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "net/http"
 "strings"
 "sync"
 "time"

 "github.com/AIFreedomTrustFederation/AIFT-OS/internal/uxi"
)

const messageJobTimeout = 100 * time.Second
const messageJobRetention = 10 * time.Minute

type messageJob struct {
 ID string `json:"id"`
 SessionID string `json:"session_id"`
 RequestDigest [32]byte `json:"-"`
 Status string `json:"status"`
 Error string `json:"error,omitempty"`
 Result *uxi.Session `json:"session,omitempty"`
 CreatedAt time.Time `json:"created_at"`
 FinishedAt time.Time `json:"finished_at,omitempty"`
}

type messageJobs struct {
 mu sync.Mutex
 jobs map[string]*messageJob
 active map[string]string
 slots chan struct{}
}

func newMessageJobs() *messageJobs {
 return &messageJobs{jobs: make(map[string]*messageJob), active: make(map[string]string), slots: make(chan struct{}, 2)}
}

func (s *Server) handleAsyncMessage(w http.ResponseWriter, r *http.Request) {
 var input struct { Content string `json:"content"` }
 if err := decodeJSON(r, &input); err != nil { writeError(w, http.StatusBadRequest, err); return }
 input.Content = strings.TrimSpace(input.Content)
 if input.Content == "" { writeError(w, http.StatusBadRequest, errors.New("message content is required")); return }
 if len(input.Content) > 4096 { writeError(w, http.StatusRequestEntityTooLarge, errors.New("message exceeds 4096 characters")); return }
 id := r.PathValue("id")
 if _, err := s.Engine.Store.GetSession(id); err != nil {
  if errors.Is(err, uxi.ErrNotFound) { writeError(w, http.StatusNotFound, err) } else { writeError(w, http.StatusInternalServerError, err) }
  return
 }
 q := s.messageJobs
 q.mu.Lock()
 now := time.Now()
 for key, job := range q.jobs {
  if !job.FinishedAt.IsZero() && now.Sub(job.FinishedAt) > messageJobRetention { delete(q.jobs, key) }
 }
 if existing := q.active[id]; existing != "" {
  job := *q.jobs[existing]
  if job.RequestDigest != sha256.Sum256([]byte(input.Content)) {
   q.mu.Unlock()
   writeError(w, http.StatusConflict, errors.New("another message is already running for this session"))
   return
  }
  q.mu.Unlock()
  writeJSON(w, http.StatusAccepted, job)
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
 job := &messageJob{ID: hex.EncodeToString(token), SessionID: id, Status: "running", RequestDigest: sha256.Sum256([]byte(input.Content)), CreatedAt: now.UTC()}
 q.jobs[job.ID] = job
 q.active[id] = job.ID
 q.mu.Unlock()
 go func() {
  defer func() { <-q.slots }()
  ctx, cancel := context.WithTimeout(context.Background(), messageJobTimeout)
  defer cancel()
  session, err := s.Engine.HandleMessage(ctx, id, input.Content)
  q.mu.Lock()
  defer q.mu.Unlock()
  if err != nil { job.Status = "failed"; job.Error = "local inference failed; inspect server diagnostics" } else { job.Status = "completed"; job.Result = &session }
  job.FinishedAt = time.Now().UTC()
  delete(q.active, id)
 }()
 writeJSON(w, http.StatusAccepted, job)
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
 writeJSON(w, http.StatusOK, snapshot)
}
