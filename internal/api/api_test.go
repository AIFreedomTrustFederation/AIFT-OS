package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestHealthEndpoint(t *testing.T) {
	srv := testServer(t)
	mux := buildMux(srv)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("health status = %d, want 200", w.Code)
	}

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("status = %q, want ok", body["status"])
	}
}

func TestHealthSecurityHeaders(t *testing.T) {
	srv := testServer(t)
	mux := buildMux(srv)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options: nosniff")
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("missing X-Frame-Options: DENY")
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Error("missing Cache-Control: no-store")
	}
}

func TestRequireAuthMissingToken(t *testing.T) {
	srv := testServer(t)
	mux := buildMux(srv)

	req := httptest.NewRequest("GET", "/state", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	if w.Header().Get("WWW-Authenticate") == "" {
		t.Error("should include WWW-Authenticate header")
	}
}

func TestRequireAuthWrongToken(t *testing.T) {
	srv := testServer(t)
	mux := buildMux(srv)

	req := httptest.NewRequest("GET", "/state", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestRequireAuthCorrectToken(t *testing.T) {
	srv := testServer(t)
	mux := buildMux(srv)

	req := httptest.NewRequest("GET", "/state", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Should not be 401 or 403 (may be 200 or 500 depending on state)
	if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
		t.Errorf("status = %d, should pass auth", w.Code)
	}
}

func TestActionRequiresPost(t *testing.T) {
	srv := testServer(t)
	mux := buildMux(srv)

	req := httptest.NewRequest("GET", "/actions/tick", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestGenerateTokenLength(t *testing.T) {
	token := generateToken()
	if len(token) != 64 {
		t.Errorf("token length = %d, want 64 hex chars", len(token))
	}
}

func TestGenerateTokenUnique(t *testing.T) {
	t1 := generateToken()
	t2 := generateToken()
	if t1 == t2 {
		t.Error("two generated tokens should not be equal")
	}
}

func TestNewWithEnvToken(t *testing.T) {
	t.Setenv("AIFT_API_TOKEN", "env-token-value")
	dir := t.TempDir()
	cfg := config.Load()

	srv := New(cfg, ":8787")
	if srv.Token != "env-token-value" {
		t.Errorf("Token = %q, want env-token-value", srv.Token)
	}
	if srv.Addr != ":8787" {
		t.Errorf("Addr = %q", srv.Addr)
	}
}

func TestNewWithoutEnvToken(t *testing.T) {
	t.Setenv("AIFT_API_TOKEN", "")
	dir := t.TempDir()
	cfg := config.Load()

	srv := New(cfg, ":9999")
	if srv.Token != "" {
		t.Errorf("Token = %q, want empty (will be generated on Serve)", srv.Token)
	}
}

// buildMux creates an http.ServeMux matching Server.Serve() for test use.
func buildMux(s Server) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.securityHeaders(s.json(map[string]string{"status": "ok"})))
	mux.HandleFunc("/state", s.securityHeaders(s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})))
	mux.HandleFunc("/actions/tick", s.securityHeaders(s.requireAuth(s.action(func() error {
		return nil
	}))))
	return mux
}

func testServer(t *testing.T) Server {
	t.Helper()
	dir := t.TempDir()
	return Server{
		Config: config.Load(),
		Addr:   ":0",
		Token:  "test-token-123",
	}
}
