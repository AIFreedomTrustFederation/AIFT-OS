package uxihttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/uxi"
)

type fakeCompleter struct{}

func (fakeCompleter) Complete(context.Context, string, []uxi.ChatMessage) (string, error) {
	return "local answer", nil
}

func newTestServer(t *testing.T) (*Server, *uxi.Store) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "AIFT-OS", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	store, err := uxi.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	engine, err := uxi.NewEngine(store, root, fakeCompleter{})
	if err != nil {
		t.Fatal(err)
	}
	server, err := New(engine)
	if err != nil {
		t.Fatal(err)
	}
	return server, store
}

func TestHealthAndConversation(t *testing.T) {
	server, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("health code = %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/sessions", bytes.NewBufferString(`{"title":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create code = %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestGovernanceEndpointsRecordButDoNotExecute(t *testing.T) {
	server, store := newTestServer(t)
	session, err := store.CreateSession("governance")
	if err != nil {
		t.Fatal(err)
	}

	plan := `{"objective":"Review Forge","steps":[{"title":"Inspect mission"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+session.ID+"/plans", bytes.NewBufferString(plan))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("plan code=%d body=%s", rr.Code, rr.Body.String())
	}

	action := `{"kind":"forge.patch.apply","target":"AIFT-Forge","risk":"moderate","approval_required":true}`
	req = httptest.NewRequest(http.MethodPost, "/v1/sessions/"+session.ID+"/actions", bytes.NewBufferString(action))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("action code=%d body=%s", rr.Code, rr.Body.String())
	}
	var updated uxi.Session
	if err := json.Unmarshal(rr.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}

	decision := `{"decision":"approved","actor":"human-local-operator"}`
	req = httptest.NewRequest(http.MethodPost, "/v1/sessions/"+session.ID+"/actions/"+updated.Actions[0].ID+"/decision", bytes.NewBufferString(decision))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("decision code=%d body=%s", rr.Code, rr.Body.String())
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Actions[0].Status != uxi.StatusApproved || len(updated.Jobs) != 0 {
		t.Fatalf("updated=%#v", updated)
	}
}

func TestReadOnlyInvocationEndpoint(t *testing.T) {
	server, store := newTestServer(t)
	session, err := store.CreateSession("invoke")
	if err != nil {
		t.Fatal(err)
	}
	session, err = store.ProposeAction(session.ID, uxi.Action{Kind: "repository.inspect", Target: "AIFT-OS", Risk: "low"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+session.ID+"/actions/"+session.Actions[0].ID+"/invoke", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("invoke code=%d body=%s", rr.Code, rr.Body.String())
	}
	var updated uxi.Session
	if err := json.Unmarshal(rr.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Jobs[0].Status != uxi.StatusSucceeded || updated.Jobs[0].ResultArtifactID == "" {
		t.Fatalf("updated=%#v", updated)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/artifacts/"+updated.Jobs[0].ResultArtifactID, nil)
	rr = httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !json.Valid(rr.Body.Bytes()) {
		t.Fatalf("artifact code=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAdaptersEndpointListsOnlyReadOnlyDefaults(t *testing.T) {
	server, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/adapters", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d", rr.Code)
	}
	var response struct {
		Adapters []uxi.AdapterDescriptor `json:"adapters"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Adapters) != 2 {
		t.Fatalf("adapters=%#v", response.Adapters)
	}
	for _, adapter := range response.Adapters {
		if adapter.Mutating {
			t.Fatalf("adapter=%#v", adapter)
		}
	}
}
