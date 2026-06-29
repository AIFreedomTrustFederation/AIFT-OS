package manifests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

func TestPath(t *testing.T) {
	got := Path("/home/user/repo")
	want := filepath.Join("/home/user/repo", ".aift", "repo.json")
	if got != want {
		t.Errorf("Path = %q, want %q", got, want)
	}
}

func TestEnsureCreatesManifest(t *testing.T) {
	dir := t.TempDir()
	repo := workspace.Repo{Name: "test-repo", Path: dir}

	err := Ensure(repo)
	if err != nil {
		t.Fatalf("Ensure failed: %v", err)
	}

	data, err := os.ReadFile(Path(dir))
	if err != nil {
		t.Fatalf("manifest not created: %v", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if m.Name != "test-repo" {
		t.Errorf("Name = %q, want test-repo", m.Name)
	}
	if m.Role != "sovereign-repository" {
		t.Errorf("Role = %q, want sovereign-repository", m.Role)
	}
	if !m.Sovereign {
		t.Error("Sovereign should be true")
	}
	if m.ManagedBy != "AIFT-OS" {
		t.Errorf("ManagedBy = %q", m.ManagedBy)
	}
	if m.CommandsPath != ".aift/commands" {
		t.Errorf("CommandsPath = %q", m.CommandsPath)
	}
}

func TestEnsureCreatesCommandsDir(t *testing.T) {
	dir := t.TempDir()
	repo := workspace.Repo{Name: "test-repo", Path: dir}

	Ensure(repo)

	cmdDir := filepath.Join(dir, ".aift", "commands")
	info, err := os.Stat(cmdDir)
	if err != nil || !info.IsDir() {
		t.Error("commands directory should be created")
	}
}

func TestEnsureAIFTOSRole(t *testing.T) {
	dir := t.TempDir()
	repo := workspace.Repo{Name: "AIFT-OS", Path: dir}

	Ensure(repo)

	data, _ := os.ReadFile(Path(dir))
	var m Manifest
	json.Unmarshal(data, &m)

	if m.Role != "federation-control-plane" {
		t.Errorf("AIFT-OS role = %q, want federation-control-plane", m.Role)
	}
}

func TestEnsureIdempotent(t *testing.T) {
	dir := t.TempDir()
	repo := workspace.Repo{Name: "test-repo", Path: dir}

	Ensure(repo)

	// Modify the file
	os.WriteFile(Path(dir), []byte(`{"name":"modified","role":"custom"}`), 0644)

	// Ensure should not overwrite
	Ensure(repo)

	data, _ := os.ReadFile(Path(dir))
	var m Manifest
	json.Unmarshal(data, &m)
	if m.Name != "modified" {
		t.Error("Ensure should not overwrite existing manifest")
	}
}

func TestValidTrue(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "repo.json"), []byte(`{
		"name": "test",
		"role": "sovereign-repository"
	}`), 0644)

	if !Valid(dir) {
		t.Error("Valid should return true for valid manifest")
	}
}

func TestValidMissingFile(t *testing.T) {
	dir := t.TempDir()
	if Valid(dir) {
		t.Error("Valid should return false for missing file")
	}
}

func TestValidInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "repo.json"), []byte("not json"), 0644)

	if Valid(dir) {
		t.Error("Valid should return false for invalid JSON")
	}
}

func TestValidMissingName(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "repo.json"), []byte(`{"role":"test"}`), 0644)

	if Valid(dir) {
		t.Error("Valid should return false when name is empty")
	}
}

func TestValidMissingRole(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "repo.json"), []byte(`{"name":"test"}`), 0644)

	if Valid(dir) {
		t.Error("Valid should return false when role is empty")
	}
}

func TestManifestJSON(t *testing.T) {
	m := Manifest{
		Name:         "test",
		Role:         "sovereign-repository",
		Sovereign:    true,
		ManagedBy:    "AIFT-OS",
		Dependencies: []string{"dep1"},
		Capabilities: []string{"build"},
		CommandsPath: ".aift/commands",
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Manifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Name != "test" || !decoded.Sovereign || len(decoded.Dependencies) != 1 {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}
