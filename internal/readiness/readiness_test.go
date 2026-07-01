package readiness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestIsValidStatus(t *testing.T) {
	for _, s := range ValidStatuses {
		if !IsValid(s) {
			t.Errorf("IsValid(%q) = false, want true", s)
		}
	}
	if IsValid("unknown") {
		t.Error("IsValid(unknown) should be false")
	}
	if IsValid("") {
		t.Error("IsValid('') should be false")
	}
}

func TestTransitionSameStatus(t *testing.T) {
	for _, s := range ValidStatuses {
		if err := Transition(s, s); err != nil {
			t.Errorf("Transition(%s, %s) = %v, want nil", s, s, err)
		}
	}
}

func TestTransitionPlannedToDetected(t *testing.T) {
	if err := Transition(StatusPlanned, StatusDetected); err != nil {
		t.Errorf("planned -> detected should be valid: %v", err)
	}
}

func TestTransitionPlannedToReady(t *testing.T) {
	if err := Transition(StatusPlanned, StatusReady); err != nil {
		t.Errorf("planned -> ready should be valid: %v", err)
	}
}

func TestTransitionPlannedToRemoved(t *testing.T) {
	if err := Transition(StatusPlanned, StatusRemoved); err != nil {
		t.Errorf("planned -> removed should be valid: %v", err)
	}
}

func TestTransitionDetectedToReady(t *testing.T) {
	if err := Transition(StatusDetected, StatusReady); err != nil {
		t.Errorf("detected -> ready should be valid: %v", err)
	}
}

func TestTransitionDetectedToBlocked(t *testing.T) {
	if err := Transition(StatusDetected, StatusBlocked); err != nil {
		t.Errorf("detected -> blocked should be valid: %v", err)
	}
}

func TestTransitionReadyToActive(t *testing.T) {
	if err := Transition(StatusReady, StatusActive); err != nil {
		t.Errorf("ready -> active should be valid: %v", err)
	}
}

func TestTransitionReadyToBlocked(t *testing.T) {
	if err := Transition(StatusReady, StatusBlocked); err != nil {
		t.Errorf("ready -> blocked should be valid: %v", err)
	}
}

func TestTransitionActiveToBlocked(t *testing.T) {
	if err := Transition(StatusActive, StatusBlocked); err != nil {
		t.Errorf("active -> blocked should be valid: %v", err)
	}
}

func TestTransitionActiveToDeprecated(t *testing.T) {
	if err := Transition(StatusActive, StatusDeprecated); err != nil {
		t.Errorf("active -> deprecated should be valid: %v", err)
	}
}

func TestTransitionBlockedToReady(t *testing.T) {
	if err := Transition(StatusBlocked, StatusReady); err != nil {
		t.Errorf("blocked -> ready should be valid: %v", err)
	}
}

func TestTransitionDeprecatedToRemoved(t *testing.T) {
	if err := Transition(StatusDeprecated, StatusRemoved); err != nil {
		t.Errorf("deprecated -> removed should be valid: %v", err)
	}
}

func TestTransitionInvalidFromPlannedToActive(t *testing.T) {
	err := Transition(StatusPlanned, StatusActive)
	if err == nil {
		t.Error("planned -> active should be invalid")
	}
}

func TestTransitionInvalidFromRemovedToReady(t *testing.T) {
	err := Transition(StatusRemoved, StatusReady)
	if err == nil {
		t.Error("removed -> ready should be invalid (removed is terminal)")
	}
}

func TestTransitionInvalidFromActiveToPlanned(t *testing.T) {
	err := Transition(StatusActive, StatusPlanned)
	if err == nil {
		t.Error("active -> planned should be invalid")
	}
}

func TestTransitionInvalidStatuses(t *testing.T) {
	err := Transition("bogus", StatusReady)
	if err == nil {
		t.Error("invalid current status should error")
	}
	err = Transition(StatusReady, "bogus")
	if err == nil {
		t.Error("invalid target status should error")
	}
}

