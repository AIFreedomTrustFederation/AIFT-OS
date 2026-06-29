package schedulerplan

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

func TestObjectToJobReady(t *testing.T) {
	obj := readinessObject{
		ID: "repo:test", Kind: "repository", Name: "test",
		Status: "ready", Evidence: "go.mod found", Source: "/path",
	}
	job := objectToJob(obj, nil)
	if job.Status != StatusRunnable {
		t.Errorf("Status = %q, want runnable", job.Status)
	}
	if job.ReadinessRef != "repo:test" {
		t.Errorf("ReadinessRef = %q, want repo:test", job.ReadinessRef)
	}
}

func TestObjectToJobActive(t *testing.T) {
	obj := readinessObject{
		ID: "event:bus", Kind: "event", Name: "event-bus",
		Status: "active", Source: "var/events",
	}
	job := objectToJob(obj, nil)
	if job.Status != StatusRunnable {
		t.Errorf("Status = %q, want runnable (active maps to runnable)", job.Status)
	}
}

func TestObjectToJobPlanned(t *testing.T) {
	obj := readinessObject{
		ID: "service:svc", Kind: "service", Name: "svc",
		Status: "planned", Source: "repo",
	}
	job := objectToJob(obj, nil)
	if job.Status != StatusPending {
		t.Errorf("Status = %q, want pending", job.Status)
	}
}

func TestObjectToJobBlocked(t *testing.T) {
	obj := readinessObject{
		ID: "capability:build", Kind: "capability", Name: "build",
		Status: "blocked", Source: "repo",
	}
	job := objectToJob(obj, nil)
	if job.Status != StatusBlocked {
		t.Errorf("Status = %q, want blocked", job.Status)
	}
	if len(job.BlockedBy) == 0 {
		t.Error("BlockedBy should not be empty for blocked job")
	}
}

func TestObjectToJobDeprecated(t *testing.T) {
	obj := readinessObject{
		ID: "module:old", Kind: "module", Name: "old",
		Status: "deprecated", Source: "repo",
	}
	job := objectToJob(obj, nil)
	if job.Status != StatusSkipped {
		t.Errorf("Status = %q, want skipped", job.Status)
	}
}

func TestObjectToJobRemoved(t *testing.T) {
	obj := readinessObject{
		ID: "module:gone", Kind: "module", Name: "gone",
		Status: "removed", Source: "repo",
	}
	job := objectToJob(obj, nil)
	if job.Status != StatusSkipped {
		t.Errorf("Status = %q, want skipped", job.Status)
	}
}

func TestObjectToJobCommandWithHandler(t *testing.T) {
	obj := readinessObject{
		ID: "command:doctor", Kind: "command", Name: "doctor",
		Status: "ready", Source: "cmd/aift",
	}
	cmds := map[string]archCommand{
		"doctor": {Name: "doctor", HasHandler: true, HasHelp: true, Status: "active"},
	}
	job := objectToJob(obj, cmds)
	if job.Command != "aift doctor" {
		t.Errorf("Command = %q, want 'aift doctor'", job.Command)
	}
	if job.Status != StatusRunnable {
		t.Errorf("Status = %q, want runnable", job.Status)
	}
}

func TestObjectToJobPlannedCommand(t *testing.T) {
	obj := readinessObject{
		ID: "command:intelligence", Kind: "command", Name: "intelligence",
		Status: "planned", Source: "cmd/aift",
	}
	cmds := map[string]archCommand{
		"intelligence": {Name: "intelligence", HasHandler: true, HasHelp: true, Status: "planned"},
	}
	job := objectToJob(obj, cmds)
	if job.Status != StatusPending {
		t.Errorf("Status = %q, want pending (planned command)", job.Status)
	}
	if job.Command != "" {
		t.Errorf("Command = %q, want empty (planned commands have no executable)", job.Command)
	}
}

func TestObjectToJobMissingCommand(t *testing.T) {
	obj := readinessObject{
		ID: "command:unknown", Kind: "command", Name: "unknown",
		Status: "ready", Source: "cmd/aift",
	}
	// Empty command index — command not found in architecture
	job := objectToJob(obj, map[string]archCommand{})
	if job.Status != StatusRunnable {
		t.Errorf("Status = %q, want runnable", job.Status)
	}
	if job.Command != "" {
		t.Errorf("Command = %q, want empty (not in architecture)", job.Command)
	}
}

