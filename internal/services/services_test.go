package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/state"
)

func TestDefaultsExposeOnlyTruthfulStatuses(t *testing.T) {
	services := Defaults()
	if len(services) == 0 {
		t.Fatal("no default services")
	}
	for _, svc := range services {
		if svc.Name == "" || svc.Status == "" || svc.Description == "" {
			t.Fatalf("incomplete service: %#v", svc)
		}
	}
}

func TestMarkPersistsServiceStateAndEmitsEvent(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}

	if err := Mark(cfg, "api", "running"); err != nil {
		t.Fatal(err)
	}

	st, err := state.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if st.Services["api"] != "running" {
		t.Fatalf("api status = %q, want running", st.Services["api"])
	}

	eventsPath := filepath.Join(dir, "var", "events", "events.jsonl")
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "service.running") {
		t.Fatalf("event log does not contain service.running: %s", data)
	}
}
