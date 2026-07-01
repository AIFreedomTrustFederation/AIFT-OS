package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestDefaultsKeepPlannedWorkflowsTruthful(t *testing.T) {
	workflows := Defaults()
	if len(workflows) != 2 {
		t.Fatalf("workflow count = %d, want 2", len(workflows))
	}
	if workflows[0].Name != "verify-federation" {
		t.Fatalf("first workflow = %q", workflows[0].Name)
	}
	if workflows[1].Steps[0].Command != "sync" || workflows[1].Steps[0].Args[0] != "--safe" {
		t.Fatalf("safe-sync step = %#v", workflows[1].Steps[0])
	}
}

func TestWriteRegistryWritesWorkflowJSON(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}
	if err := WriteRegistry(cfg); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "registry", "workflows.json"))
	if err != nil {
		t.Fatal(err)
	}
	var workflows []Workflow
	if err := json.Unmarshal(data, &workflows); err != nil {
		t.Fatal(err)
	}
	if len(workflows) != len(Defaults()) {
		t.Fatalf("written workflows = %d, want %d", len(workflows), len(Defaults()))
	}
}
