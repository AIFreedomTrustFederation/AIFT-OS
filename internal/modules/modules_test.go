package modules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInferKindOS(t *testing.T) {
	dir := t.TempDir()
	got := inferKind("AIFT-OS", dir)
	if got != "kernel" {
		t.Errorf("inferKind(AIFT-OS) = %q, want kernel", got)
	}
}

func TestInferKindForge(t *testing.T) {
	dir := t.TempDir()
	got := inferKind("AIFT-Forge", dir)
	if got != "forge" {
		t.Errorf("inferKind(AIFT-Forge) = %q, want forge", got)
	}
}

func TestInferKindBooksmith(t *testing.T) {
	dir := t.TempDir()
	got := inferKind("BookSmith-AI", dir)
	if got != "publishing" {
		t.Errorf("inferKind(BookSmith-AI) = %q, want publishing", got)
	}
}

func TestInferKindWebsite(t *testing.T) {
	dir := t.TempDir()
	if inferKind("www-portal", dir) != "website" {
		t.Error("expected website for www-portal")
	}
	if inferKind("my-github.io", dir) != "website" {
		t.Error("expected website for github.io")
	}
}

func TestInferKindGoModule(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	got := inferKind("my-service", dir)
	if got != "go-module" {
		t.Errorf("inferKind with go.mod = %q, want go-module", got)
	}
}

func TestInferKindNodeApp(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)
	got := inferKind("my-app", dir)
	if got != "node-app" {
		t.Errorf("inferKind with package.json = %q, want node-app", got)
	}
}

func TestInferKindRustCrate(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]"), 0644)
	got := inferKind("my-crate", dir)
	if got != "rust-crate" {
		t.Errorf("inferKind with Cargo.toml = %q, want rust-crate", got)
	}
}

func TestInferKindDefault(t *testing.T) {
	dir := t.TempDir()
	got := inferKind("random-repo", dir)
	if got != "repository" {
		t.Errorf("inferKind(random-repo) = %q, want repository", got)
	}
}

func TestBuildRepoManifestMinimal(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)

	m := BuildRepoManifest("test-repo", dir)

	if m.ID != "repo.test-repo" {
		t.Errorf("ID = %q, want repo.test-repo", m.ID)
	}
	if m.Name != "test-repo" {
		t.Errorf("Name = %q, want test-repo", m.Name)
	}
	if m.Version != "0.1.0" {
		t.Errorf("Version = %q, want 0.1.0", m.Version)
	}
	if m.Status != "planned" {
		t.Errorf("Status = %q, want planned (no capabilities/services/commands)", m.Status)
	}
	if m.Kind != "repository" {
		t.Errorf("Kind = %q, want repository", m.Kind)
	}
	if m.GeneratedAt == "" {
		t.Error("GeneratedAt should not be empty")
	}
	if len(m.DependsOn) != 0 {
		t.Errorf("DependsOn should be empty, got %v", m.DependsOn)
	}
}

func TestBuildRepoManifestWithGoMod(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)

	m := BuildRepoManifest("test-repo", dir)

	if m.Commands["go:test"] != "go test ./..." {
		t.Error("should have go:test command")
	}
	if m.Commands["go:build"] != "go build ./..." {
		t.Error("should have go:build command")
	}
	if m.Status != "detected" {
		t.Errorf("Status = %q, want detected (has commands but no verify)", m.Status)
	}
	found := false
	for _, p := range m.Provides {
		if p == "go.module" {
			found = true
		}
	}
	if !found {
		t.Error("should provide go.module")
	}
}

func TestBuildRepoManifestWithReadmeAndDocs(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644)
	os.MkdirAll(filepath.Join(dir, "docs"), 0755)

	m := BuildRepoManifest("test-repo", dir)

	foundReadme := false
	foundDocs := false
	for _, d := range m.Docs {
		if d == "README.md" {
			foundReadme = true
		}
		if d == "docs/" {
			foundDocs = true
		}
	}
	if !foundReadme {
		t.Error("should have README.md in docs")
	}
	if !foundDocs {
		t.Error("should have docs/ in docs")
	}
}

