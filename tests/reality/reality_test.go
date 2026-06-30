package reality

import (
	"os"
	"path/filepath"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root")
		}

		dir = parent
	}
}

func TestRepositoryHasGoModule(t *testing.T) {
	root := repoRoot(t)

	required := []string{
		"go.mod",
		"cmd/aift/main.go",
		"internal/config",
	}

	for _, name := range required {
		path := filepath.Join(root, name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("required path missing: %s: %v", name, err)
		}
	}
}

func TestRepositoryHasRuntimeArchitectureDocs(t *testing.T) {
	root := repoRoot(t)

	candidates := []string{
		"docs/architecture/PHASE2-CENTRAL-FEDERATION-RUNTIME.md",
		"docs/architecture/PHASE3-RECURSIVE-DISCOVERY-ENGINE.md",
		"docs/architecture/PHASE4-FEDERATION-RUNTIME-GRAPH.md",
		"docs/architecture/PHASE6-RUNTIME-EXECUTION-ENGINE.md",
		"docs/architecture/PHASE7-AIFT-DOCTOR-GIT-HOUSEKEEPING.md",
	}

	found := 0
	for _, name := range candidates {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			found++
		}
	}

	if found == 0 {
		t.Fatal("expected at least one runtime architecture document")
	}
}
