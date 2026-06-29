package capabilities

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestCapabilityNames(t *testing.T) {
	names := capabilityNames()
	if len(names) != 10 {
		t.Errorf("expected 10 capability names, got %d", len(names))
	}
	expected := map[string]bool{
		"status": true, "verify": true, "test": true, "build": true,
		"start": true, "stop": true, "health": true, "deploy": true,
		"sync": true, "docs": true,
	}
	for _, n := range names {
		if !expected[n] {
			t.Errorf("unexpected capability name: %s", n)
		}
	}
}

func TestDescription(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"status", "Report repository state"},
		{"verify", "Validate repository health"},
		{"test", "Run test suite"},
		{"build", "Build project artifacts"},
		{"start", "Start local service"},
		{"stop", "Stop local service"},
		{"health", "Check local service health"},
		{"deploy", "Deploy project"},
		{"sync", "Synchronize safely"},
		{"docs", "Documentation present or generated"},
		{"unknown", "Capability"},
	}
	for _, tt := range tests {
		got := description(tt.name)
		if got != tt.want {
			t.Errorf("description(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestDetectCapabilityStatusWithGit(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)

	cap := detectCapability(dir, "status")
	if cap.Name != "status" {
		t.Errorf("Name = %q", cap.Name)
	}
	if cap.Status != StatusReady {
		t.Errorf("Status = %q, want ready (git detected)", cap.Status)
	}
	if cap.Command != "git status --short" {
		t.Errorf("Command = %q", cap.Command)
	}
}

func TestDetectCapabilitySyncWithGit(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)

	cap := detectCapability(dir, "sync")
	if cap.Status != StatusReady {
		t.Errorf("Status = %q, want ready (git detected)", cap.Status)
	}
}

func TestDetectCapabilityDocsWithReadme(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644)

	cap := detectCapability(dir, "docs")
	if cap.Status != StatusDetected {
		t.Errorf("Status = %q, want detected (README.md)", cap.Status)
	}
}

func TestDetectCapabilityDocsWithDocsDir(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "docs"), 0755)

	cap := detectCapability(dir, "docs")
	if cap.Status != StatusDetected {
		t.Errorf("Status = %q, want detected (docs dir)", cap.Status)
	}
}

func TestDetectCapabilityTestWithGoMod(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)

	cap := detectCapability(dir, "test")
	if cap.Status != StatusDetected {
		t.Errorf("Status = %q, want detected (go.mod)", cap.Status)
	}
}

func TestDetectCapabilityTestWithPackageJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)

	cap := detectCapability(dir, "test")
	if cap.Status != StatusDetected {
		t.Errorf("Status = %q, want detected (package.json)", cap.Status)
	}
}

func TestDetectCapabilityBuildWithGoMod(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)

	cap := detectCapability(dir, "build")
	if cap.Status != StatusDetected {
		t.Errorf("Status = %q, want detected (go.mod for build)", cap.Status)
	}
}

func TestDetectCapabilityBuildWithMakefile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte("all:"), 0644)

	cap := detectCapability(dir, "build")
	if cap.Status != StatusDetected {
		t.Errorf("Status = %q, want detected (Makefile for build)", cap.Status)
	}
}

func TestDetectCapabilityPlannedNoEvidence(t *testing.T) {
	dir := t.TempDir()

	cap := detectCapability(dir, "deploy")
	if cap.Status != StatusPlanned {
		t.Errorf("Status = %q, want planned (no evidence)", cap.Status)
	}
	if cap.Evidence != "not proven yet" {
		t.Errorf("Evidence = %q, want 'not proven yet'", cap.Evidence)
	}
}

func TestDetectCapabilityWithCommand(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, ".aift", "commands")
	os.MkdirAll(cmdDir, 0755)
	// Create a verify.sh that succeeds
	os.WriteFile(filepath.Join(cmdDir, "verify.sh"), []byte("#!/bin/sh\nexit 0\n"), 0755)

	cap := detectCapability(dir, "verify")
	if cap.Command != ".aift/commands/verify.sh" {
		t.Errorf("Command = %q, want .aift/commands/verify.sh", cap.Command)
	}
	if cap.Status != StatusReady {
		t.Errorf("Status = %q, want ready (command passes)", cap.Status)
	}
}

