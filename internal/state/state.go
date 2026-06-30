package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type ServiceState struct {
	Name        string `json:"name"`
	Repo        string `json:"repo"`
	Status      string `json:"status"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updatedAt"`
}

type RuntimeState struct {
	Name      string                  `json:"name"`
	Status    string                  `json:"status"`
	StartedAt string                  `json:"startedAt"`
	UpdatedAt string                  `json:"updatedAt"`
	Services  map[string]string       `json:"services"`
	Catalog   map[string]ServiceState `json:"catalog"`
}

func Path(cfg config.Config) string {
	return filepath.Join(cfg.OSHome, "var", "runtime-state.json")
}

func New() RuntimeState {
	now := time.Now().Format(time.RFC3339)
	return RuntimeState{
		Name:      "AIFT-OS",
		Status:    "running",
		StartedAt: now,
		UpdatedAt: now,
		Services:  map[string]string{},
		Catalog:   map[string]ServiceState{},
	}
}

func Save(cfg config.Config, s RuntimeState) error {
	now := time.Now().Format(time.RFC3339)
	s.UpdatedAt = now
	if s.StartedAt == "" {
		s.StartedAt = now
	}
	if s.Services == nil {
		s.Services = map[string]string{}
	}
	if s.Catalog == nil {
		s.Catalog = map[string]ServiceState{}
	}
	if err := os.MkdirAll(filepath.Dir(Path(cfg)), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(cfg), append(b, '\n'), 0644)
}

func Load(cfg config.Config) (RuntimeState, error) {
	b, err := os.ReadFile(Path(cfg))
	if err != nil {
		return New(), err
	}
	var s RuntimeState
	if err := json.Unmarshal(b, &s); err != nil {
		return New(), err
	}
	if s.Services == nil {
		s.Services = map[string]string{}
	}
	if s.Catalog == nil {
		s.Catalog = map[string]ServiceState{}
	}
	return s, nil
}
