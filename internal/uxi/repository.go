package uxi

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func ResolveAIFTRoot() (string, error) {
	if root := strings.TrimSpace(os.Getenv("AIFT_ROOT")); root != "" {
		return filepath.Abs(root)
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidate := filepath.Join(home, "AIFT")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate, nil
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if filepath.Base(cwd) == "AIFT-OS" {
		return filepath.Dir(cwd), nil
	}
	return cwd, nil
}

func DiscoverRepositories(root string) ([]Repository, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read AIFT root %s: %w", root, err)
	}
	var repos []Repository
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		path := filepath.Join(root, entry.Name())
		if !isDir(filepath.Join(path, ".git")) {
			continue
		}
		repo, err := inspectRepository(path)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	sort.Slice(repos, func(i, j int) bool { return repos[i].Name < repos[j].Name })
	return repos, nil
}

func inspectRepository(path string) (Repository, error) {
	name := filepath.Base(path)
	now := time.Now().UTC()
	repo := Repository{
		ID:     strings.ToLower(strings.ReplaceAll(name, "_", "-")),
		Name:   name,
		Path:   path,
		Role:   classifyRepositoryRole(name),
		Status: "detected",
		Git:    true,
		Evidence: []Evidence{{
			ID:         newID("evd"),
			Kind:       "filesystem",
			Source:     filepath.Join(path, ".git"),
			Summary:    "Git repository detected",
			Status:     "observed",
			ObservedAt: now,
		}},
	}

	for file, language := range map[string]string{
		"go.mod":         "Go",
		"package.json":   "JavaScript/TypeScript",
		"pyproject.toml": "Python",
		"Cargo.toml":     "Rust",
	} {
		if isFile(filepath.Join(path, file)) {
			repo.Languages = append(repo.Languages, language)
			repo.Evidence = append(repo.Evidence, Evidence{
				ID: newID("evd"), Kind: "manifest", Source: filepath.Join(path, file),
				Summary: language + " project manifest detected", Status: "observed", ObservedAt: now,
			})
		}
	}
	sort.Strings(repo.Languages)

	capabilities, err := readCapabilities(filepath.Join(path, ".aift", "capabilities.json"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		repo.Status = "blocked"
		repo.Evidence = append(repo.Evidence, Evidence{
			ID: newID("evd"), Kind: "error", Source: filepath.Join(path, ".aift", "capabilities.json"),
			Summary: "Capability manifest could not be decoded", Detail: err.Error(), Status: "failed", ObservedAt: now,
		})
		return repo, nil
	}
	if len(capabilities) > 0 {
		repo.Capabilities = capabilities
		repo.Status = "ready"
		repo.Evidence = append(repo.Evidence, Evidence{
			ID: newID("evd"), Kind: "capability_manifest", Source: filepath.Join(path, ".aift", "capabilities.json"),
			Summary: fmt.Sprintf("%d declared capabilities discovered", len(capabilities)), Status: "observed", ObservedAt: now,
		})
	}
	return repo, nil
}

func readCapabilities(path string) ([]Capability, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Capabilities []Capability `json:"capabilities"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	return envelope.Capabilities, nil
}

func classifyRepositoryRole(name string) string {
	roles := map[string]string{
		"AIFT-OS":                     "federation-kernel",
		"AIFT-Runtime":                "runtime-prototype",
		"AIFT-Forge":                  "software-mission-engine",
		"AIFT-Genesis":                "federation-genome",
		"AI-Freedom-Trust":            "doctrine-and-research",
		"VPS":                         "infrastructure-and-nodes",
		"mobox":                       "compatibility-runtime",
		"booksmith-ai":                "knowledge-application",
		"BookSmith-Federation-OS":     "knowledge-product-specification",
		"OpenMontage":                 "video-application",
		"Aether_Coin_biozonecurrency": "stewardship-application",
		"TheMindofAll":                "model-registry",
	}
	if role, ok := roles[name]; ok {
		return role
	}
	return "federated-application"
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
