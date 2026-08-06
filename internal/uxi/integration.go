package uxi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const integrationSchemaV1 = "aift.uxi.integration.v1"
const maxIntegrationManifestBytes = 256 << 10
const maxIntegrationSourceBytes = 4 << 20

type IntegrationManifest struct {
	Schema     string              `json:"schema"`
	Repository string              `json:"repository"`
	Role       string              `json:"role,omitempty"`
	Sources    []IntegrationSource `json:"sources"`
	SourcePath string              `json:"source_path,omitempty"`
}

type IntegrationSource struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	Repository  string `json:"repository,omitempty"`
	RepoPath    string `json:"repo_path,omitempty"`
	Manifest    string `json:"manifest,omitempty"`
}

func DiscoverIntegrationSources(aiftRoot string) ([]IntegrationSource, []Evidence, error) {
	entries, err := os.ReadDir(aiftRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("read AIFT root: %w", err)
	}
	var sources []IntegrationSource
	var evidence []Evidence
	ids := map[string]string{}
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		repoPath := filepath.Join(aiftRoot, entry.Name())
		if !exists(filepath.Join(repoPath, ".git")) {
			continue
		}
		manifestPath := filepath.Join(repoPath, ".aift", "uxi.json")
		manifest, readErr := readIntegrationManifest(manifestPath)
		if errors.Is(readErr, os.ErrNotExist) {
			continue
		}
		if readErr != nil {
			evidence = append(evidence, Evidence{ID: newID("evd"), Kind: "integration_manifest", Source: manifestPath, Summary: "UXI integration manifest is invalid", Detail: readErr.Error(), Status: "failed", ObservedAt: time.Now().UTC()})
			continue
		}
		if manifest.Repository != entry.Name() {
			evidence = append(evidence, Evidence{ID: newID("evd"), Kind: "integration_manifest", Source: manifestPath, Summary: "UXI integration manifest repository does not match directory", Detail: fmt.Sprintf("declared=%s directory=%s", manifest.Repository, entry.Name()), Status: "failed", ObservedAt: time.Now().UTC()})
			continue
		}
		for _, source := range manifest.Sources {
			if previous, duplicate := ids[source.ID]; duplicate {
				return nil, evidence, fmt.Errorf("duplicate integration source %q in %s and %s", source.ID, previous, manifestPath)
			}
			ids[source.ID] = manifestPath
			source.Repository = manifest.Repository
			source.RepoPath = repoPath
			source.Manifest = manifestPath
			sources = append(sources, source)
		}
		evidence = append(evidence, Evidence{ID: newID("evd"), Kind: "integration_manifest", Source: manifestPath, Summary: fmt.Sprintf("%d UXI integration sources discovered", len(manifest.Sources)), Status: "observed", ObservedAt: time.Now().UTC()})
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].ID < sources[j].ID })
	return sources, evidence, nil
}

func readIntegrationManifest(path string) (IntegrationManifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return IntegrationManifest{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxIntegrationManifestBytes+1))
	if err != nil {
		return IntegrationManifest{}, err
	}
	if len(data) > maxIntegrationManifestBytes {
		return IntegrationManifest{}, errors.New("integration manifest exceeds size limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest IntegrationManifest
	if err := decoder.Decode(&manifest); err != nil {
		return IntegrationManifest{}, err
	}
	if manifest.Schema != integrationSchemaV1 {
		return IntegrationManifest{}, fmt.Errorf("unsupported integration schema %q", manifest.Schema)
	}
	if strings.TrimSpace(manifest.Repository) == "" {
		return IntegrationManifest{}, errors.New("integration repository is required")
	}
	seen := map[string]bool{}
	for i := range manifest.Sources {
		source := &manifest.Sources[i]
		source.ID = strings.TrimSpace(source.ID)
		source.Kind = strings.TrimSpace(source.Kind)
		source.Path = filepath.ToSlash(strings.TrimSpace(source.Path))
		source.Status = strings.TrimSpace(source.Status)
		if source.ID == "" || source.Kind == "" || source.Path == "" || source.Status == "" {
			return IntegrationManifest{}, fmt.Errorf("integration source %d requires id, kind, path, and status", i+1)
		}
		if seen[source.ID] {
			return IntegrationManifest{}, fmt.Errorf("duplicate source id %q", source.ID)
		}
		seen[source.ID] = true
		if source.Kind != "json" {
			return IntegrationManifest{}, fmt.Errorf("unsupported source kind %q", source.Kind)
		}
		if source.Status != "ready" && source.Status != "detected" && source.Status != "planned" {
			return IntegrationManifest{}, fmt.Errorf("unsupported source status %q", source.Status)
		}
		if err := validateRelativePath(source.Path); err != nil {
			return IntegrationManifest{}, fmt.Errorf("source %s path: %w", source.ID, err)
		}
	}
	manifest.SourcePath = path
	return manifest, nil
}

func validateRelativePath(path string) error {
	if filepath.IsAbs(path) {
		return errors.New("absolute paths are not allowed")
	}
	clean := filepath.Clean(filepath.FromSlash(path))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return errors.New("path must remain inside the repository")
	}
	return nil
}

func readIntegrationJSON(source IntegrationSource) (map[string]any, []Evidence, error) {
	if source.Status != "ready" {
		return nil, nil, fmt.Errorf("integration source %s is %s, not ready", source.ID, source.Status)
	}
	if err := validateRelativePath(source.Path); err != nil {
		return nil, nil, err
	}
	repoReal, err := filepath.EvalSymlinks(source.RepoPath)
	if err != nil {
		return nil, nil, err
	}
	candidate := filepath.Join(source.RepoPath, filepath.FromSlash(source.Path))
	candidateReal, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return nil, nil, err
	}
	if candidateReal != repoReal && !strings.HasPrefix(candidateReal, repoReal+string(os.PathSeparator)) {
		return nil, nil, errors.New("integration source resolves outside repository")
	}
	info, err := os.Stat(candidateReal)
	if err != nil {
		return nil, nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxIntegrationSourceBytes {
		return nil, nil, errors.New("integration source must be a regular JSON file within size limit")
	}
	data, err := os.ReadFile(candidateReal)
	if err != nil {
		return nil, nil, err
	}
	var document any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		return nil, nil, err
	}
	digest := sha256.Sum256(data)
	evidence := []Evidence{
		{ID: newID("evd"), Kind: "integration_manifest", Source: source.Manifest, Summary: "Canonical UXI source declaration observed", Status: "observed", ObservedAt: time.Now().UTC()},
		{ID: newID("evd"), Kind: "integration_source", Source: candidateReal, Summary: "Canonical JSON source read", Detail: "sha256=" + hex.EncodeToString(digest[:]), Status: "observed", ObservedAt: time.Now().UTC()},
	}
	return map[string]any{"source": source, "document": document, "sha256": hex.EncodeToString(digest[:])}, evidence, nil
}