func TestMapStatus(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"ready", StatusReady},
		{"v1", StatusReady},
		{"detected", StatusDetected},
		{"planned", StatusPlanned},
		{"broken", StatusBlocked},
		{"active", StatusActive},
		{"deprecated", StatusDeprecated},
		{"removed", StatusRemoved},
		{"", StatusPlanned},
		{"something-else", StatusDetected},
	}

	for _, tc := range cases {
		got := mapStatus(tc.input)
		if got != tc.want {
			t.Errorf("mapStatus(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSummarize(t *testing.T) {
	objects := []Object{
		{Kind: "repository", Status: StatusReady},
		{Kind: "repository", Status: StatusDetected},
		{Kind: "command", Status: StatusReady},
		{Kind: "command", Status: StatusPlanned},
		{Kind: "service", Status: StatusActive},
	}

	s := summarize(objects)

	if s.Total != 5 {
		t.Errorf("Total = %d, want 5", s.Total)
	}
	if s.ReadyCount != 3 {
		t.Errorf("ReadyCount = %d, want 3 (2 ready + 1 active)", s.ReadyCount)
	}
	if s.ByStatus[StatusReady] != 2 {
		t.Errorf("ByStatus[ready] = %d, want 2", s.ByStatus[StatusReady])
	}
	if s.ByKind["repository"] != 2 {
		t.Errorf("ByKind[repository] = %d, want 2", s.ByKind["repository"])
	}
}

func TestSummarizeEmpty(t *testing.T) {
	s := summarize(nil)
	if s.Total != 0 {
		t.Errorf("Total = %d, want 0", s.Total)
	}
	if s.ReadyCount != 0 {
		t.Errorf("ReadyCount = %d, want 0", s.ReadyCount)
	}
}

func TestObjectJSON(t *testing.T) {
	obj := Object{
		ID:       "repo:test",
		Kind:     "repository",
		Name:     "test",
		Status:   StatusReady,
		Evidence: "go.mod found",
		Source:   "/path/to/test",
	}

	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Object
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.ID != "repo:test" {
		t.Errorf("ID = %q, want repo:test", decoded.ID)
	}
	if decoded.Status != StatusReady {
		t.Errorf("Status = %q, want ready", decoded.Status)
	}
}

func TestRegistryJSON(t *testing.T) {
	reg := Registry{
		GeneratedAt: "2024-01-01T00:00:00Z",
		Objects: []Object{
			{ID: "repo:test", Kind: "repository", Name: "test", Status: StatusReady},
		},
		Summary: Summary{Total: 1, ByStatus: map[string]int{"ready": 1}, ByKind: map[string]int{"repository": 1}, ReadyCount: 1},
	}

	data, err := json.Marshal(reg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Registry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(decoded.Objects) != 1 {
		t.Errorf("Objects len = %d, want 1", len(decoded.Objects))
	}
	if decoded.Summary.ReadyCount != 1 {
		t.Errorf("ReadyCount = %d, want 1", decoded.Summary.ReadyCount)
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]int{"c": 3, "a": 1, "b": 2}
	keys := sortedKeys(m)
	if len(keys) != 3 || keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("sortedKeys = %v, want [a b c]", keys)
	}
}

func TestSortedKeysEmpty(t *testing.T) {
	keys := sortedKeys(map[string]int{})
	if len(keys) != 0 {
		t.Errorf("sortedKeys(empty) = %v, want []", keys)
	}
}

func TestScanRepositories(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "root")
	os.MkdirAll(root, 0755)

	// Create a test repo
	repoDir := filepath.Join(root, "test-repo")
	os.MkdirAll(filepath.Join(repoDir, ".git"), 0755)

	cfg := makeTestConfig(dir, root)
	objects := scanRepositories(cfg)

	if len(objects) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(objects))
	}
	if objects[0].Kind != "repository" {
		t.Errorf("Kind = %q, want repository", objects[0].Kind)
	}
	if objects[0].Status != StatusDetected {
		t.Errorf("Status = %q, want detected", objects[0].Status)
	}
}

func TestScanRepositoriesWithManifest(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "root")
	os.MkdirAll(root, 0755)

	repoDir := filepath.Join(root, "test-repo")
	os.MkdirAll(filepath.Join(repoDir, ".git"), 0755)
	os.MkdirAll(filepath.Join(repoDir, ".aift"), 0755)
	os.WriteFile(filepath.Join(repoDir, ".aift", "repo.json"), []byte(`{}`), 0644)

	cfg := makeTestConfig(dir, root)
	objects := scanRepositories(cfg)

	if len(objects) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(objects))
	}
	if objects[0].Status != StatusReady {
		t.Errorf("Status = %q, want ready (has .aift/repo.json)", objects[0].Status)
	}
}

