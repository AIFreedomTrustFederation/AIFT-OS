package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestFindReposDiscoversGitDirectoriesAndSkipsNodeModules(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{
		filepath.Join(root, "beta", ".git"),
		filepath.Join(root, "alpha", ".git"),
		filepath.Join(root, "node_modules", "ignored", ".git"),
	} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}

	repos, err := FindRepos(config.Config{Root: root, OSHome: filepath.Join(root, "AIFT-OS")})
	if err != nil {
		t.Fatal(err)
	}

	if len(repos) != 2 {
		t.Fatalf("repos = %#v, want 2 repos", repos)
	}
	if repos[0].Name != "alpha" || repos[1].Name != "beta" {
		t.Fatalf("repos not sorted or incorrectly discovered: %#v", repos)
	}
}
