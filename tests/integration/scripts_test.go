package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAllShellScriptsPassSyntaxCheck runs bash -n on every .sh file in the repo.
func TestAllShellScriptsPassSyntaxCheck(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash is required for shell syntax checks")
	}

	root := repoRoot()

	var scripts []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Skip hidden dirs except .github
		base := filepath.Base(path)
		if info.IsDir() && strings.HasPrefix(base, ".") && base != ".github" {
			return filepath.SkipDir
		}
		// Skip node_modules, vendor, etc.
		if info.IsDir() && (base == "node_modules" || base == "vendor" || base == "legacy" || base == "reports" || base == "AI-Code-Training") {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(path, ".sh") {
			scripts = append(scripts, path)
		}
		return nil
	})

	if len(scripts) == 0 {
		t.Fatal("found zero .sh files in repo")
	}

	for _, script := range scripts {
		rel, _ := filepath.Rel(root, script)
		rel = filepath.ToSlash(rel)
		t.Run(rel, func(t *testing.T) {
			cmd := exec.Command("bash", "-n", script)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("bash -n failed: %s\n%s", err, out)
			}
		})
	}
}

// TestMutatingScriptsSourceHarness checks that scripts that make changes
// either source aift-run.sh or document why not.
func TestMutatingScriptsSourceHarness(t *testing.T) {
	root := repoRoot()

	// Scripts that are standalone utilities or legacy (predate the harness)
	exempt := map[string]bool{
		"scripts/coverage.sh":          true,
		"scripts/check-coverage.sh":    true,
		"scripts/lib/discovery.sh":     true,
		"scripts/lib/aift-run.sh":      true,
		"tests/harness-syntax-test.sh": true,
		// Legacy scripts that predate the harness
		"scripts/federation-graph.sh":   true,
		"scripts/generate-dashboard.sh": true,
		"scripts/inspect.sh":            true,
		"scripts/install-deps-all.sh":   true,
		"scripts/pull-all.sh":           true,
		"scripts/status-all.sh":         true,
		"scripts/verify-all.sh":         true,
	}

	var scripts []string
	filepath.Walk(filepath.Join(root, "scripts"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".sh") {
			scripts = append(scripts, path)
		}
		return nil
	})

	for _, script := range scripts {
		rel, _ := filepath.Rel(root, script)
		rel = filepath.ToSlash(rel)
		if exempt[rel] {
			continue
		}

		data, err := os.ReadFile(script)
		if err != nil {
			t.Errorf("cannot read %s: %v", rel, err)
			continue
		}
		content := string(data)

		// Check if it sources aift-run.sh or explicitly documents why not
		sourcesHarness := strings.Contains(content, "aift-run.sh")
		documentsExemption := strings.Contains(content, "# no-harness:") ||
			strings.Contains(content, "# harness-exempt:")

		if !sourcesHarness && !documentsExemption {
			t.Errorf("%s is a mutating script that does not source aift-run.sh "+
				"and has no exemption comment (# no-harness: <reason>)", rel)
		}
	}
}