func TestScanModules(t *testing.T) {
	dir := t.TempDir()
	regDir := filepath.Join(dir, "registry")
	os.MkdirAll(regDir, 0755)

	reg := map[string]interface{}{
		"modules": []map[string]string{
			{"repo": "test-repo", "name": "test.module", "kind": "go-module", "status": "ready"},
			{"repo": "test-repo", "name": "other.module", "kind": "go-module", "status": "planned"},
		},
	}
	data, _ := json.Marshal(reg)
	os.WriteFile(filepath.Join(regDir, "modules.json"), data, 0644)

	cfg := makeTestConfig(dir, dir)
	objects := scanModules(cfg)

	if len(objects) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(objects))
	}
	if objects[0].Status != StatusReady {
		t.Errorf("first module status = %q, want ready", objects[0].Status)
	}
	if objects[1].Status != StatusPlanned {
		t.Errorf("second module status = %q, want planned", objects[1].Status)
	}
}

func TestScanModulesMissing(t *testing.T) {
	dir := t.TempDir()
	cfg := makeTestConfig(dir, dir)
	objects := scanModules(cfg)
	if objects != nil {
		t.Errorf("expected nil for missing modules.json, got %d objects", len(objects))
	}
}

func TestScanCapabilities(t *testing.T) {
	dir := t.TempDir()
	regDir := filepath.Join(dir, "registry")
	os.MkdirAll(regDir, 0755)

	reg := map[string]interface{}{
		"repos": []map[string]interface{}{
			{
				"repo": "test-repo",
				"capabilities": []map[string]string{
					{"name": "build", "status": "ready"},
					{"name": "deploy", "status": "broken"},
				},
			},
		},
	}
	data, _ := json.Marshal(reg)
	os.WriteFile(filepath.Join(regDir, "capabilities.json"), data, 0644)

	cfg := makeTestConfig(dir, dir)
	objects := scanCapabilities(cfg)

	if len(objects) != 2 {
		t.Fatalf("expected 2 capabilities, got %d", len(objects))
	}
	if objects[0].Status != StatusReady {
		t.Errorf("build status = %q, want ready", objects[0].Status)
	}
	if objects[1].Status != StatusBlocked {
		t.Errorf("deploy status = %q, want blocked (broken maps to blocked)", objects[1].Status)
	}
}

func TestScanServices(t *testing.T) {
	dir := t.TempDir()
	regDir := filepath.Join(dir, "registry")
	os.MkdirAll(regDir, 0755)

	reg := map[string]interface{}{
		"services": []map[string]string{
			{"repo": "r1", "name": "svc.r1", "status": "planned", "evidence": "test"},
		},
	}
	data, _ := json.Marshal(reg)
	os.WriteFile(filepath.Join(regDir, "service-contracts.json"), data, 0644)

	cfg := makeTestConfig(dir, dir)
	objects := scanServices(cfg)

	if len(objects) != 1 {
		t.Fatalf("expected 1 service, got %d", len(objects))
	}
	if objects[0].Status != StatusPlanned {
		t.Errorf("service status = %q, want planned", objects[0].Status)
	}
}

func TestScanEventsNoFile(t *testing.T) {
	dir := t.TempDir()
	cfg := makeTestConfig(dir, dir)
	objects := scanEvents(cfg)
	if objects != nil {
		t.Errorf("expected nil for missing event bus, got %d objects", len(objects))
	}
}

func TestScanEventsWithContent(t *testing.T) {
	dir := t.TempDir()
	evDir := filepath.Join(dir, "var", "events")
	os.MkdirAll(evDir, 0755)
	os.WriteFile(filepath.Join(evDir, "event-bus.jsonl"), []byte(`{"topic":"test"}`+"\n"), 0644)

	cfg := makeTestConfig(dir, dir)
	objects := scanEvents(cfg)

	if len(objects) != 1 {
		t.Fatalf("expected 1 event object, got %d", len(objects))
	}
	if objects[0].Status != StatusActive {
		t.Errorf("event bus status = %q, want active", objects[0].Status)
	}
}

func TestScanCommands(t *testing.T) {
	dir := t.TempDir()
	regDir := filepath.Join(dir, "registry")
	os.MkdirAll(regDir, 0755)

	arch := map[string]interface{}{
		"commands": []map[string]interface{}{
			{"name": "doctor", "has_handler": true, "has_help": true, "status": "active"},
			{"name": "intelligence", "has_handler": true, "has_help": true, "status": "planned"},
		},
	}
	data, _ := json.Marshal(arch)
	os.WriteFile(filepath.Join(regDir, "architecture.json"), data, 0644)

	cfg := makeTestConfig(dir, dir)
	objects := scanCommands(cfg)

	if len(objects) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(objects))
	}
	if objects[0].Status != StatusReady {
		t.Errorf("doctor status = %q, want ready", objects[0].Status)
	}
	if objects[1].Status != StatusPlanned {
		t.Errorf("intelligence status = %q, want planned", objects[1].Status)
	}
}

