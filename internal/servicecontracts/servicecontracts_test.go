package servicecontracts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInferKindControlPlane(t *testing.T) {
	dir := t.TempDir()
	if inferKind("AIFT-OS", dir) != "control-plane" {
		t.Errorf("inferKind(AIFT-OS) = %q, want control-plane", inferKind("AIFT-OS", dir))
	}
	if inferKind("aift-os-dev", dir) != "control-plane" {
		t.Errorf("inferKind(aift-os-dev) = %q, want control-plane", inferKind("aift-os-dev", dir))
	}
}

func TestInferKindForge(t *testing.T) {
	dir := t.TempDir()
	got := inferKind("AIFT-Forge", dir)
	if got != "forge" {
		t.Errorf("inferKind(AIFT-Forge) = %q, want forge", got)
	}
}

func TestInferKindPublishing(t *testing.T) {
	dir := t.TempDir()
	got := inferKind("BookSmith-AI", dir)
	if got != "publishing" {
		t.Errorf("inferKind(BookSmith-AI) = %q, want publishing", got)
	}
}

func TestInferKindInfrastructure(t *testing.T) {
	dir := t.TempDir()
	got := inferKind("my-VPS-node", dir)
	if got != "infrastructure" {
		t.Errorf("inferKind(my-VPS-node) = %q, want infrastructure", got)
	}
}

func TestInferKindWebsite(t *testing.T) {
	dir := t.TempDir()
	if inferKind("www-portal", dir) != "website" {
		t.Error("expected website for www-portal")
	}
	if inferKind("org.github.io", dir) != "website" {
		t.Error("expected website for github.io")
	}
}

func TestInferKindGoService(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	got := inferKind("my-svc", dir)
	if got != "go-service" {
		t.Errorf("inferKind with go.mod = %q, want go-service", got)
	}
}

func TestInferKindWebApp(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)
	got := inferKind("my-app", dir)
	if got != "web-app" {
		t.Errorf("inferKind with package.json = %q, want web-app", got)
	}
}

func TestInferKindDefault(t *testing.T) {
	dir := t.TempDir()
	got := inferKind("random-repo", dir)
	if got != "repository" {
		t.Errorf("inferKind(random-repo) = %q, want repository", got)
	}
}

func TestReadContract(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "services.json"), []byte(`{
		"repo": "test-repo",
		"services": [
			{
				"name": "api-svc",
				"kind": "http",
				"status": "ready",
				"version": "1.0",
				"provides": ["api"],
				"requires": ["db"],
				"events": ["repo.changed"],
				"evidence": "test"
			}
		]
	}`), 0644)

	c, ok := readContract("test-repo", dir)
	if !ok {
		t.Fatal("readContract should succeed")
	}
	if c.Repo != "test-repo" {
		t.Errorf("Repo = %q, want test-repo", c.Repo)
	}
	if len(c.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(c.Services))
	}
	svc := c.Services[0]
	if svc.Name != "api-svc" {
		t.Errorf("service name = %q, want api-svc", svc.Name)
	}
	if svc.Kind != "http" {
		t.Errorf("service kind = %q, want http", svc.Kind)
	}
}

func TestReadContractMissing(t *testing.T) {
	dir := t.TempDir()
	_, ok := readContract("test-repo", dir)
	if ok {
		t.Error("readContract should fail for missing file")
	}
}

func TestReadContractInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "services.json"), []byte("not json"), 0644)

	_, ok := readContract("test-repo", dir)
	if ok {
		t.Error("readContract should fail for invalid JSON")
	}
}

func TestReadContractDefaultsRepo(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "services.json"), []byte(`{
		"services": [{"name": "svc"}]
	}`), 0644)

	c, ok := readContract("fallback-name", dir)
	if !ok {
		t.Fatal("readContract should succeed")
	}
	if c.Repo != "fallback-name" {
		t.Errorf("Repo = %q, want fallback-name (defaulted from param)", c.Repo)
	}
}

func TestServiceJSON(t *testing.T) {
	svc := Service{
		Name:     "api",
		Kind:     "http",
		Status:   "ready",
		Version:  "1.0",
		Provides: []string{"api"},
		Requires: []string{"db"},
		Events:   []string{"repo.changed"},
		Health:   "curl localhost:8080/health",
		Start:    "go run .",
		Stop:     "kill $PID",
		Evidence: "discovered",
	}

	data, err := json.Marshal(svc)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Service
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Name != "api" {
		t.Errorf("Name = %q", decoded.Name)
	}
	if decoded.Health != "curl localhost:8080/health" {
		t.Errorf("Health = %q", decoded.Health)
	}
}

func TestServiceRecordJSON(t *testing.T) {
	sr := ServiceRecord{
		Repo:     "test-repo",
		Name:     "api-svc",
		Kind:     "http",
		Status:   "ready",
		Version:  "1.0",
		Evidence: "test",
	}

	data, err := json.Marshal(sr)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded ServiceRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Repo != "test-repo" || decoded.Name != "api-svc" {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestContractJSON(t *testing.T) {
	c := Contract{
		Repo: "test",
		Services: []Service{
			{Name: "svc", Kind: "http", Status: "planned", Version: "0.1.0",
				Provides: []string{}, Requires: []string{}, Events: []string{},
				Evidence: "test"},
		},
		GeneratedAt: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Contract
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Repo != "test" || len(decoded.Services) != 1 {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestRegistryJSON(t *testing.T) {
	reg := Registry{
		GeneratedAt: "2024-01-01T00:00:00Z",
		Contracts:   []Contract{},
		Services:    []ServiceRecord{},
	}

	data, err := json.Marshal(reg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Registry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.GeneratedAt != "2024-01-01T00:00:00Z" {
		t.Errorf("GeneratedAt = %q", decoded.GeneratedAt)
	}
}
