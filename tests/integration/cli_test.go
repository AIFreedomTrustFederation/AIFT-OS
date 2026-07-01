package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binary string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "aift-cli-test-*")
	if err != nil {
		panic(err)
	}

	name := "aift"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary = filepath.Join(tmp, name)

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

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = repoRoot()
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestHelp(t *testing.T) {
	out, err := run(t, "help")
	if err != nil {
		t.Fatalf("help failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Commands:") {
		t.Fatalf("help output should contain commands: %s", out)
	}
}

func TestHelpFlagAndSeparator(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"--", "help"}} {
		out, err := run(t, args...)
		if err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
		if !strings.Contains(out, "Commands:") {
			t.Fatalf("%v should show help: %s", args, out)
		}
	}
}

func TestNoArgsShowsHelp(t *testing.T) {
	out, err := run(t)
	if err != nil {
		t.Fatalf("no args should not fail: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Commands:") {
		t.Fatalf("no args should show help: %s", out)
	}
}

func TestUnknownCommand(t *testing.T) {
	out, err := run(t, "nonexistent-command")
	if err == nil {
		t.Fatal("unknown command should return non-zero")
	}
	if !strings.Contains(out, "unknown command: nonexistent-command") {
		t.Fatalf("unknown command output missing command name: %s", out)
	}
}

func TestStatusJSON(t *testing.T) {
	out, err := run(t, "status")
	if err != nil {
		t.Fatalf("status failed: %v\n%s", err, out)
	}
	var payload struct {
		Status string `json:"status"`
		Root   string `json:"root"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"checks"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("status should be JSON: %v\n%s", err, out)
	}
	if payload.Status != "pass" {
		t.Fatalf("status = %q, want pass", payload.Status)
	}
	if payload.Root != repoRoot() {
		t.Fatalf("root = %q, want %q", payload.Root, repoRoot())
	}
	if len(payload.Checks) == 0 {
		t.Fatal("status should include checks")
	}
}

func TestVerifyJSON(t *testing.T) {
	out, err := run(t, "verify")
	if err != nil {
		t.Fatalf("verify failed: %v\n%s", err, out)
	}
	var payload struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("verify should be JSON: %v\n%s", err, out)
	}
	if payload.Status != "pass" {
		t.Fatalf("verify status = %q, want pass", payload.Status)
	}
}

func TestRegistryJSON(t *testing.T) {
	out, err := run(t, "registry")
	if err != nil {
		t.Fatalf("registry failed: %v\n%s", err, out)
	}
	var payload struct {
		Commands []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"commands"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("registry should be JSON: %v\n%s", err, out)
	}
	seen := map[string]string{}
	for _, cmd := range payload.Commands {
		seen[cmd.Name] = cmd.Status
	}
	for _, name := range []string{"help", "status", "verify", "registry", "bootstrap"} {
		if seen[name] != "active" {
			t.Fatalf("%s status = %q, want active", name, seen[name])
		}
	}
	for _, name := range []string{"federation", "repo", "workflow"} {
		if seen[name] != "planned" {
			t.Fatalf("%s status = %q, want planned", name, seen[name])
		}
	}
}

func TestBootstrapJSON(t *testing.T) {
	out, err := run(t, "bootstrap")
	if err != nil {
		t.Fatalf("bootstrap failed: %v\n%s", err, out)
	}
	var payload struct {
		Discovery map[string]bool `json:"discovery"`
		Commands  []any           `json:"commands"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("bootstrap should be JSON: %v\n%s", err, out)
	}
	if !payload.Discovery["cmd_aift"] {
		t.Fatal("bootstrap should report cmd_aift discovery")
	}
	if len(payload.Commands) == 0 {
		t.Fatal("bootstrap should include commands")
	}
}

func TestPlannedCommandsReturnHonestJSON(t *testing.T) {
	for _, name := range []string{"federation", "repo", "workflow"} {
		out, err := run(t, name)
		if err != nil {
			t.Fatalf("%s should report planned JSON without failing: %v\n%s", name, err, out)
		}
		var payload struct {
			Command string `json:"command"`
			Status  string `json:"status"`
		}
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatalf("%s should return JSON: %v\n%s", name, err, out)
		}
		if payload.Command != name || payload.Status != "planned" {
			t.Fatalf("%s payload = %+v, want planned command JSON", name, payload)
		}
	}
}
