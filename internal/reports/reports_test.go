package reports

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func makeReportRepo(t *testing.T, root, name string, manifest bool) {
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
		if err := os.WriteFile(filepath.Join(dir, "repo.json"), []byte(`{"name":"`+name+`","role":"test"}`), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDashboardAndDepsWriteReports(t *testing.T) {
	root := t.TempDir()
	osHome := filepath.Join(root, "AIFT-OS")
	cfg := config.Config{Root: root, OSHome: osHome}
	makeReportRepo(t, root, "alpha", true)
	makeReportRepo(t, root, "beta", false)

	if err := Dashboard(cfg); err != nil {
		t.Fatal(err)
	}
	if err := Deps(cfg); err != nil {
		t.Fatal(err)
	}

	dashboard, err := os.ReadFile(filepath.Join(osHome, "reports", "dashboard.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(dashboard), "Total repositories: 2") {
		t.Fatalf("dashboard did not include repo total:\n%s", dashboard)
	}
	if !strings.Contains(string(dashboard), "Valid manifests: 1") {
		t.Fatalf("dashboard did not include valid manifest count:\n%s", dashboard)
	}

	deps, err := os.ReadFile(filepath.Join(osHome, "reports", "dependency-graph.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(deps), "| `alpha` | `[]` |") {
		t.Fatalf("dependency report missing alpha row:\n%s", deps)
	}
}