func TestObjectToJobScript(t *testing.T) {
	obj := readinessObject{
		ID: "script:coverage.sh", Kind: "script", Name: "coverage.sh",
		Status: "ready", Source: "scripts/coverage.sh",
	}
	job := objectToJob(obj, nil)
	if job.Script != "scripts/coverage.sh" {
		t.Errorf("Script = %q, want scripts/coverage.sh", job.Script)
	}
}

func TestBuildJobsSortsOutput(t *testing.T) {
	rr := readinessRegistry{
		Objects: []readinessObject{
			{ID: "script:z.sh", Kind: "script", Name: "z.sh", Status: "ready"},
			{ID: "command:a", Kind: "command", Name: "a", Status: "ready"},
			{ID: "repo:m", Kind: "repository", Name: "m", Status: "ready"},
		},
	}
	jobs := buildJobs(rr, archRegistry{})

	if len(jobs) != 3 {
		t.Fatalf("expected 3 jobs, got %d", len(jobs))
	}
	// Should be sorted by kind then ID
	if jobs[0].Kind != "command" {
		t.Errorf("first job kind = %q, want command", jobs[0].Kind)
	}
	if jobs[1].Kind != "repository" {
		t.Errorf("second job kind = %q, want repository", jobs[1].Kind)
	}
	if jobs[2].Kind != "script" {
		t.Errorf("third job kind = %q, want script", jobs[2].Kind)
	}
}

func TestResolveDependenciesBlocksCommandsWithoutRepos(t *testing.T) {
	jobs := []Job{
		{ID: "job:repo:test", Kind: "repository", Name: "test", Status: StatusPending, ReadinessRef: "repo:test"},
		{ID: "job:command:doctor", Kind: "command", Name: "doctor", Status: StatusRunnable, ReadinessRef: "command:doctor"},
	}

	resolveDependencies(jobs)

	// Command should be blocked because no repos are runnable
	if jobs[1].Status != StatusBlocked {
		t.Errorf("command status = %q, want blocked (no runnable repos)", jobs[1].Status)
	}
	if len(jobs[1].BlockedBy) == 0 {
		t.Error("BlockedBy should list reason")
	}
}

func TestResolveDependenciesAllowsCommandsWithRunnableRepo(t *testing.T) {
	jobs := []Job{
		{ID: "job:repo:test", Kind: "repository", Name: "test", Status: StatusRunnable, ReadinessRef: "repo:test"},
		{ID: "job:command:doctor", Kind: "command", Name: "doctor", Status: StatusRunnable, ReadinessRef: "command:doctor"},
	}

	resolveDependencies(jobs)

	if jobs[1].Status != StatusRunnable {
		t.Errorf("command status = %q, want runnable (repo is ready)", jobs[1].Status)
	}
}

func TestResolveDependenciesBlocksCapabilityWithoutRepo(t *testing.T) {
	jobs := []Job{
		{ID: "job:repo:my-repo", Kind: "repository", Name: "my-repo", Status: StatusPending, ReadinessRef: "repo:my-repo", Source: "my-repo"},
		{ID: "job:capability:build", Kind: "capability", Name: "build", Status: StatusRunnable, ReadinessRef: "capability:build", Source: "my-repo"},
	}

	resolveDependencies(jobs)

	if jobs[1].Status != StatusBlocked {
		t.Errorf("capability status = %q, want blocked (source repo not ready)", jobs[1].Status)
	}
	if len(jobs[1].DependsOn) == 0 || jobs[1].DependsOn[0] != "repo:my-repo" {
		t.Errorf("DependsOn = %v, want [repo:my-repo]", jobs[1].DependsOn)
	}
}

func TestResolveDependenciesBlocksServiceWithoutRepo(t *testing.T) {
	jobs := []Job{
		{ID: "job:repo:svc-repo", Kind: "repository", Name: "svc-repo", Status: StatusPending, ReadinessRef: "repo:svc-repo", Source: "svc-repo"},
		{ID: "job:service:api", Kind: "service", Name: "api", Status: StatusRunnable, ReadinessRef: "service:api", Source: "svc-repo"},
	}

	resolveDependencies(jobs)

	if jobs[1].Status != StatusBlocked {
		t.Errorf("service status = %q, want blocked", jobs[1].Status)
	}
}

func TestResolveDependenciesBlocksModuleWithoutRepo(t *testing.T) {
	jobs := []Job{
		{ID: "job:repo:mod-repo", Kind: "repository", Name: "mod-repo", Status: StatusPending, ReadinessRef: "repo:mod-repo", Source: "mod-repo"},
		{ID: "job:module:core", Kind: "module", Name: "core", Status: StatusRunnable, ReadinessRef: "module:core", Source: "mod-repo"},
	}

	resolveDependencies(jobs)

	if jobs[1].Status != StatusBlocked {
		t.Errorf("module status = %q, want blocked", jobs[1].Status)
	}
}

