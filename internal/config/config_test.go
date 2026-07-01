package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("AIFT_ROOT")
	os.Unsetenv("AIFT_OS_HOME")

	cfg := Load()

	home, _ := os.UserHomeDir()
	expectedRoot := filepath.Join(home, "AIFT")
	if cfg.Root != expectedRoot {
		t.Errorf("Root = %q, want %q", cfg.Root, expectedRoot)
	}

	expectedOSHome := filepath.Join(expectedRoot, "AIFT-OS")
	if cfg.OSHome != expectedOSHome {
		t.Errorf("OSHome = %q, want %q", cfg.OSHome, expectedOSHome)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	t.Setenv("AIFT_ROOT", "/custom/root")
	t.Setenv("AIFT_OS_HOME", "/custom/oshome")

	cfg := Load()

	if cfg.Root != "/custom/root" {
		t.Errorf("Root = %q, want /custom/root", cfg.Root)
	}
	if cfg.OSHome != "/custom/oshome" {
		t.Errorf("OSHome = %q, want /custom/oshome", cfg.OSHome)
	}
}

func TestLoadRootOnly(t *testing.T) {
	t.Setenv("AIFT_ROOT", "/custom/root")
	os.Unsetenv("AIFT_OS_HOME")

	cfg := Load()

	if cfg.Root != "/custom/root" {
		t.Errorf("Root = %q", cfg.Root)
	}
	expectedOSHome := filepath.Join("/custom/root", "AIFT-OS")
	if cfg.OSHome != expectedOSHome {
		t.Errorf("OSHome = %q, want %q", cfg.OSHome, expectedOSHome)
	}
}

func TestConfigFields(t *testing.T) {
	cfg := Config{Root: "/a", OSHome: "/b"}
	if cfg.Root != "/a" || cfg.OSHome != "/b" {
		t.Error("Config fields not set correctly")
	}
}
