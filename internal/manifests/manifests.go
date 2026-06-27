package manifests

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Manifest struct {
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	Sovereign    bool     `json:"sovereign"`
	ManagedBy    string   `json:"managedBy"`
	Dependencies []string `json:"dependencies"`
	Capabilities []string `json:"capabilities"`
	CommandsPath string   `json:"commandsPath"`
}

func Path(repo string) string {
	return filepath.Join(repo, ".aift", "repo.json")
}

func EnsureAll(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	for _, repo := range repos {
		if err := Ensure(repo); err != nil {
			return err
		}
	}

	return nil
}

func Ensure(repo workspace.Repo) error {
	dir := filepath.Join(repo.Path, ".aift", "commands")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := Path(repo.Path)
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	role := "sovereign-repository"
	if repo.Name == "AIFT-OS" {
		role = "federation-control-plane"
	}

	m := Manifest{
		Name:         repo.Name,
		Role:         role,
		Sovereign:    true,
		ManagedBy:    "AIFT-OS",
		Dependencies: []string{},
		Capabilities: []string{},
		CommandsPath: ".aift/commands",
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0644)
}

func Valid(repoPath string) bool {
	data, err := os.ReadFile(Path(repoPath))
	if err != nil {
		return false
	}

	var m Manifest
	if json.Unmarshal(data, &m) != nil {
		return false
	}

	return m.Name != "" && m.Role != ""
}
