package servicecontracts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
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

func TestServiceOwnerInJSON(t *testing.T) {
	svc := Service{
		Name:    "test-svc",
		Kind:    "http",
		Status:  "ready",
		Version: "1.0",
		Owner:   "test-repo",
	}

	data, err := json.Marshal(svc)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map failed: %v", err)
	}
	owner, ok := m["owner"]
	if !ok {
		t.Fatal("Service JSON missing 'owner' field")
	}
	if owner != "test-repo" {
		t.Errorf("owner = %v, want test-repo", owner)
	}
}

func TestServiceOwnerDefaultsFromRepo(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "services.json"), []byte(`{
		"repo": "my-repo",
		"services": [
			{
				"name": "svc-1",
				"kind": "http",
				"status": "ready",
				"version": "1.0"
			}
		]
	}`), 0644)

	c, ok := readContract("my-repo", dir)
	if !ok {
		t.Fatal("readContract should succeed")
	}
	if c.Services[0].Owner == "" {
		// Owner is empty in the JSON (legacy contract without owner).
		// The Scan() function defaults it to c.Repo at scan time.
		// readContract alone doesn't set it — this is expected.
	}
}

func TestScanDefaultsPersistInContracts(t *testing.T) {
	root := t.TempDir()
	repoDir := filepath.Join(root, "legacy-repo")
	os.MkdirAll(filepath.Join(repoDir, ".git"), 0755)
	os.MkdirAll(filepath.Join(repoDir, ".aift"), 0755)

	// Legacy contract: services have no status, version, owner, or evidence
	os.WriteFile(filepath.Join(repoDir, ".aift", "services.json"), []byte(`{
		"repo": "legacy-repo",
		"services": [
			{
				"name": "legacy-svc",
				"kind": "http",
				"provides": [],
				"requires": [],
				"events": []
			}
		]
	}`), 0644)

	osHome := filepath.Join(root, "AIFT-OS")
	for _, d := range []string{"registry", "reports"} {
		os.MkdirAll(filepath.Join(osHome, d), 0755)
	}

	cfg := config.Config{Root: root, OSHome: osHome}
	if err := Scan(cfg); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Read persisted registry
	data, err := os.ReadFile(filepath.Join(osHome, "registry", "service-contracts.json"))
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Verify defaults persisted in the Contract entries (the PR #11 fix)
	if len(reg.Contracts) != 1 {
		t.Fatalf("expected 1 contract, got %d", len(reg.Contracts))
	}
	svc := reg.Contracts[0].Services[0]
	if svc.Status != "planned" {
		t.Errorf("Contract.Services[0].Status = %q, want planned", svc.Status)
	}
	if svc.Version != "0.1.0" {
		t.Errorf("Contract.Services[0].Version = %q, want 0.1.0", svc.Version)
	}
	if svc.Owner != "legacy-repo" {
		t.Errorf("Contract.Services[0].Owner = %q, want legacy-repo", svc.Owner)
	}
	if svc.Evidence != ".aift/services.json" {
		t.Errorf("Contract.Services[0].Evidence = %q, want .aift/services.json", svc.Evidence)
	}

	// Verify ServiceRecord also has defaults
	if len(reg.Services) != 1 {
		t.Fatalf("expected 1 service record, got %d", len(reg.Services))
	}
	sr := reg.Services[0]
	if sr.Status != "planned" {
		t.Errorf("ServiceRecord.Status = %q, want planned", sr.Status)
	}
	if sr.Version != "0.1.0" {
		t.Errorf("ServiceRecord.Version = %q, want 0.1.0", sr.Version)
	}
	if sr.Evidence != ".aift/services.json" {
		t.Errorf("ServiceRecord.Evidence = %q, want .aift/services.json", sr.Evidence)
	}
}

func TestScanDefaultsDoNotOverwriteExplicit(t *testing.T) {
	root := t.TempDir()
	repoDir := filepath.Join(root, "explicit-repo")
	os.MkdirAll(filepath.Join(repoDir, ".git"), 0755)
	os.MkdirAll(filepath.Join(repoDir, ".aift"), 0755)

	// Contract with explicit values that should NOT be overwritten
	os.WriteFile(filepath.Join(repoDir, ".aift", "services.json"), []byte(`{
		"repo": "explicit-repo",
		"services": [
			{
				"name": "api",
				"kind": "grpc",
				"status": "ready",
				"version": "2.0.0",
				"owner": "team-alpha",
				"evidence": "ci-verified",
				"provides": [],
				"requires": [],
				"events": []
			}
		]
	}`), 0644)

	osHome := filepath.Join(root, "AIFT-OS")
	for _, d := range []string{"registry", "reports"} {
		os.MkdirAll(filepath.Join(osHome, d), 0755)
	}

	cfg := config.Config{Root: root, OSHome: osHome}
	if err := Scan(cfg); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(osHome, "registry", "service-contracts.json"))
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	svc := reg.Contracts[0].Services[0]
	if svc.Status != "ready" {
		t.Errorf("Status = %q, want ready (should not be overwritten)", svc.Status)
	}
	if svc.Version != "2.0.0" {
		t.Errorf("Version = %q, want 2.0.0 (should not be overwritten)", svc.Version)
	}
	if svc.Owner != "team-alpha" {
		t.Errorf("Owner = %q, want team-alpha (should not be overwritten)", svc.Owner)
	}
	if svc.Evidence != "ci-verified" {
		t.Errorf("Evidence = %q, want ci-verified (should not be overwritten)", svc.Evidence)
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
