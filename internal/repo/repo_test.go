package repo

import (
	"encoding/json"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"  spaces  ", "spaces"},
		{"path/traversal", "traversal"},
		{"../parent", "parent"},
		{".", ""},
		{"..", ""},
		{"normal-repo", "normal-repo"},
		{"", ""},
	}
	for _, tt := range tests {
		got := NormalizeName(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeNameBackslash(t *testing.T) {
	// On Linux, backslash is not a path separator
	got := NormalizeName("path\\traversal")
	if got == "" {
		t.Error("NormalizeName with backslash should not return empty")
	}
}

func TestInfoJSON(t *testing.T) {
	info := Info{
		Name:          "test-repo",
		Path:          "/tmp/test",
		Branch:        "main",
		Remote:        "https://github.com/test/test.git",
		Dirty:         false,
		ManifestPath:  "/tmp/test/.aift/repo.json",
		ManifestValid: true,
		CommandsPath:  "/tmp/test/.aift/commands",
		WorkflowsPath: "/tmp/test/.aift/workflows.json",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Info
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Name != "test-repo" {
		t.Errorf("Name = %q", decoded.Name)
	}
	if decoded.Branch != "main" {
		t.Errorf("Branch = %q", decoded.Branch)
	}
	if decoded.ManifestValid != true {
		t.Error("ManifestValid should be true")
	}
}

func TestInfoDirtyJSON(t *testing.T) {
	info := Info{Name: "test", Dirty: true}
	data, _ := json.Marshal(info)
	var decoded Info
	json.Unmarshal(data, &decoded)
	if !decoded.Dirty {
		t.Error("Dirty should be true after round-trip")
	}
}

func TestRunCommandInvalidName(t *testing.T) {
	dir := t.TempDir()
	cfg := testCfg(t, dir)

	// Command names with slashes should be rejected
	err := RunCommand(cfg, "nonexistent", "../../etc/passwd", nil)
	if err == nil {
		t.Error("RunCommand should reject command names with slashes")
	}
}

func TestRunCommandDotDot(t *testing.T) {
	dir := t.TempDir()
	cfg := testCfg(t, dir)

	err := RunCommand(cfg, "nonexistent", "..", nil)
	if err == nil {
		t.Error("RunCommand should reject '..' as command name")
	}
}

func TestRunCommandDot(t *testing.T) {
	dir := t.TempDir()
	cfg := testCfg(t, dir)

	err := RunCommand(cfg, "nonexistent", ".", nil)
	if err == nil {
		t.Error("RunCommand should reject '.' as command name")
	}
}

func testCfg(t *testing.T, dir string) config.Load()
}
