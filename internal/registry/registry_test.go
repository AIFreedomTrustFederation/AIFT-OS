package registry

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func makeRepo(t *testing.T, root, name string, manifest bool) string {
	t.Helper()
	repo := filepath.Join(root, name)
	if err := os.MkdirAll(repo, 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := exec.LookPath("git"); err == nil {
		cmd := exec.Command("git", "init")
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git init failed: %v\n%s", err, out)
		}
	} else if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if manifest {
		dir := filepath.Join(repo, ".aift")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		data := []byte(`{"name":"` + name + `","role":"test"}`)
		if err := os.WriteFile(filepath.Join(dir, "repo.json"), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return repo
}

func TestGenerateWritesRepositoryRegistry(t *testing.T) {
	root := t.TempDir()
	osHome := filepath.Join(root, "AIFT-OS")
	makeRepo(t, root, "alpha", true)
	makeRepo(t, root, "beta", false)

	if err := Generate(config.Config{Root: root, OSHome: osHome}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(osHome, "registry", "repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var records []Record
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("record count = %d, want 2", len(records))
	}
	if records[0].Name != "alpha" || !records[0].ManifestValid {
		t.Fatalf("first record = %#v", records[0])
	}
	if records[1].Name != "beta" || records[1].ManifestValid {
		t.Fatalf("second record = %#v", records[1])
	}
}
