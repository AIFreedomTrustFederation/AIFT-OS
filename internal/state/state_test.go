package state

import (
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestSaveAndLoadRuntimeState(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}

	st := New()
	st.Services["api"] = "running"
	if err := Save(cfg, st); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name != "AIFT-OS" || loaded.Status != "running" {
		t.Fatalf("loaded state = %#v", loaded)
	}
	if loaded.Services["api"] != "running" {
		t.Fatalf("api service status = %q", loaded.Services["api"])
	}
	if Path(cfg) != filepath.Join(dir, "var", "runtime-state.json") {
		t.Fatalf("path = %q", Path(cfg))
	}
}

func TestLoadMissingStateReturnsNewStateAndError(t *testing.T) {
	dir := t.TempDir()
	st, err := Load(config.Config{Root: dir, OSHome: dir})
	if err == nil {
		t.Fatal("Load missing state error = nil")
	}
	if st.Name != "AIFT-OS" || st.Status != "running" {
		t.Fatalf("fallback state = %#v", st)
	}
}
