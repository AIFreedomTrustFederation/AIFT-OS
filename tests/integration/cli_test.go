package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binary is the path to the compiled aiftd binary, built once in TestMain.
var binary string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "aift-cli-test-*")
	if err != nil {
		panic(err)
	}
	binary = filepath.Join(tmp, "aiftd")

	cmd := exec.Command("go", "build", "-o", binary, "./cmd/aift")
	cmd.Dir = repoRoot()
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("build failed: " + string(out))
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

func repoRoot() string {
	// Walk up from this test file to find the repo root.
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("cannot find repo root")
		}
		dir = parent
	}
}

// setupWorkspace creates a minimal AIFT workspace with a fake repo containing
// .git and go.mod so commands that call workspace.FindRepos can operate.
func setupWorkspace(t *testing.T) (root string) {
	t.Helper()
	root = t.TempDir()
	repoDir := filepath.Join(root, "test-repo")
	os.MkdirAll(filepath.Join(repoDir, ".git"), 0755)
	os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte("module test-repo\n\ngo 1.22\n"), 0644)
	os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("# Test\n"), 0644)

	// Create required AIFT-OS directory structure for doctor
	osHome := filepath.Join(root, "AIFT-OS")
	for _, d := range []string{
		"cmd/aift", "internal/config", "internal/workspace", "internal/gitx",
		"internal/doctor", "internal/registry", "internal/manifests",
		"internal/reports", "internal/plugins", "internal/sync",
		"internal/kernel", "install", "tests", "docs", "schemas",
		"registry", "reports", "bin",
	} {
		os.MkdirAll(filepath.Join(osHome, d), 0755)
	}

	return root
}

