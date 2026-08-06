package uxi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverRepositories(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "AIFT-OS")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, ".aift"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".aift", "capabilities.json"), []byte(`{"capabilities":[{"name":"status","status":"ready"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	repos, err := DiscoverRepositories(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 {
		t.Fatalf("repos = %d", len(repos))
	}
	if repos[0].Role != "federation-kernel" || repos[0].Status != "ready" {
		t.Fatalf("repo = %#v", repos[0])
	}
}
