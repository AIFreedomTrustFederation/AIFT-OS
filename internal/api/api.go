package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
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

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"name":    version.Name,
			"version": version.Version,
		})
	})

	mux.HandleFunc("/registry/repos", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(s.Config.OSHome, "registry", "repos.json"))
	})

	mux.HandleFunc("/registry/providers", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(s.Config.OSHome, "registry", "providers.json"))
	})

	mux.HandleFunc("/reports/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(s.Config.OSHome, "reports", "dashboard.md"))
	})

	fmt.Println("AIFT-OS API listening on", s.Addr)
	return http.ListenAndServe(s.Addr, mux)
}
