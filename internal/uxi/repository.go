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

// ResolveAIFTRoot finds the local federation workspace without assuming repository names.
func ResolveAIFTRoot() (string, error) {
	if root := strings.TrimSpace(os.Getenv("AIFT_ROOT")); root != "" {
		return filepath.Abs(root)
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidate := filepath.Join(home, "AIFT")
		if isDir(candidate) {
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

// DiscoverRepositories inspects direct child directories and symlinked directories for Git evidence.
func DiscoverRepositories(root string) ([]Repository, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read AIFT root %s: %w", root, err)
	}
	var repos []Repository
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		path := filepath.Join(root, entry.Name())
		if !isDir(path) || !exists(filepath.Join(path, ".git")) {
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
		ID: strings.ToLower(strings.ReplaceAll(name, "_", "-")), Name: name, Path: path,
		Role: classifyRepositoryRole(name), Status: "detected", Git: true,
		Evidence: []Evidence{newEvidence("filesystem", filepath.Join(path, ".git"), "Git repository detected", "", "observed", now)},
	}

	manifests := []struct{ file, language string }{
		{"Cargo.toml", "Rust"},
		{"go.mod", "Go"},
		{"package.json", "JavaScript/TypeScript"},
		{"pyproject.toml", "Python"},
	}
	for _, manifest := range manifests {
		manifestPath := filepath.Join(path, manifest.file)
		if isFile(manifestPath) {
			repo.Languages = append(repo.Languages, manifest.language)
			repo.Evidence = append(repo.Evidence, newEvidence("manifest", manifestPath, manifest.language+" project manifest detected", "", "observed", now))
		}
	}
	sort.Strings(repo.Languages)

	capabilityPath := filepath.Join(path, ".aift", "capabilities.json")
	capabilities, err := readCapabilities(capabilityPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		repo.Status = "blocked"
		repo.Evidence = append(repo.Evidence, newEvidence("error", capabilityPath, "Capability manifest could not be decoded", err.Error(), "failed", now))
		return repo, nil
	}
	if len(capabilities) > 0 {
		repo.Capabilities = capabilities
		repo.Status = aggregateCapabilityStatus(capabilities)
		repo.Evidence = append(repo.Evidence, newEvidence("capability_manifest", capabilityPath, fmt.Sprintf("%d declared capabilities discovered; aggregate status %s", len(capabilities), repo.Status), "", "observed", now))
	}
	return repo, nil
}

func newEvidence(kind, source, summary, detail, status string, observedAt time.Time) Evidence {
	return Evidence{ID: stableID("evd", kind+":"+source+":"+summary+":"+detail+":"+status), Kind: kind, Source: source, Summary: summary, Detail: detail, Status: status, ObservedAt: observedAt}
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

func aggregateCapabilityStatus(capabilities []Capability) string {
	ready := false
	for _, capability := range capabilities {
		switch strings.ToLower(strings.TrimSpace(capability.Status)) {
		case "blocked", "broken", "fail", "failed":
			return "blocked"
		case "ready", "active", "v1":
			ready = true
		}
	}
	if ready {
		return "ready"
	}
	return "detected"
}

func classifyRepositoryRole(name string) string {
	roles := map[string]string{
		"AIFT-OS": "federation-kernel", "AIFT-Runtime": "runtime-prototype", "AIFT-Forge": "software-mission-engine",
		"AIFT-Genesis": "federation-genome", "AI-Freedom-Trust": "doctrine-and-research", "VPS": "infrastructure-and-nodes",
		"mobox": "compatibility-runtime", "booksmith-ai": "knowledge-application", "BookSmith-Federation-OS": "knowledge-product-specification",
		"OpenMontage": "video-application", "Aether_Coin_biozonecurrency": "stewardship-application", "TheMindofAll": "model-registry",
	}
	if role, ok := roles[name]; ok {
		return role
	}
	return "federated-application"
}

func exists(path string) bool { _, err := os.Stat(path); return err == nil }
func isDir(path string) bool  { info, err := os.Stat(path); return err == nil && info.IsDir() }
func isFile(path string) bool { info, err := os.Stat(path); return err == nil && !info.IsDir() }
