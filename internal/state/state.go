package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type RuntimeState struct {
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	StartedAt string            `json:"startedAt"`
	UpdatedAt string            `json:"updatedAt"`
	Services  map[string]string `json:"services"`
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
	}
}

func Save(cfg config.Config, s RuntimeState) error {
	s.UpdatedAt = time.Now().Format(time.RFC3339)
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
	err = json.Unmarshal(b, &s)
	return s, err
}
