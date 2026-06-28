package planner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasCapability(t *testing.T) {
	rp := RepoPlan{
		Capabilities: []Capability{
			{Name: "build", Status: "ready"},
			{Name: "test", Status: "v1"},
			{Name: "deploy", Status: "planned"},
			{Name: "sync", Status: "broken"},
		},
	}

	if !hasCapability(rp, "build") {
		t.Error("should find ready capability 'build'")
	}
	if !hasCapability(rp, "test") {
		t.Error("should find v1 capability 'test'")
	}
	if hasCapability(rp, "deploy") {
		t.Error("should not find planned capability 'deploy'")
	}
	if hasCapability(rp, "sync") {
		t.Error("should not find broken capability 'sync'")
	}
	if hasCapability(rp, "nonexistent") {
		t.Error("should not find nonexistent capability")
	}
}

func TestRecommendations(t *testing.T) {
	rp := RepoPlan{
		Detected: []string{"build", "test"},
		Planned:  []string{"start"},
	}
	recs := recommendations(rp)

	expectKeywords := []string{
		"verify",
		"build",
		"test",
		"start",
		"services.json",
	}
	for _, kw := range expectKeywords {
		found := false
		for _, rec := range recs {
			if strContains(rec, kw) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("recommendations should mention %q", kw)
		}
	}
}

func TestRecommendationsVerifyReady(t *testing.T) {
	rp := RepoPlan{
		Ready: []string{"verify"},
	}
	recs := recommendations(rp)
	for _, rec := range recs {
		if strContains(rec, "verify") {
			t.Error("should not recommend verify when already ready")
		}
	}
}

func TestRecommendationsBroken(t *testing.T) {
	rp := RepoPlan{
		Broken: []string{"build"},
	}
	recs := recommendations(rp)
	found := false
	for _, rec := range recs {
		if strContains(rec, "broken") || strContains(rec, "Fix") {
			found = true
			break
		}
	}
	if !found {
		t.Error("should recommend fixing broken capabilities")
	}
}

func TestRecommendationsNoBlocers(t *testing.T) {
	rp := RepoPlan{
		Ready:    []string{"verify", "build", "test"},
		Services: []Service{{Name: "svc"}},
	}
	recs := recommendations(rp)
	found := false
	for _, rec := range recs {
		if strContains(rec, "No immediate planner blockers") {
			found = true
			break
		}
	}
	if !found {
		t.Error("should report no blockers when all is good")
	}
}

func TestEvaluateRepoStateReady(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "capabilities.json"), []byte(`{
		"capabilities": [{"name": "verify", "status": "ready"}]
	}`), 0644)

	rp := evaluateRepo("test-repo", dir)

	if rp.State != "READY" {
		t.Errorf("state = %q, want READY", rp.State)
	}
	if len(rp.Ready) != 1 || rp.Ready[0] != "verify" {
		t.Errorf("ready = %v, want [verify]", rp.Ready)
	}
}

func TestEvaluateRepoStateBroken(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "capabilities.json"), []byte(`{
		"capabilities": [
			{"name": "verify", "status": "ready"},
			{"name": "build", "status": "broken"}
		]
	}`), 0644)

	rp := evaluateRepo("test-repo", dir)

	if rp.State != "BROKEN" {
		t.Errorf("state = %q, want BROKEN", rp.State)
	}
}

func TestEvaluateRepoStatePlanned(t *testing.T) {
	dir := t.TempDir()

	rp := evaluateRepo("test-repo", dir)

	if rp.State != "PLANNED" {
		t.Errorf("state = %q, want PLANNED", rp.State)
	}
}

func TestEvaluateRepoStateDetected(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "capabilities.json"), []byte(`{
		"capabilities": [{"name": "build", "status": "detected"}]
	}`), 0644)

	rp := evaluateRepo("test-repo", dir)

	if rp.State != "DETECTED" {
		t.Errorf("state = %q, want DETECTED", rp.State)
	}
}

func TestEvaluateRepoStateBlocked(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "capabilities.json"), []byte(`{
		"capabilities": [{"name": "verify", "status": "ready"}]
	}`), 0644)
	os.WriteFile(filepath.Join(aiftDir, "services.json"), []byte(`{
		"services": [{
			"name": "my-svc",
			"kind": "api",
			"status": "ready",
			"version": "1.0",
			"provides": [],
			"requires": ["nonexistent-cap"],
			"events": [],
			"evidence": "test"
		}]
	}`), 0644)

	rp := evaluateRepo("test-repo", dir)

	if rp.State != "BLOCKED" {
		t.Errorf("state = %q, want BLOCKED", rp.State)
	}
}

func TestLoadCapabilities(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "capabilities.json"), []byte(`{
		"capabilities": [
			{"name": "build", "status": "ready", "command": "go build"},
			{"name": "test", "status": "v1"}
		]
	}`), 0644)

	caps := loadCapabilities(dir)
	if len(caps) != 2 {
		t.Fatalf("expected 2 capabilities, got %d", len(caps))
	}
	if caps[0].Name != "build" || caps[0].Status != "ready" {
		t.Errorf("cap[0] = %+v, want build/ready", caps[0])
	}
}

func TestLoadCapabilitiesMissing(t *testing.T) {
	dir := t.TempDir()
	caps := loadCapabilities(dir)
	if len(caps) != 0 {
		t.Errorf("expected empty capabilities for missing file, got %d", len(caps))
	}
}

func TestLoadCapabilitiesInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "capabilities.json"), []byte(`invalid json`), 0644)

	caps := loadCapabilities(dir)
	if len(caps) != 0 {
		t.Errorf("expected empty capabilities for invalid JSON, got %d", len(caps))
	}
}

func TestLoadServices(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "services.json"), []byte(`{
		"services": [{
			"name": "api-svc",
			"kind": "http",
			"status": "ready",
			"version": "1.0",
			"provides": ["api"],
			"requires": ["db"],
			"events": [],
			"evidence": "test"
		}]
	}`), 0644)

	services := loadServices(dir)
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
	if services[0].Name != "api-svc" {
		t.Errorf("service name = %q, want api-svc", services[0].Name)
	}
}

func TestLoadServicesMissing(t *testing.T) {
	dir := t.TempDir()
	services := loadServices(dir)
	if len(services) != 0 {
		t.Errorf("expected empty services for missing file, got %d", len(services))
	}
}

func strContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