func TestScanScripts(t *testing.T) {
	dir := t.TempDir()
	scriptsDir := filepath.Join(dir, "scripts")
	os.MkdirAll(filepath.Join(scriptsDir, "lib"), 0755)

	os.WriteFile(filepath.Join(scriptsDir, "example.sh"), []byte("#!/bin/bash\nsource scripts/lib/aift-run.sh\n"), 0644)
	os.WriteFile(filepath.Join(scriptsDir, "legacy.sh"), []byte("#!/bin/bash\necho hello\n"), 0644)

	cfg := makeTestConfig(dir, dir)
	objects := scanScripts(cfg)

	if len(objects) != 2 {
		t.Fatalf("expected 2 scripts, got %d", len(objects))
	}

	byName := map[string]Object{}
	for _, o := range objects {
		byName[o.Name] = o
	}
	if byName["example.sh"].Status != StatusReady {
		t.Errorf("example.sh status = %q, want ready (sources harness)", byName["example.sh"].Status)
	}
	if byName["legacy.sh"].Status != StatusDetected {
		t.Errorf("legacy.sh status = %q, want detected", byName["legacy.sh"].Status)
	}
}

func TestScanScriptsNoDir(t *testing.T) {
	dir := t.TempDir()
	cfg := makeTestConfig(dir, dir)
	objects := scanScripts(cfg)
	if objects != nil {
		t.Errorf("expected nil for missing scripts dir, got %d", len(objects))
	}
}

func TestScanServicesConsumesPersistedDefaults(t *testing.T) {
	dir := t.TempDir()
	regDir := filepath.Join(dir, "registry")
	os.MkdirAll(regDir, 0755)

	// Simulates service-contracts.json produced by Scan() after PR #11 fix:
	// defaults (status, evidence) are present even for legacy services
	reg := map[string]interface{}{
		"services": []map[string]string{
			{"repo": "legacy-repo", "name": "legacy-svc", "status": "planned", "evidence": ".aift/services.json"},
			{"repo": "other-repo", "name": "other-svc", "status": "ready", "evidence": "ci-verified"},
		},
	}
	data, _ := json.Marshal(reg)
	os.WriteFile(filepath.Join(regDir, "service-contracts.json"), data, 0644)

	cfg := makeTestConfig(dir, dir)
	objects := scanServices(cfg)

	if len(objects) != 2 {
		t.Fatalf("expected 2 services, got %d", len(objects))
	}

	byName := map[string]Object{}
	for _, o := range objects {
		byName[o.Name] = o
	}

	// Verify readiness correctly maps the persisted defaults
	legacy := byName["legacy-svc"]
	if legacy.Status != StatusPlanned {
		t.Errorf("legacy-svc status = %q, want planned", legacy.Status)
	}
	if legacy.Evidence != ".aift/services.json" {
		t.Errorf("legacy-svc evidence = %q, want .aift/services.json", legacy.Evidence)
	}

	other := byName["other-svc"]
	if other.Status != StatusReady {
		t.Errorf("other-svc status = %q, want ready", other.Status)
	}
	if other.Evidence != "ci-verified" {
		t.Errorf("other-svc evidence = %q, want ci-verified", other.Evidence)
	}
}

func TestScanServicesEmptyEvidence(t *testing.T) {
	dir := t.TempDir()
	regDir := filepath.Join(dir, "registry")
	os.MkdirAll(regDir, 0755)

	// Service with empty evidence should get a fallback
	reg := map[string]interface{}{
		"services": []map[string]string{
			{"repo": "r1", "name": "svc1", "status": "planned", "evidence": ""},
		},
	}
	data, _ := json.Marshal(reg)
	os.WriteFile(filepath.Join(regDir, "service-contracts.json"), data, 0644)

	cfg := makeTestConfig(dir, dir)
	objects := scanServices(cfg)

	if len(objects) != 1 {
		t.Fatalf("expected 1 service, got %d", len(objects))
	}
	if objects[0].Evidence != "registry/service-contracts.json" {
		t.Errorf("evidence = %q, want fallback registry/service-contracts.json", objects[0].Evidence)
	}
}

func makeTestConfig(osHome, root string) config.Config {
	return config.Config{
		Root:   root,
		OSHome: osHome,
	}
}
