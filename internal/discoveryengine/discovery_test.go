package discoveryengine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSafeID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"My Repo", "my-repo"},
		{"org/repo-name", "org.repo-name"},
		{"under_score", "under-score"},
		{"UPPER", "upper"},
		{"simple", "simple"},
		{"", ""},
	}
	for _, tt := range tests {
		got := safeID(tt.input)
		if got != tt.want {
			t.Errorf("safeID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRuntimeNames(t *testing.T) {
	runtimes := []Runtime{
		{Name: "node", Kind: "javascript"},
		{Name: "go", Kind: "go"},
		{Name: "docker", Kind: "container"},
	}
	got := runtimeNames(runtimes)
	if got != "docker,go,node" {
		t.Errorf("runtimeNames = %q, want %q", got, "docker,go,node")
	}
}

func TestRuntimeNamesEmpty(t *testing.T) {
	got := runtimeNames(nil)
	if got != "" {
		t.Errorf("runtimeNames(nil) = %q, want empty", got)
	}
}

func TestRuntimeNamesDeduplicated(t *testing.T) {
	runtimes := []Runtime{
		{Name: "go", Kind: "go"},
		{Name: "go", Kind: "go"},
	}
	got := runtimeNames(runtimes)
	if got != "go" {
		t.Errorf("runtimeNames with duplicates = %q, want %q", got, "go")
	}
}

func TestAddEvidence(t *testing.T) {
	obj := &DiscoveryObject{Evidence: []Evidence{}}
	addEvidence(obj, "2024-01-01T00:00:00Z", "git", "/path", "desc")

	if len(obj.Evidence) != 1 {
		t.Fatalf("expected 1 evidence, got %d", len(obj.Evidence))
	}
	if obj.Evidence[0].Kind != "git" {
		t.Errorf("evidence kind = %q, want git", obj.Evidence[0].Kind)
	}
	if obj.Evidence[0].Path != "/path" {
		t.Errorf("evidence path = %q, want /path", obj.Evidence[0].Path)
	}
}

func TestEv(t *testing.T) {
	e := ev("2024-01-01T00:00:00Z", "manifest", "/file", "File exists.")
	if e.Kind != "manifest" || e.Path != "/file" || e.Description != "File exists." || e.ObservedAt != "2024-01-01T00:00:00Z" {
		t.Errorf("ev() = %+v", e)
	}
}

func TestDiscoverRepositoryBasic(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)

	obj := DiscoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	if obj.ID != "repository.test-repo" {
		t.Errorf("ID = %q, want repository.test-repo", obj.ID)
	}
	if obj.Kind != "repository" {
		t.Errorf("Kind = %q, want repository", obj.Kind)
	}
	if obj.Name != "test-repo" {
		t.Errorf("Name = %q, want test-repo", obj.Name)
	}
	if obj.Status != "detected" {
		t.Errorf("Status = %q, want detected", obj.Status)
	}
}

func TestDiscoverRepositoryWithGoMod(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)

	obj := DiscoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	if obj.Status != "ready" {
		t.Errorf("status with go.mod should be ready, got %q", obj.Status)
	}
	if obj.Commands["go:test"] != "go test ./..." {
		t.Error("should have go:test command")
	}
	if obj.Commands["go:build"] != "go build ./..." {
		t.Error("should have go:build command")
	}

	foundGoRuntime := false
	for _, rt := range obj.Runtimes {
		if rt.Name == "go" {
			foundGoRuntime = true
		}
	}
	if !foundGoRuntime {
		t.Error("should detect Go runtime")
	}
}

func TestDiscoverRepositoryWithDocs(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644)
	os.MkdirAll(filepath.Join(dir, "docs"), 0755)

	obj := DiscoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	foundReadme := false
	foundDocs := false
	for _, d := range obj.Docs {
		if d == "README.md" {
			foundReadme = true
		}
		if d == "docs" {
			foundDocs = true
		}
	}
	if !foundReadme {
		t.Error("should discover README.md")
	}
	if !foundDocs {
		t.Error("should discover docs dir")
	}
}

func TestDiscoverRepositoryWithSchemas(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.MkdirAll(filepath.Join(dir, "schemas"), 0755)

	obj := DiscoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	found := false
	for _, s := range obj.Schemas {
		if s == "schemas" {
			found = true
		}
	}
	if !found {
		t.Error("should discover schemas dir")
	}
}

func TestDiscoverRepositoryWithWorkflows(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.MkdirAll(filepath.Join(dir, ".github", "workflows"), 0755)

	obj := DiscoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	found := false
	for _, w := range obj.Workflows {
		if w == ".github/workflows" {
			found = true
		}
	}
	if !found {
		t.Error("should discover .github/workflows")
	}
}

func TestDiscoverRepositoryWithManifests(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"scripts":{"build":"next build"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM alpine"), 0644)

	obj := DiscoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	foundPkg := false
	foundDockerfile := false
	for _, m := range obj.Manifests {
		if m == "package.json" {
			foundPkg = true
		}
		if m == "Dockerfile" {
			foundDockerfile = true
		}
	}
	if !foundPkg {
		t.Error("should discover package.json manifest")
	}
	if !foundDockerfile {
		t.Error("should discover Dockerfile manifest")
	}
}

func TestDiscoverRepositoryWithAIFTContracts(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.MkdirAll(filepath.Join(dir, ".aift", "commands"), 0755)
	os.WriteFile(filepath.Join(dir, ".aift", "module.json"), []byte(`{}`), 0644)
	os.WriteFile(filepath.Join(dir, ".aift", "commands", "verify.sh"), []byte("#!/bin/sh\nexit 0"), 0644)

	obj := DiscoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	foundModule := false
	for _, c := range obj.Capabilities {
		if c == "aift.module" {
			foundModule = true
		}
	}
	if !foundModule {
		t.Error("should discover aift.module capability")
	}
	if obj.Commands["aift:verify"] != "sh .aift/commands/verify.sh" {
		t.Error("should discover aift:verify command")
	}
	foundVerifyHealth := false
	for _, h := range obj.HealthChecks {
		if h == ".aift/commands/verify.sh" {
			foundVerifyHealth = true
		}
	}
	if !foundVerifyHealth {
		t.Error("should add verify.sh to health checks")
	}
}
