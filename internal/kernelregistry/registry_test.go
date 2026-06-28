package kernelregistry

import (
	"encoding/json"
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

func TestHasProvide(t *testing.T) {
	obj := Object{
		Provides: []string{"aift.module.contract", "go.module", "repository"},
	}
	if !hasProvide(obj, "aift.module.contract") {
		t.Error("should find aift.module.contract")
	}
	if !hasProvide(obj, "go.module") {
		t.Error("should find go.module")
	}
	if hasProvide(obj, "nonexistent") {
		t.Error("should not find nonexistent")
	}
	if hasProvide(Object{Provides: nil}, "anything") {
		t.Error("should not find in nil provides")
	}
}

func TestEv(t *testing.T) {
	e := ev("2024-01-01T00:00:00Z", "manifest", "/path/file", "File exists.")
	if e.Kind != "manifest" {
		t.Errorf("Kind = %q, want manifest", e.Kind)
	}
	if e.Path != "/path/file" {
		t.Errorf("Path = %q", e.Path)
	}
	if e.Description != "File exists." {
		t.Errorf("Description = %q", e.Description)
	}
	if e.ObservedAt != "2024-01-01T00:00:00Z" {
		t.Errorf("ObservedAt = %q", e.ObservedAt)
	}
}

func TestDiscoverRepositoryMinimal(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)

	obj := discoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	if obj.ID != "repository.test-repo" {
		t.Errorf("ID = %q, want repository.test-repo", obj.ID)
	}
	if obj.Kind != "repository" {
		t.Errorf("Kind = %q, want repository", obj.Kind)
	}
	if obj.Status != StatusDetected {
		t.Errorf("Status = %q, want detected (no commands)", obj.Status)
	}
	if len(obj.Evidence) != 1 {
		t.Errorf("expected 1 evidence entry (git), got %d", len(obj.Evidence))
	}
	if obj.Evidence[0].Kind != "git" {
		t.Errorf("evidence kind = %q, want git", obj.Evidence[0].Kind)
	}
	if obj.Diagnostics["git_status"] != "git status --short" {
		t.Error("should have git_status diagnostic")
	}
}

func TestDiscoverRepositoryWithGoMod(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)

	obj := discoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	if obj.Status != StatusReady {
		t.Errorf("Status = %q, want ready (has commands)", obj.Status)
	}
	if obj.Commands["go:test"] != "go test ./..." {
		t.Error("should have go:test command")
	}
	if obj.Commands["go:build"] != "go build ./..." {
		t.Error("should have go:build command")
	}
	foundGoProvide := false
	for _, p := range obj.Provides {
		if p == "go.module" {
			foundGoProvide = true
		}
	}
	if !foundGoProvide {
		t.Error("should provide go.module")
	}
}

func TestDiscoverRepositoryWithReadme(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644)

	obj := discoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	foundDocSeed := false
	for _, p := range obj.Provides {
		if p == "documentation.seed" {
			foundDocSeed = true
		}
	}
	if !foundDocSeed {
		t.Error("should provide documentation.seed")
	}
}

