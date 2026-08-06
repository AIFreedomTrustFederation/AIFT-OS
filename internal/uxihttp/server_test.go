package uxihttp

import (
	"bytes"
	"context"
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

func TestHealthAndConversation(t *testing.T) {
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