// run executes the aiftd binary with the given args and workspace env vars.
func run(t *testing.T, root string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(),
		"AIFT_ROOT="+root,
		"AIFT_OS_HOME="+filepath.Join(root, "AIFT-OS"),
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// ── help / version ───────────────────────────────────────────────────

func TestHelp(t *testing.T) {
	out, err := run(t, t.TempDir(), "help")
	if err != nil {
		t.Fatalf("help failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Commands:") {
		t.Error("help output should contain 'Commands:'")
	}
}

func TestHelpFlag(t *testing.T) {
	out, _ := run(t, t.TempDir(), "--help")
	if !strings.Contains(out, "Commands:") {
		t.Error("--help should show commands")
	}
}

func TestVersion(t *testing.T) {
	out, err := run(t, t.TempDir(), "version")
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(out, "AIFT-OS") {
		t.Error("version should contain AIFT-OS")
	}
}

// ── doctor ───────────────────────────────────────────────────────────

func TestDoctor(t *testing.T) {
	root := setupWorkspace(t)
	out, err := run(t, root, "doctor")
	if err != nil {
		t.Fatalf("doctor failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "OK") {
		t.Error("doctor should report OK")
	}
}

// ── status ───────────────────────────────────────────────────────────

func TestStatus(t *testing.T) {
	root := setupWorkspace(t)
	out, err := run(t, root, "status")
	if err != nil {
		t.Fatalf("status failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "REPOSITORY") {
		t.Error("status should print table header")
	}
	if !strings.Contains(out, "test-repo") {
		t.Error("status should list test-repo")
	}
}

// ── manifest ─────────────────────────────────────────────────────────

func TestManifest(t *testing.T) {
	root := setupWorkspace(t)
	out, err := run(t, root, "manifest")
	if err != nil {
		t.Fatalf("manifest failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "OK") {
		t.Error("manifest should report OK")
	}

	// Verify manifest was created
	manifest := filepath.Join(root, "test-repo", ".aift", "repo.json")
	data, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("manifest file not created: %v", err)
	}
	var m map[string]interface{}
	if json.Unmarshal(data, &m) != nil {
		t.Error("manifest should be valid JSON")
	}
}

// ── registry ─────────────────────────────────────────────────────────

func TestRegistry(t *testing.T) {
	root := setupWorkspace(t)
	_, err := run(t, root, "registry")
	if err != nil {
		t.Fatalf("registry failed: %v", err)
	}

	registryFile := filepath.Join(root, "AIFT-OS", "registry", "repos.json")
	data, err := os.ReadFile(registryFile)
	if err != nil {
		t.Fatalf("registry file not created: %v", err)
	}
	var records []interface{}
	if json.Unmarshal(data, &records) != nil {
		t.Error("registry should be valid JSON array")
	}
}

// ── events ───────────────────────────────────────────────────────────

func TestEvents(t *testing.T) {
	root := setupWorkspace(t)
	out, err := run(t, root, "events")
	if err != nil {
		t.Fatalf("events failed: %v\n%s", err, out)
	}
	// With no events file, should say "No events yet"
	if !strings.Contains(out, "No events") {
		t.Error("events with empty log should say 'No events'")
	}
}

// ── event-bus ────────────────────────────────────────────────────────

func TestEventBusPublishAndList(t *testing.T) {
	root := setupWorkspace(t)

	// Publish an event
	out, err := run(t, root, "event-bus", "publish", "test.topic", "hello world")
	if err != nil {
		t.Fatalf("event-bus publish failed: %v\n%s", err, out)
	}

	// List events
	out, err = run(t, root, "event-bus", "list")
	if err != nil {
		t.Fatalf("event-bus list failed: %v\n%s", err, out)
	}
}

func TestEventBusReport(t *testing.T) {
	root := setupWorkspace(t)

	// Publish so there's data
	run(t, root, "event-bus", "publish", "test.topic", "msg")

	out, err := run(t, root, "event-bus", "report")
	if err != nil {
		t.Fatalf("event-bus report failed: %v\n%s", err, out)
	}
}

// ── capabilities ─────────────────────────────────────────────────────

func TestCapabilitiesScan(t *testing.T) {
	root := setupWorkspace(t)
	_, err := run(t, root, "capabilities", "scan")
	if err != nil {
		t.Fatalf("capabilities scan failed: %v", err)
	}
}

func TestCapabilitiesReport(t *testing.T) {
	root := setupWorkspace(t)
	// Scan first to generate data
	run(t, root, "capabilities", "scan")

	out, err := run(t, root, "capabilities", "report")
	if err != nil {
		t.Fatalf("capabilities report failed: %v\n%s", err, out)
	}
}

// ── modules ──────────────────────────────────────────────────────────

func TestModulesScan(t *testing.T) {
	root := setupWorkspace(t)
	_, err := run(t, root, "modules", "scan")
	if err != nil {
		t.Fatalf("modules scan failed: %v", err)
	}
}

func TestModulesList(t *testing.T) {
	root := setupWorkspace(t)
	out, err := run(t, root, "modules", "list")
	if err != nil {
		t.Fatalf("modules list failed: %v\n%s", err, out)
	}
}

func TestModulesInitAll(t *testing.T) {
	root := setupWorkspace(t)
	_, err := run(t, root, "modules", "init-all")
	if err != nil {
		t.Fatalf("modules init-all failed: %v", err)
	}
}

// ── graph ────────────────────────────────────────────────────────────

func TestGraph(t *testing.T) {
	root := setupWorkspace(t)
	out, err := run(t, root, "graph")
	if err != nil {
		t.Fatalf("graph failed: %v\n%s", err, out)
	}
}

// ── verify ───────────────────────────────────────────────────────────

func TestVerify(t *testing.T) {
	root := setupWorkspace(t)
	out, err := run(t, root, "verify")
	if err != nil {
		t.Fatalf("verify failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "OK") {
		t.Error("verify should report OK on success")
	}
}

// ── planned commands ─────────────────────────────────────────────────

func TestPlannedCommands(t *testing.T) {
	planned := []string{
		"intelligence",
		"manual",
		"mesh",
		"service-contracts",
		"plan",
	}

	root := setupWorkspace(t)
	for _, cmd := range planned {
		out, err := run(t, root, cmd)
		if err == nil {
			t.Errorf("%s should return error (planned)", cmd)
			continue
		}
		if !strings.Contains(out, "planned") {
			t.Errorf("%s error should mention 'planned': %s", cmd, out)
		}
	}
}

// ── unknown command ──────────────────────────────────────────────────

func TestUnknownCommand(t *testing.T) {
	_, err := run(t, t.TempDir(), "nonexistent-command")
	if err == nil {
		t.Error("unknown command should return error")
	}
}

// ── no panics on empty args ──────────────────────────────────────────

func TestNoArgsPanic(t *testing.T) {
	root := setupWorkspace(t)
	// Running with no args should show help, not panic
	out, err := run(t, root)
	if err != nil {
		t.Fatalf("no args should not fail: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Commands:") {
		t.Error("no args should show help")
	}
}

// ── runtime ──────────────────────────────────────────────────────────

func TestRuntimeScan(t *testing.T) {
	root := setupWorkspace(t)
	out, err := run(t, root, "runtime", "scan")
	if err != nil {
		t.Fatalf("runtime scan failed: %v\n%s", err, out)
	}

	// Check registry file was created
	regPath := filepath.Join(root, "AIFT-OS", "registry", "runtime-readiness.json")
	data, err := os.ReadFile(regPath)
	if err != nil {
		t.Fatalf("registry/runtime-readiness.json not created: %v", err)
	}
	var reg map[string]interface{}
	if err := json.Unmarshal(data, &reg); err != nil {
		t.Fatalf("invalid JSON in runtime-readiness.json: %v", err)
	}
	if _, ok := reg["objects"]; !ok {
		t.Error("runtime-readiness.json missing 'objects' field")
	}
	if _, ok := reg["summary"]; !ok {
		t.Error("runtime-readiness.json missing 'summary' field")
	}
}

func TestRuntimeStatus(t *testing.T) {
	root := setupWorkspace(t)
	// Scan first to populate data
	run(t, root, "runtime", "scan")

	out, err := run(t, root, "runtime", "status")
	if err != nil {
		t.Fatalf("runtime status failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "KIND") || !strings.Contains(out, "STATUS") {
		t.Error("runtime status should print table headers")
	}
}

func TestRuntimeReady(t *testing.T) {
	root := setupWorkspace(t)
	run(t, root, "runtime", "scan")

	out, err := run(t, root, "runtime", "ready")
	if err != nil {
		t.Fatalf("runtime ready failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "KIND") {
		t.Error("runtime ready should print table headers")
	}
}

func TestRuntimeBlocked(t *testing.T) {
	root := setupWorkspace(t)
	run(t, root, "runtime", "scan")

	out, err := run(t, root, "runtime", "blocked")
	if err != nil {
		t.Fatalf("runtime blocked failed: %v\n%s", err, out)
	}
	// May show "No blocked objects." or a table
	if !strings.Contains(out, "KIND") && !strings.Contains(out, "No blocked") {
		t.Error("runtime blocked should print headers or 'No blocked'")
	}
}

func TestRuntimeReport(t *testing.T) {
	root := setupWorkspace(t)
	run(t, root, "runtime", "scan")

	out, err := run(t, root, "runtime", "report")
	if err != nil {
		t.Fatalf("runtime report failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Runtime Readiness") {
		t.Error("runtime report should contain 'Runtime Readiness'")
	}
}

func TestRuntimeNoArgs(t *testing.T) {
	root := setupWorkspace(t)
	_, err := run(t, root, "runtime")
	if err == nil {
		t.Error("runtime with no args should fail with usage")
	}
}

func TestCommandsNoArgsDontPanic(t *testing.T) {
	// Commands that accept subcommands should not panic when called with no sub-args
	commands := []string{
		"federation",
		"repo",
		"workflow",
		"capabilities",
		"modules",
		"kernel-registry",
		"discovery",
		"event-bus",
		"patch-engine",
		"kernel",
		"runtime",
	}

	root := setupWorkspace(t)
	for _, cmd := range commands {
		out, err := run(t, root, cmd)
		// Some may fail (e.g. no workspace), but none should panic
		if err != nil {
			if strings.Contains(out, "panic") || strings.Contains(out, "runtime error") {
				t.Errorf("%s panicked: %s", cmd, out)
			}
		}
	}
}