func TestDiscoverRepositoryWithPackageJSON(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
		"scripts": {"build": "next build"}
	}`), 0644)

	obj := discoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	foundNodePkg := false
	for _, p := range obj.Provides {
		if p == "node.package" {
			foundNodePkg = true
		}
	}
	if !foundNodePkg {
		t.Error("should provide node.package")
	}
	if obj.Commands["npm:build"] != "npm run build" {
		t.Error("should have npm:build command")
	}
}

func TestDiscoverRepositoryWithCargo(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]"), 0644)

	obj := discoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	if obj.Commands["cargo:test"] != "cargo test" {
		t.Error("should have cargo:test command")
	}
	foundRustCrate := false
	for _, p := range obj.Provides {
		if p == "rust.crate" {
			foundRustCrate = true
		}
	}
	if !foundRustCrate {
		t.Error("should provide rust.crate")
	}
}

func TestDiscoverRepositoryWithWorkflows(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.MkdirAll(filepath.Join(dir, ".github", "workflows"), 0755)

	obj := discoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	foundWorkflows := false
	for _, p := range obj.Provides {
		if p == "github.workflows" {
			foundWorkflows = true
		}
	}
	if !foundWorkflows {
		t.Error("should provide github.workflows")
	}
}

func TestDiscoverRepositoryWithAIFTModule(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.MkdirAll(filepath.Join(dir, ".aift"), 0755)
	os.WriteFile(filepath.Join(dir, ".aift", "module.json"), []byte(`{}`), 0644)

	obj := discoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	foundModuleContract := false
	for _, p := range obj.Provides {
		if p == "aift.module.contract" {
			foundModuleContract = true
		}
	}
	if !foundModuleContract {
		t.Error("should provide aift.module.contract")
	}
}

func TestDiscoverRepositoryRelationships(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)

	obj := discoverRepository("2024-01-01T00:00:00Z", "test-repo", dir)

	containedBy, ok := obj.Relationships["containedBy"]
	if !ok {
		t.Fatal("should have containedBy relationship")
	}
	if len(containedBy) != 1 || containedBy[0] != "federation.local" {
		t.Errorf("containedBy = %v, want [federation.local]", containedBy)
	}
}

func TestDiscoverModuleObjectsWithContract(t *testing.T) {
	repo := Object{
		ID:       "repository.test-repo",
		Name:     "test-repo",
		Status:   StatusReady,
		Location: "/tmp/test",
		Provides: []string{"repository", "aift.module.contract", "go.module"},
		Evidence: []Evidence{{Kind: "git", Path: "/tmp/test/.git", Description: "git", ObservedAt: "now"}},
	}

	objects := discoverModuleObjects("2024-01-01T00:00:00Z", repo)

	foundModule := false
	capCount := 0
	for _, obj := range objects {
		if obj.Kind == "module" {
			foundModule = true
			if obj.ID != "module.test-repo" {
				t.Errorf("module ID = %q, want module.test-repo", obj.ID)
			}
			if len(obj.DependsOn) != 1 || obj.DependsOn[0] != "repository.test-repo" {
				t.Errorf("module DependsOn = %v, want [repository.test-repo]", obj.DependsOn)
			}
		}
		if obj.Kind == "capability" {
			capCount++
		}
	}
	if !foundModule {
		t.Error("should create module object when aift.module.contract is provided")
	}
	if capCount != 3 {
		t.Errorf("expected 3 capability objects (one per provide), got %d", capCount)
	}
}

func TestDiscoverModuleObjectsWithoutContract(t *testing.T) {
	repo := Object{
		ID:       "repository.test-repo",
		Name:     "test-repo",
		Status:   StatusDetected,
		Provides: []string{"repository"},
		Evidence: []Evidence{},
	}

	objects := discoverModuleObjects("2024-01-01T00:00:00Z", repo)

	for _, obj := range objects {
		if obj.Kind == "module" {
			t.Error("should not create module object without aift.module.contract")
		}
	}
	capCount := 0
	for _, obj := range objects {
		if obj.Kind == "capability" {
			capCount++
		}
	}
	if capCount != 1 {
		t.Errorf("expected 1 capability object for 'repository', got %d", capCount)
	}
}

func TestDiscoverModuleObjectsCapabilityFields(t *testing.T) {
	repo := Object{
		ID:       "repository.test",
		Name:     "test",
		Status:   StatusReady,
		Location: "/tmp/test",
		Provides: []string{"go.module"},
		Evidence: []Evidence{{Kind: "manifest", Path: "/tmp/go.mod", Description: "test", ObservedAt: "now"}},
	}

	objects := discoverModuleObjects("2024-01-01T00:00:00Z", repo)

	if len(objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(objects))
	}
	cap := objects[0]
	if cap.ID != "capability.test.go.module" {
		t.Errorf("capability ID = %q, want capability.test.go.module", cap.ID)
	}
	if cap.Kind != "capability" {
		t.Errorf("Kind = %q, want capability", cap.Kind)
	}
	if cap.Status != StatusReady {
		t.Errorf("Status = %q, want ready (inherits from repo)", cap.Status)
	}
	if len(cap.DependsOn) != 1 || cap.DependsOn[0] != "repository.test" {
		t.Errorf("DependsOn = %v, want [repository.test]", cap.DependsOn)
	}
}

func TestStatusConstants(t *testing.T) {
	if StatusPlanned != "planned" {
		t.Error("StatusPlanned mismatch")
	}
	if StatusDetected != "detected" {
		t.Error("StatusDetected mismatch")
	}
	if StatusReady != "ready" {
		t.Error("StatusReady mismatch")
	}
	if StatusActive != "active" {
		t.Error("StatusActive mismatch")
	}
	if StatusDeprecated != "deprecated" {
		t.Error("StatusDeprecated mismatch")
	}
	if StatusRemoved != "removed" {
		t.Error("StatusRemoved mismatch")
	}
}

func TestRegistryJSON(t *testing.T) {
	reg := Registry{
		SchemaVersion: "aift.kernel.registry.v1",
		GeneratedAt:   "2024-01-01T00:00:00Z",
		Source:        "test",
		Objects: []Object{
			{
				ID:            "test.obj",
				Kind:          "test",
				Name:          "test",
				Status:        StatusDetected,
				Location:      "/tmp",
				Version:       "0.1.0",
				Evidence:      []Evidence{},
				Provides:      []string{},
				Consumes:      []string{},
				DependsOn:     []string{},
				Publishes:     []string{},
				Subscribes:    []string{},
				Commands:      map[string]string{},
				Diagnostics:   map[string]string{},
				Relationships: map[string][]string{},
				GeneratedAt:   "2024-01-01T00:00:00Z",
			},
		},
	}

	data, err := json.Marshal(reg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Registry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.SchemaVersion != "aift.kernel.registry.v1" {
		t.Errorf("SchemaVersion = %q", decoded.SchemaVersion)
	}
	if len(decoded.Objects) != 1 {
		t.Errorf("Objects count = %d, want 1", len(decoded.Objects))
	}
}