func TestDetectCapabilityWithFailingCommand(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, ".aift", "commands")
	os.MkdirAll(cmdDir, 0755)
	os.WriteFile(filepath.Join(cmdDir, "verify.sh"), []byte("#!/bin/sh\nexit 1\n"), 0755)

	cap := detectCapability(dir, "verify")
	if cap.Status != StatusBroken {
		t.Errorf("Status = %q, want broken (command fails)", cap.Status)
	}
}

func TestCommandPassesPathTraversal(t *testing.T) {
	dir := t.TempDir()
	got := commandPasses(dir, "/etc/passwd")
	if got {
		t.Error("commandPasses should reject paths outside repo")
	}
}

func TestCommandPassesRelativeTraversal(t *testing.T) {
	dir := t.TempDir()
	got := commandPasses(dir, "../../../etc/passwd")
	if got {
		t.Error("commandPasses should reject relative path traversal")
	}
}

func TestReadExistingValid(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	rc := RepoCapabilities{
		Repo: "test",
		Capabilities: []Capability{
			{Name: "build", Status: StatusReady, Version: 0},
			{Name: "test", Status: StatusV1, Version: 1},
		},
	}
	data, _ := json.MarshalIndent(rc, "", "  ")
	os.WriteFile(filepath.Join(aiftDir, "capabilities.json"), data, 0644)

	result := readExisting(dir)
	if len(result) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(result))
	}
	if result["build"].Status != StatusReady {
		t.Errorf("build status = %q, want ready", result["build"].Status)
	}
	if result["test"].Status != StatusV1 {
		t.Errorf("test status = %q, want v1", result["test"].Status)
	}
}

func TestReadExistingMissing(t *testing.T) {
	dir := t.TempDir()
	result := readExisting(dir)
	if len(result) != 0 {
		t.Errorf("expected empty for missing file, got %d", len(result))
	}
}

func TestReadExistingInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "capabilities.json"), []byte("not json"), 0644)

	result := readExisting(dir)
	if len(result) != 0 {
		t.Errorf("expected empty for invalid JSON, got %d", len(result))
	}
}

func TestStatusConstants(t *testing.T) {
	if StatusPlanned != "planned" {
		t.Error("StatusPlanned")
	}
	if StatusDetected != "detected" {
		t.Error("StatusDetected")
	}
	if StatusReady != "ready" {
		t.Error("StatusReady")
	}
	if StatusV1 != "v1" {
		t.Error("StatusV1")
	}
	if StatusBroken != "broken" {
		t.Error("StatusBroken")
	}
	if StatusMissing != "missing" {
		t.Error("StatusMissing")
	}
}

func TestCapabilityJSON(t *testing.T) {
	c := Capability{
		Name:        "build",
		Status:      StatusReady,
		Version:     1,
		Command:     "go build",
		Evidence:    "test",
		Description: "Build",
		LastChecked: "2024-01-01",
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Capability
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Name != "build" || decoded.Version != 1 {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestRepoCapabilitiesJSON(t *testing.T) {
	rc := RepoCapabilities{
		Repo: "test-repo",
		Path: "/tmp/test",
		Capabilities: []Capability{
			{Name: "build", Status: StatusReady},
		},
	}
	data, err := json.Marshal(rc)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded RepoCapabilities
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Repo != "test-repo" || len(decoded.Capabilities) != 1 {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestFederationCapabilitiesJSON(t *testing.T) {
	fc := FederationCapabilities{
		GeneratedAt: "2024-01-01",
		Repos:       []RepoCapabilities{},
	}
	data, err := json.Marshal(fc)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded FederationCapabilities
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.GeneratedAt != "2024-01-01" {
		t.Errorf("GeneratedAt = %q", decoded.GeneratedAt)
	}
}

func TestWriteReport(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}

	fc := FederationCapabilities{
		GeneratedAt: "2024-01-01",
		Repos: []RepoCapabilities{
			{
				Repo: "test-repo",
				Capabilities: []Capability{
					{Name: "build", Status: StatusReady, Version: 1, Evidence: "test"},
				},
			},
		},
	}

	err := writeReport(cfg, fc)
	if err != nil {
		t.Fatalf("writeReport failed: %v", err)
	}

	reportPath := filepath.Join(dir, "reports", "capabilities.md")
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("report not created: %v", err)
	}
	content := string(data)
	if !containsStr(content, "test-repo") {
		t.Error("report should contain repo name")
	}
	if !containsStr(content, "build") {
		t.Error("report should contain capability name")
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
