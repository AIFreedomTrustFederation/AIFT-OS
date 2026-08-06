package uxihttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/uxi"
)

type Server struct {
	Engine *uxi.Engine
	mux    *http.ServeMux
}

func New(engine *uxi.Engine) (*Server, error) {
	if engine == nil {
		return nil, errors.New("engine is required")
	}
	s := &Server{Engine: engine, mux: http.NewServeMux()}
	s.routes()
	return s, nil
}

func (s *Server) Handler() http.Handler {
	return securityHeaders(s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /v1/system", s.handleSystem)
	s.mux.HandleFunc("GET /v1/repositories", s.handleRepositories)
	s.mux.HandleFunc("GET /v1/adapters", s.handleAdapters)
	s.mux.HandleFunc("GET /v1/artifacts/{id}", s.handleArtifact)
	s.mux.HandleFunc("GET /v1/adapters/forge/mission", s.handleForgeMission)
	s.mux.HandleFunc("GET /v1/sessions", s.handleListSessions)
	s.mux.HandleFunc("POST /v1/sessions", s.handleCreateSession)
	s.mux.HandleFunc("GET /v1/sessions/{id}", s.handleGetSession)
	s.mux.HandleFunc("POST /v1/sessions/{id}/messages", s.handleMessage)
	s.mux.HandleFunc("POST /v1/sessions/{id}/plans", s.handlePlan)
	s.mux.HandleFunc("POST /v1/sessions/{id}/actions", s.handleAction)
	s.mux.HandleFunc("POST /v1/sessions/{id}/actions/{actionID}/decision", s.handleActionDecision)
	s.mux.HandleFunc("POST /v1/sessions/{id}/actions/{actionID}/invoke", s.handleActionInvoke)
	s.mux.HandleFunc("GET /v1/events", s.handleEvents)
	s.mux.HandleFunc("GET /", s.handleIndex)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "pass",
		"service": "aiftd",
		"time":    time.Now().UTC(),
	})
}

func (s *Server) handleSystem(w http.ResponseWriter, r *http.Request) {
	repos, err := s.Engine.Repositories()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":           "active",
		"mode":             "governed-read-only-execution",
		"aift_root":        s.Engine.AIFTRoot,
		"data_root":        s.Engine.Store.Root(),
		"repository_count": len(repos),
		"truth_contract":   "observed evidence outranks generated claims",
	})
}

func (s *Server) handleRepositories(w http.ResponseWriter, r *http.Request) {
	repos, err := s.Engine.Repositories()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"repositories": repos})
}

func (s *Server) handleAdapters(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"adapters": s.Engine.Adapters.List()})
}

func (s *Server) handleArtifact(w http.ResponseWriter, r *http.Request) {
	artifact, err := s.Engine.Store.ReadArtifact(r.PathValue("id"))
	if errors.Is(err, uxi.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(artifact)
}

func (s *Server) handleForgeMission(w http.ResponseWriter, r *http.Request) {
	mission, evidence, err := uxi.InspectForgeMission(s.Engine.AIFTRoot)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"mission": mission, "evidence": evidence})
}

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.Engine.Store.ListSessions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	session, err := s.Engine.Store.CreateSession(input.Title)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	session, err := s.Engine.Store.GetSession(r.PathValue("id"))
	if errors.Is(err, uxi.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *Server) handleMessage(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content string `json:"content"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	session, err := s.Engine.HandleMessage(r.Context(), r.PathValue("id"), input.Content)
	if errors.Is(err, uxi.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *Server) handlePlan(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Objective string         `json:"objective"`
		Steps     []uxi.PlanStep `json:"steps"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	session, err := s.Engine.Store.AddPlan(r.PathValue("id"), uxi.Plan{Objective: input.Objective, Steps: input.Steps})
	if errors.Is(err, uxi.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (s *Server) handleAction(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Kind             string         `json:"kind"`
		Target           string         `json:"target"`
		Risk             string         `json:"risk"`
		ApprovalRequired bool           `json:"approval_required"`
		Parameters       map[string]any `json:"parameters"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	session, err := s.Engine.Store.ProposeAction(r.PathValue("id"), uxi.Action{
		Kind: input.Kind, Target: input.Target, Risk: input.Risk,
		ApprovalRequired: input.ApprovalRequired, Parameters: input.Parameters,
	})
	if errors.Is(err, uxi.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (s *Server) handleActionDecision(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Decision string `json:"decision"`
		Actor    string `json:"actor"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	session, err := s.Engine.Store.DecideAction(r.PathValue("id"), r.PathValue("actionID"), input.Decision, input.Actor)
	if errors.Is(err, uxi.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *Server) handleActionInvoke(w http.ResponseWriter, r *http.Request) {
	session, err := s.Engine.InvokeAction(r.Context(), r.PathValue("id"), r.PathValue("actionID"))
	if errors.Is(err, uxi.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if errors.Is(err, uxi.ErrMutationDisabled) {
		writeError(w, http.StatusForbidden, err)
		return
	}
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]any{"status": "fail", "error": err.Error(), "session": session})
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	events, err := s.Engine.Store.ListEvents(200)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func decodeJSON(r *http.Request, target any) error {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return errors.New("Content-Type must be application/json")
	}
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("invalid JSON: request must contain exactly one object")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, "response encoding failed", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"status": "fail", "error": err.Error()})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; connect-src 'self'")
		next.ServeHTTP(w, r)
	})
}
