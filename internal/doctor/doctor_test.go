package doctor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func testConfig(t *testing.T) config.Config {
	t.Helper()
	root := t.TempDir()
	osDir := filepath.Join(root, "AIFT-OS")

	required := []string{
		"cmd/aift",
		"internal/config",
		"internal/workspace",
		"internal/gitx",
		"internal/doctor",
		"internal/registry",
		"internal/manifests",
		"internal/reports",
		"internal/plugins",
		"internal/sync",
		"internal/kernel",
		"install",
		"tests",
		"docs",
		"schemas",
		"registry",
		"reports",
		"bin",
	}

	for _, dir := range required {
		if err := os.MkdirAll(filepath.Join(osDir, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("AIFT_ROOT", root)
	t.Setenv("AIFT_OS_HOME", osDir)
	return config.Load()
}

func skipIfNoShell(t *testing.T, err error) {
	t.Helper()
	if errors.Is(err, ErrShellUnavailable) {
		t.Skip(err.Error())
	}
}

func TestDoctorRunDoesNotPanic(t *testing.T) {
	cfg := testConfig(t)
	if err := Run(cfg); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func TestDoctorRepairDoesNotPanic(t *testing.T) {
	cfg := testConfig(t)
	if err := Repair(cfg); err != nil {
		skipIfNoShell(t, err)
		t.Fatalf("Repair returned error: %v", err)
	}
}

func TestDoctorGitDoesNotPanic(t *testing.T) {
	cfg := testConfig(t)
	if err := Git(cfg); err != nil {
		skipIfNoShell(t, err)
		t.Fatalf("Git returned error: %v", err)
	}
}

func TestDoctorFullDoesNotPanic(t *testing.T) {
	cfg := testConfig(t)
	if err := Full(cfg); err != nil {
		skipIfNoShell(t, err)
		t.Fatalf("Full returned error: %v", err)
	}
}
