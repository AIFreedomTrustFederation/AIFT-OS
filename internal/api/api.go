package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/jobs"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/providers"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/services"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/state"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/sync"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/version"
)

type Server struct {
	Config config.Config
	Addr   string
	// Token is used for Bearer authentication on mutating endpoints.
	// If empty, a random token is generated and printed at startup.
	Token string
}

func New(cfg config.Config, addr string) Server {
	token := os.Getenv("AIFT_API_TOKEN")
	return Server{Config: cfg, Addr: addr, Token: token}
}

func (s Server) Serve() error {
	if s.Token == "" {
		s.Token = generateToken()
		fmt.Println("Generated API token (set AIFT_API_TOKEN to use a fixed value):")
		fmt.Println(" ", s.Token)
	}

	mux := http.NewServeMux()

	// Public read-only endpoint.
	mux.HandleFunc("/health", s.securityHeaders(s.json(map[string]string{
		"status":  "ok",
		"name":    version.Name,
		"version": version.Version,
	})))

	// Authenticated read-only endpoints.
	mux.HandleFunc("/state", s.securityHeaders(s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		st, err := state.Load(s.Config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "api: failed to load state: %v\n", err)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(st); err != nil {
			fmt.Fprintf(os.Stderr, "api: failed to encode state response: %v\n", err)
		}
	})))

	mux.HandleFunc("/services", s.securityHeaders(s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(services.Defaults()); err != nil {
			fmt.Fprintf(os.Stderr, "api: failed to encode services response: %v\n", err)
		}
	})))

	// Authenticated mutating endpoints.
	mux.HandleFunc("/actions/verify", s.securityHeaders(s.requireAuth(s.action(func() error {
		if err := manifests.EnsureAll(s.Config); err != nil {
			return err
		}
		if err := providers.WriteRegistry(s.Config); err != nil {
			return err
		}
		if err := registry.Generate(s.Config); err != nil {
			return err
		}
		if err := reports.Dashboard(s.Config); err != nil {
			return err
		}
		return reports.Deps(s.Config)
	}))))

	mux.HandleFunc("/actions/tick", s.securityHeaders(s.requireAuth(s.action(func() error {
		return jobs.RunAll(s.Config)
	}))))

	mux.HandleFunc("/actions/sync-safe", s.securityHeaders(s.requireAuth(s.action(func() error {
		return sync.Safe(s.Config)
	}))))

	// Authenticated file-serving endpoints.
	mux.HandleFunc("/registry/repos", s.securityHeaders(s.requireAuth(file(filepath.Join(s.Config.OSHome, "registry", "repos.json")))))
	mux.HandleFunc("/registry/providers", s.securityHeaders(s.requireAuth(file(filepath.Join(s.Config.OSHome, "registry", "providers.json")))))
	mux.HandleFunc("/reports/dashboard", s.securityHeaders(s.requireAuth(file(filepath.Join(s.Config.OSHome, "reports", "dashboard.md")))))
	mux.HandleFunc("/events", s.securityHeaders(s.requireAuth(file(filepath.Join(s.Config.OSHome, "var", "events", "events.jsonl")))))

	// Default to localhost-only binding when no explicit host is specified.
	addr := s.Addr
	if !strings.Contains(addr, ":") || strings.HasPrefix(addr, ":") {
		if strings.HasPrefix(addr, ":") {
			addr = "127.0.0.1" + addr
		} else {
			addr = "127.0.0.1:" + addr
		}
	}

	fmt.Println("AIFT-OS API listening on", addr)
	return http.ListenAndServe(addr, mux)
}

func (s Server) json(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(v); err != nil {
			fmt.Fprintf(os.Stderr, "api: failed to encode json response: %v\n", err)
		}
	}
}

func (s Server) action(fn func() error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte("method not allowed\n"))
			return
		}
		if err := fn(); err != nil {
			log.Printf("action error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			if encErr := json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "internal error"}); encErr != nil {
				fmt.Fprintf(os.Stderr, "api: failed to encode error response: %v\n", encErr)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			fmt.Fprintf(os.Stderr, "api: failed to encode ok response: %v\n", err)
		}
	}
}

// requireAuth wraps a handler with Bearer token authentication.
func (s Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.Header().Set("WWW-Authenticate", `Bearer realm="aift-os"`)
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if subtle.ConstantTimeCompare([]byte(token), []byte(s.Token)) != 1 {
			http.Error(w, "invalid token", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// securityHeaders adds common protective HTTP headers.
func (s Server) securityHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "no-store")
		next(w, r)
	}
}

func file(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}

func generateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}