func TestResolveDependenciesAllowsCapabilityWithRunnableRepo(t *testing.T) {
	jobs := []Job{
		{ID: "job:repo:my-repo", Kind: "repository", Name: "my-repo", Status: StatusRunnable, ReadinessRef: "repo:my-repo", Source: "my-repo"},
		{ID: "job:capability:build", Kind: "capability", Name: "build", Status: StatusRunnable, ReadinessRef: "capability:build", Source: "my-repo"},
	}

	resolveDependencies(jobs)

	if jobs[1].Status != StatusRunnable {
		t.Errorf("capability status = %q, want runnable", jobs[1].Status)
	}
}

func TestSummarize(t *testing.T) {
	jobs := []Job{
		{Kind: "repository", Status: StatusRunnable},
		{Kind: "repository", Status: StatusPending},
		{Kind: "command", Status: StatusRunnable},
		{Kind: "command", Status: StatusBlocked},
		{Kind: "service", Status: StatusSkipped},
	}

	s := summarize(jobs)

	if s.Total != 5 {
		t.Errorf("Total = %d, want 5", s.Total)
	}
	if s.RunnableCount != 2 {
		t.Errorf("RunnableCount = %d, want 2", s.RunnableCount)
	}
	if s.BlockedCount != 1 {
		t.Errorf("BlockedCount = %d, want 1", s.BlockedCount)
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
}

func TestJobJSON(t *testing.T) {
	job := Job{
		ID:        "job:repo:test",
		Name:      "test",
		Kind:      "repository",
		Status:    StatusRunnable,
		Command:   "aift verify",
		DependsOn: []string{"repo:other"},
		Evidence:  "readiness: ready",
		Source:    "/path/to/test",
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded Job
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.ID != "job:repo:test" {
		t.Errorf("ID = %q", decoded.ID)
	}
	if decoded.Command != "aift verify" {
		t.Errorf("Command = %q", decoded.Command)
	}
}

func TestPlanJSON(t *testing.T) {
	plan := Plan{
		GeneratedAt: "2024-01-01T00:00:00Z",
		Jobs:        []Job{{ID: "job:test", Name: "test", Kind: "repository", Status: StatusRunnable}},
		Summary:     Summary{Total: 1, ByStatus: map[string]int{"runnable": 1}, ByKind: map[string]int{"repository": 1}, RunnableCount: 1},
	}

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded Plan
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Jobs) != 1 {
		t.Errorf("Jobs len = %d, want 1", len(decoded.Jobs))
	}
}

func TestBuildPlanFromReadiness(t *testing.T) {
	dir := t.TempDir()
	regDir := filepath.Join(dir, "registry")
	os.MkdirAll(regDir, 0755)
	os.MkdirAll(filepath.Join(dir, "reports"), 0755)
	os.MkdirAll(filepath.Join(dir, "var", "events"), 0755)

	// Write readiness registry
	rr := map[string]interface{}{
		"objects": []map[string]string{
			{"id": "repo:test", "kind": "repository", "name": "test", "status": "ready", "evidence": "go.mod", "source": "/test"},
			{"id": "command:doctor", "kind": "command", "name": "doctor", "status": "ready", "evidence": "handler", "source": "cmd/aift"},
			{"id": "service:api", "kind": "service", "name": "api", "status": "planned", "evidence": "contract", "source": "test"},
		},
	}
	data, _ := json.Marshal(rr)
	os.WriteFile(filepath.Join(regDir, "runtime-readiness.json"), data, 0644)

	// Write architecture registry
	ar := map[string]interface{}{
		"commands": []map[string]interface{}{
			{"name": "doctor", "has_handler": true, "has_help": true, "status": "active"},
		},
	}
	archData, _ := json.Marshal(ar)
	os.WriteFile(filepath.Join(regDir, "architecture.json"), archData, 0644)

	cfg := config.Config{Root: dir, OSHome: dir}
	plan, err := BuildPlan(cfg)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}

	if plan.Summary.Total != 3 {
		t.Errorf("Total = %d, want 3", plan.Summary.Total)
	}

	byName := map[string]Job{}
	for _, j := range plan.Jobs {
		byName[j.Name] = j
	}

	if byName["test"].Status != StatusRunnable {
		t.Errorf("test repo status = %q, want runnable", byName["test"].Status)
	}
	if byName["doctor"].Status != StatusRunnable {
		t.Errorf("doctor status = %q, want runnable", byName["doctor"].Status)
	}
	if byName["doctor"].Command != "aift doctor" {
		t.Errorf("doctor command = %q, want 'aift doctor'", byName["doctor"].Command)
	}
	if byName["api"].Status != StatusPending {
		t.Errorf("api status = %q, want pending", byName["api"].Status)
	}
}

