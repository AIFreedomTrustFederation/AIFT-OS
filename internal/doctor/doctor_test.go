package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestDoctorRunDoesNotPanic(t *testing.T) {
	root := t.TempDir()
	osDir := filepath.Join(root, "AIFT-OS")

	if err := os.MkdirAll(osDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Load()

	if err := Run(cfg); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func TestDoctorRepairDoesNotPanic(t *testing.T) {
	root := t.TempDir()
	osDir := filepath.Join(root, "AIFT-OS")

	if err := os.MkdirAll(filepath.Join(osDir, "registry"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Load()

	if err := Repair(cfg); err != nil {
		t.Fatalf("Repair returned error: %v", err)
	}
}

func TestDoctorGitDoesNotPanic(t *testing.T) {
	root := t.TempDir()
	osDir := filepath.Join(root, "AIFT-OS")

	if err := os.MkdirAll(osDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Load()

	if err := Git(cfg); err != nil {
		t.Fatalf("Git returned error: %v", err)
	}
}

func TestDoctorFullDoesNotPanic(t *testing.T) {
	root := t.TempDir()
	osDir := filepath.Join(root, "AIFT-OS")

	if err := os.MkdirAll(filepath.Join(osDir, "registry"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Load()

	if err := Full(cfg); err != nil {
		t.Fatalf("Full returned error: %v", err)
	}
}
