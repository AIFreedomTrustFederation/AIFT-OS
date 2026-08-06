package uxi

import (
	"os"
	"path/filepath"
	"reflect"
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

func TestCapabilityStatusIsEvidenceBased(t *testing.T) {
	cases := []struct {
		name string
		caps []Capability
		want string
	}{
		{"planned only", []Capability{{Name: "deploy", Status: "planned"}}, "detected"},
		{"ready", []Capability{{Name: "status", Status: "ready"}}, "ready"},
		{"blocked wins", []Capability{{Name: "status", Status: "ready"}, {Name: "build", Status: "broken"}}, "blocked"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := aggregateCapabilityStatus(tc.caps); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestDiscoverRepositoryWithGitFile(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "worktree")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".git"), []byte("gitdir: ../.git/worktrees/worktree\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	repos, err := DiscoverRepositories(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 {
		t.Fatalf("repos = %d", len(repos))
	}
}

func TestDiscoverSymlinkedRepository(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	repo := filepath.Join(outside, "linked")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(repo, filepath.Join(root, "linked")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	repos, err := DiscoverRepositories(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 {
		t.Fatalf("repos=%#v", repos)
	}
}

func TestRepositoryEvidenceOrderIsStable(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "mixed")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"go.mod", "package.json", "Cargo.toml", "pyproject.toml"} {
		if err := os.WriteFile(filepath.Join(repo, name), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	first, err := DiscoverRepositories(root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := DiscoverRepositories(root)
	if err != nil {
		t.Fatal(err)
	}
	var a, b []string
	for _, evidence := range first[0].Evidence {
		a = append(a, evidence.Kind+":"+filepath.Base(evidence.Source))
	}
	for _, evidence := range second[0].Evidence {
		b = append(b, evidence.Kind+":"+filepath.Base(evidence.Source))
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("first=%v second=%v", a, b)
	}
}
