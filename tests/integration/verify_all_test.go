package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func writeVerifyFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
}

func runWorkspaceVerifier(t *testing.T, manifest, tool, failingRepo string) (string, string, error) {
	t.Helper()
	root := t.TempDir()
	workspace := filepath.Join(root, "workspace")
	bin := filepath.Join(root, "bin")
	logPath := filepath.Join(root, "calls.log")
	for _, name := range []string{"broken", "healthy"} {
		writeVerifyFixture(t, filepath.Join(workspace, name, ".git", "HEAD"), "ref: refs/heads/main\n")
		writeVerifyFixture(t, filepath.Join(workspace, name, manifest), "{}\n")
	}
	writeVerifyFixture(t, filepath.Join(bin, "git"), "#!/bin/sh\nexit 0\n")
	if tool != "" {
		writeVerifyFixture(t, filepath.Join(bin, tool), `#!/bin/sh
printf '%s:%s\n' "${PWD##*/}" "$*" >> "$AIFT_AUDIT_LOG"
if [ "${PWD##*/}" = "$AIFT_AUDIT_FAIL_REPO" ]; then exit 17; fi
exit 0
`)
	}
	cmd := exec.Command("bash", filepath.Join(repoRoot(), "scripts", "verify-all.sh"), workspace)
	cmd.Env = append(os.Environ(), "PATH="+bin, "AIFT_AUDIT_LOG="+logPath, "AIFT_AUDIT_FAIL_REPO="+failingRepo)
	output, err := cmd.CombinedOutput()
	calls, readErr := os.ReadFile(logPath)
	if readErr != nil && !os.IsNotExist(readErr) {
		t.Fatal(readErr)
	}
	return string(output), string(calls), err
}

func TestWorkspaceVerifierPropagatesCheckFailures(t *testing.T) {
	cases := []struct{ name, manifest, tool string }{
		{"npm", "package.json", "npm"},
		{"cargo", "Cargo.toml", "cargo"},
		{"go", "go.mod", "go"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, calls, err := runWorkspaceVerifier(t, tc.manifest, tc.tool, "broken")
			if err == nil {
				t.Fatalf("failed %s check reported success:\n%s", tc.tool, output)
			}
			if !strings.Contains(calls, "healthy:") {
				t.Fatalf("verifier stopped before checking healthy repo; calls: %q", calls)
			}
			if !strings.Contains(output, "FAIL:") {
				t.Fatalf("missing failure evidence: %s", output)
			}
		})
	}
}

func TestWorkspaceVerifierFailsWithoutRequiredTool(t *testing.T) {
	output, _, err := runWorkspaceVerifier(t, "package.json", "", "")
	if err == nil || !strings.Contains(output, "npm unavailable") {
		t.Fatalf("missing npm must fail with explanation; err=%v, output=%s", err, output)
	}
}

func TestWorkspaceVerifierFailsWithNoRepositories(t *testing.T) {
	workspace := t.TempDir()
	cmd := exec.Command("bash", filepath.Join(repoRoot(), "scripts", "verify-all.sh"), workspace)
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "no Git repositories") {
		t.Fatalf("empty workspace must fail; err=%v, output=%s", err, output)
	}
}

func TestWorkspaceVerifierSucceedsWhenChecksPass(t *testing.T) {
	output, calls, err := runWorkspaceVerifier(t, "package.json", "npm", "")
	if err != nil || !strings.Contains(output, "PASS: verification completed") {
		t.Fatalf("successful checks must pass; err=%v, output=%s", err, output)
	}
	if !strings.Contains(calls, "broken:") || !strings.Contains(calls, "healthy:") {
		t.Fatalf("both repositories should be checked: %q", calls)
	}
}