func TestBuildPlanMissingReadiness(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}
	_, err := BuildPlan(cfg)
	if err == nil {
		t.Error("BuildPlan should fail when readiness registry is missing")
	}
}

func TestBuildPlanMissingArchitecture(t *testing.T) {
	dir := t.TempDir()
	regDir := filepath.Join(dir, "registry")
	os.MkdirAll(regDir, 0755)

	// Readiness only, no architecture
	rr := map[string]interface{}{
		"objects": []map[string]string{
			{"id": "repo:test", "kind": "repository", "name": "test", "status": "ready", "evidence": "go.mod", "source": "/test"},
		},
	}
	data, _ := json.Marshal(rr)
	os.WriteFile(filepath.Join(regDir, "runtime-readiness.json"), data, 0644)

	cfg := config.Config{Root: dir, OSHome: dir}
	plan, err := BuildPlan(cfg)
	if err != nil {
		t.Fatalf("BuildPlan should succeed without architecture: %v", err)
	}
	if plan.Summary.Total != 1 {
		t.Errorf("Total = %d, want 1", plan.Summary.Total)
	}
}

func TestGeneratePlanWritesFiles(t *testing.T) {
	dir := t.TempDir()
	regDir := filepath.Join(dir, "registry")
	os.MkdirAll(regDir, 0755)
	os.MkdirAll(filepath.Join(dir, "reports"), 0755)
	os.MkdirAll(filepath.Join(dir, "var", "events"), 0755)

	rr := map[string]interface{}{
		"objects": []map[string]string{
			{"id": "repo:test", "kind": "repository", "name": "test", "status": "ready", "evidence": "go.mod", "source": "/test"},
		},
	}
	data, _ := json.Marshal(rr)
	os.WriteFile(filepath.Join(regDir, "runtime-readiness.json"), data, 0644)

	cfg := config.Config{Root: dir, OSHome: dir}
	if err := GeneratePlan(cfg); err != nil {
		t.Fatalf("GeneratePlan: %v", err)
	}

	// Verify registry file written
	planJSON := filepath.Join(regDir, "scheduler-plan.json")
	if _, err := os.Stat(planJSON); err != nil {
		t.Errorf("scheduler-plan.json not created: %v", err)
	}

	// Verify report written
	planMD := filepath.Join(dir, "reports", "scheduler-plan.md")
	if _, err := os.Stat(planMD); err != nil {
		t.Errorf("scheduler-plan.md not created: %v", err)
	}

	// Verify JSON is valid
	planData, _ := os.ReadFile(planJSON)
	var plan Plan
	if err := json.Unmarshal(planData, &plan); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if plan.Summary.Total != 1 {
		t.Errorf("Total = %d, want 1", plan.Summary.Total)
	}
}

func TestReportContainsExpectedSections(t *testing.T) {
	dir := t.TempDir()
	regDir := filepath.Join(dir, "registry")
	os.MkdirAll(regDir, 0755)
	os.MkdirAll(filepath.Join(dir, "reports"), 0755)
	os.MkdirAll(filepath.Join(dir, "var", "events"), 0755)

	rr := map[string]interface{}{
		"objects": []map[string]string{
			{"id": "repo:test", "kind": "repository", "name": "test", "status": "ready", "evidence": "go.mod", "source": "/test"},
			{"id": "service:api", "kind": "service", "name": "api", "status": "blocked", "evidence": "missing dep", "source": "test"},
		},
	}
	data, _ := json.Marshal(rr)
	os.WriteFile(filepath.Join(regDir, "runtime-readiness.json"), data, 0644)

	cfg := config.Config{Root: dir, OSHome: dir}
	if err := GeneratePlan(cfg); err != nil {
		t.Fatalf("GeneratePlan: %v", err)
	}

	report, _ := os.ReadFile(filepath.Join(dir, "reports", "scheduler-plan.md"))
	content := string(report)

	for _, section := range []string{"# Scheduler Plan", "## Summary", "## Runnable Jobs", "## Blocked Jobs", "## All Jobs"} {
		if !contains(content, section) {
			t.Errorf("report missing section %q", section)
		}
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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
