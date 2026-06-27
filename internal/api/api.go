package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

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
}

func New(cfg config.Config, addr string) Server {
	return Server{Config: cfg, Addr: addr}
}

func (s Server) Serve() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.json(map[string]string{
		"status":  "ok",
		"name":    version.Name,
		"version": version.Version,
	}))

	mux.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		st, _ := state.Load(s.Config)
		_ = json.NewEncoder(w).Encode(st)
	})

	mux.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(services.Defaults())
	})

	mux.HandleFunc("/actions/verify", s.action(func() error {
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
	}))

	mux.HandleFunc("/actions/tick", s.action(func() error {
		return jobs.RunAll(s.Config)
	}))

	mux.HandleFunc("/actions/sync-safe", s.action(func() error {
		return sync.Safe(s.Config)
	}))

	mux.HandleFunc("/registry/repos", file(filepath.Join(s.Config.OSHome, "registry", "repos.json")))
	mux.HandleFunc("/registry/providers", file(filepath.Join(s.Config.OSHome, "registry", "providers.json")))
	mux.HandleFunc("/reports/dashboard", file(filepath.Join(s.Config.OSHome, "reports", "dashboard.md")))
	mux.HandleFunc("/events", file(filepath.Join(s.Config.OSHome, "var", "events", "events.jsonl")))

	fmt.Println("AIFT-OS API listening on", s.Addr)
	return http.ListenAndServe(s.Addr, mux)
}

func (s Server) json(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
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
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func file(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}
