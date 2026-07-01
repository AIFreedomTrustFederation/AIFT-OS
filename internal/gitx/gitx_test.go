package gitx

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	requireGit(t)
	dir := t.TempDir()
	if out, err := Run(dir, "init"); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}
	if out, err := Run(dir, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config email failed: %v\n%s", err, out)
	}
	if out, err := Run(dir, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config name failed: %v\n%s", err, out)
	}
	return dir
}

func TestGitHelpersReportBranchRemoteAndDirtyState(t *testing.T) {
	dir := initRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if out, err := Run(dir, "add", "README.md"); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}
	if out, err := Run(dir, "commit", "-m", "initial"); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, out)
	}
	if got := Branch(dir); got == "" || got == "unknown" {
		t.Fatalf("branch = %q", got)
	}
	if Remote(dir) != "" {
		t.Fatalf("remote = %q, want empty before origin configured", Remote(dir))
	}
	if Dirty(dir) {
		t.Fatal("new repository should not be dirty")
	}

	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("change"), 0644); err != nil {
		t.Fatal(err)
	}
	if !Dirty(dir) {
		t.Fatal("repository with untracked file should be dirty")
	}
	if out, err := Run(dir, "remote", "add", "origin", "https://example.com/repo.git"); err != nil {
		t.Fatalf("git remote add failed: %v\n%s", err, out)
	}
	if got := Remote(dir); got != "https://example.com/repo.git" {
		t.Fatalf("remote = %q", got)
	}
}