func TestBuildRepoManifestWithCapabilities(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "capabilities.json"), []byte(`{
		"capabilities": [
			{"name": "build", "status": "ready"},
			{"name": "test", "status": "v1"}
		]
	}`), 0644)

	m := BuildRepoManifest("test-repo", dir)

	if len(m.Capabilities) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(m.Capabilities))
	}
}

func TestBuildRepoManifestWithServices(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "services.json"), []byte(`{
		"services": [{"name": "api-svc"}]
	}`), 0644)

	m := BuildRepoManifest("test-repo", dir)

	foundSvc := false
	for _, s := range m.Services {
		if s == "api-svc" {
			foundSvc = true
		}
	}
	if !foundSvc {
		t.Error("should discover api-svc service")
	}
	foundProvide := false
	for _, p := range m.Provides {
		if p == "service.contract" {
			foundProvide = true
		}
	}
	if !foundProvide {
		t.Error("should provide service.contract")
	}
}

func TestBuildRepoManifestWithVerify(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.MkdirAll(filepath.Join(dir, ".aift", "commands"), 0755)
	os.WriteFile(filepath.Join(dir, ".aift", "commands", "verify.sh"), []byte("#!/bin/sh\nexit 0"), 0644)

	m := BuildRepoManifest("test-repo", dir)

	if m.Status != "ready" {
		t.Errorf("Status = %q, want ready (has verify command)", m.Status)
	}
	if m.Commands["aift:verify"] != "sh .aift/commands/verify.sh" {
		t.Error("should have aift:verify command")
	}
	foundHealth := false
	for _, h := range m.HealthChecks {
		if h == ".aift/commands/verify.sh" {
			foundHealth = true
		}
	}
	if !foundHealth {
		t.Error("should have verify.sh in health checks")
	}
}

func TestBuildRepoManifestWithVerifyCapability(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "capabilities.json"), []byte(`{
		"capabilities": [{"name": "verify", "status": "ready"}]
	}`), 0644)

	m := BuildRepoManifest("test-repo", dir)

	if m.Status != "ready" {
		t.Errorf("Status = %q, want ready (has verify capability)", m.Status)
	}
}

func TestBuildRepoManifestWithPackageJSON(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
		"scripts": {"build": "next build", "dev": "next dev"}
	}`), 0644)

	m := BuildRepoManifest("test-repo", dir)

	if m.Commands["npm:build"] != "npm run build" {
		t.Error("should have npm:build command")
	}
	if m.Commands["npm:dev"] != "npm run dev" {
		t.Error("should have npm:dev command")
	}
}

func TestBuildRepoManifestWithManualJSON(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "manual.json"), []byte(`{}`), 0644)

	m := BuildRepoManifest("test-repo", dir)

	foundManual := false
	for _, p := range m.Provides {
		if p == "manual.contract" {
			foundManual = true
		}
	}
	if !foundManual {
		t.Error("should provide manual.contract")
	}
}

func TestBuildRepoManifestJSON(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)

	m := BuildRepoManifest("test-repo", dir)

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded ModuleManifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.ID != m.ID {
		t.Errorf("round-trip ID mismatch: %q != %q", decoded.ID, m.ID)
	}
}

func TestBuildRepoManifestEvidenceIncludesGit(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)

	m := BuildRepoManifest("test-repo", dir)

	found := false
	for _, e := range m.Evidence {
		if e == ".git" {
			found = true
		}
	}
	if !found {
		t.Error("evidence should include .git")
	}
}

func TestBuildRepoManifestMigrationLevel(t *testing.T) {
	dir := t.TempDir()
	m := BuildRepoManifest("test-repo", dir)

	if m.MigrationLevel != "phase-17" {
		t.Errorf("MigrationLevel = %q, want phase-17", m.MigrationLevel)
	}
}

func TestBuildRepoManifestWithCargo(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]"), 0644)

	m := BuildRepoManifest("test-repo", dir)

	if m.Commands["cargo:test"] != "cargo test" {
		t.Error("should have cargo:test command")
	}
	if m.Commands["cargo:build"] != "cargo build" {
		t.Error("should have cargo:build command")
	}
	foundProvide := false
	for _, p := range m.Provides {
		if p == "rust.crate" {
			foundProvide = true
		}
	}
	if !foundProvide {
		t.Error("should provide rust.crate")
	}
}
