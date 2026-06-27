package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type WorkflowStep struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

type Workflow struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Steps       []WorkflowStep `json:"steps"`
}

func Defaults() []Workflow {
	return []Workflow{
		{
			Name:        "verify-federation",
			Description: "Generate manifests, registry, providers, reports, and dependency graph.",
			Steps: []WorkflowStep{
				{Name: "manifest", Command: "manifest"},
				{Name: "registry", Command: "registry"},
				{Name: "providers", Command: "providers"},
				{Name: "dashboard", Command: "dashboard"},
				{Name: "deps", Command: "deps"},
				{Name: "verify", Command: "verify"},
			},
		},
		{
			Name:        "safe-sync",
			Description: "Run safe federation sync without auto-committing dirty repositories.",
			Steps: []WorkflowStep{
				{Name: "sync-safe", Command: "sync", Args: []string{"--safe"}},
			},
		},
	}
}

func WriteRegistry(cfg config.Config) error {
	out := filepath.Join(cfg.OSHome, "registry", "workflows.json")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(Defaults(), "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(out, append(data, '\n'), 0644); err != nil {
		return err
	}

	fmt.Println("Wrote", out)
	return nil
}

func List(cfg config.Config) error {
	if err := WriteRegistry(cfg); err != nil {
		return err
	}

	fmt.Printf("%-24s %s\n", "WORKFLOW", "DESCRIPTION")
	for _, wf := range Defaults() {
		fmt.Printf("%-24s %s\n", wf.Name, wf.Description)
	}

	return nil
}
