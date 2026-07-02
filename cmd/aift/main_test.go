package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func captureOutput(t *testing.T, fn func() int) (int, string, string) {
	t.Helper()

	oldOut := os.Stdout
	oldErr := os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = outW
	os.Stderr = errW

	code := fn()

	if err := outW.Close(); err != nil {
		t.Fatal(err)
	}
	if err := errW.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = oldOut
	os.Stderr = oldErr

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if _, err := io.Copy(&stdout, outR); err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(&stderr, errR); err != nil {
		t.Fatal(err)
	}

	return code, stdout.String(), stderr.String()
}

func TestRunUsesFirstArgumentAsCommand(t *testing.T) {
	code, stdout, stderr := captureOutput(t, func() int {
		return run([]string{"--", "registry"})
	})
	if code != 0 {
		t.Fatalf("run registry exit code = %d, stderr = %s", code, stderr)
	}

	var payload struct {
		Commands []Command `json:"commands"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("registry output is not JSON: %v\n%s", err, stdout)
	}
	if len(payload.Commands) == 0 {
		t.Fatal("registry returned no commands")
	}
}

func TestRunUnknownCommandReturnsUsageError(t *testing.T) {
	code, _, stderr := captureOutput(t, func() int {
		return run([]string{"does-not-exist"})
	})
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if want := "unknown command: does-not-exist"; !bytes.Contains([]byte(stderr), []byte(want)) {
		t.Fatalf("stderr = %q, want %q", stderr, want)
	}
}

func TestResolveFindsAliasesAndPlannedCommandsStayPlanned(t *testing.T) {
	cmd, ok := resolve(commands(), "fed")
	if !ok {
		t.Fatal("resolve fed alias failed")
	}
	if cmd.Name != "federation" || cmd.Status != "planned" {
		t.Fatalf("resolved command = %#v", cmd)
	}

	code, stdout, stderr := captureOutput(t, func() int {
		return run([]string{"federation"})
	})
	if code != 0 {
		t.Fatalf("planned command exit code = %d, stderr = %s", code, stderr)
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("planned command output is not JSON: %v\n%s", err, stdout)
	}
	if payload["status"] != "planned" {
		t.Fatalf("status = %q, want planned", payload["status"])
	}
}

func TestAggregateAndFileCheck(t *testing.T) {
	if got := aggregate([]Check{{Status: "pass"}, {Status: "planned"}}); got != "partial" {
		t.Fatalf("aggregate planned = %q, want partial", got)
	}
	if got := aggregate([]Check{{Status: "pass"}, {Status: "fail"}}); got != "fail" {
		t.Fatalf("aggregate fail = %q, want fail", got)
	}

	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatal(err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "present.txt"), []byte("ok"), 0644); err != nil {
		t.Fatal(err)
	}

	if got := fileCheck("present", "present.txt", "exists").Status; got != "pass" {
		t.Fatalf("present status = %q, want pass", got)
	}
	if got := fileCheck("missing", "missing.txt", "missing").Status; got != "planned" {
		t.Fatalf("missing status = %q, want planned", got)
	}
}

func TestDiscoverAppsFindsRegistryLocalAndSiblingManifests(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "AIFT-OS")
	sibling := filepath.Join(parent, "BookSmith-AI")

	writeApp(t, root, "registry/apps/core.json", `{"id":"core-app","name":"Core App"}`)
	writeApp(t, root, ".aift/apps/local.json", `{"id":"local-app","name":"Local App"}`)
	writeApp(t, sibling, ".aift/apps/sibling.json", `{"id":"sibling-app","name":"Sibling App"}`)

	result, err := discoverApps(root)
	if err != nil {
		t.Fatalf("discoverApps error = %v", err)
	}

	ids := map[string]bool{}
	for _, app := range result.Apps {
		ids[app.ID] = true
		if app.Source == "" {
			t.Fatalf("app source missing: %#v", app)
		}
	}

	for _, id := range []string{"core-app", "local-app", "sibling-app"} {
		if !ids[id] {
			t.Fatalf("expected discovered app id %q in %#v", id, result.Apps)
		}
	}
}

func TestDuplicateAppIDsAreHandledTruthfully(t *testing.T) {
	root := t.TempDir()
	writeApp(t, root, "registry/apps/one.json", `{"id":"dup-app","name":"First"}`)
	writeApp(t, root, ".aift/apps/two.json", `{"id":"dup-app","name":"Second"}`)

	result, err := discoverApps(root)
	if err != nil {
		t.Fatalf("discoverApps error = %v", err)
	}

	if len(result.DuplicateIDs) != 1 || result.DuplicateIDs[0] != "dup-app" {
		t.Fatalf("duplicate ids = %#v, want dup-app", result.DuplicateIDs)
	}

	matches := appsByID(result.Apps, "dup-app")
	if len(matches) != 2 {
		t.Fatalf("matches = %d, want 2", len(matches))
	}
	for _, app := range matches {
		if !app.Duplicate {
			t.Fatalf("duplicate app not marked truthfully: %#v", app)
		}
	}
}

func TestAppsLaunchIsPlannedNotActive(t *testing.T) {
	dir := t.TempDir()
	writeApp(t, dir, "registry/apps/booksmith-studio.json", `{"id":"booksmith-studio","name":"BookSmith Studio"}`)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatal(err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := captureOutput(t, func() int {
		return run([]string{"apps", "launch", "booksmith-studio", "--plan"})
	})
	if code != 0 {
		t.Fatalf("launch exit code = %d, stderr = %s", code, stderr)
	}

	var payload struct {
		Status string `json:"status"`
		Active bool  `json:"active"`
		ID     string `json:"id"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("launch output is not JSON: %v\n%s", err, stdout)
	}
	if payload.Status != "planned" {
		t.Fatalf("status = %q, want planned", payload.Status)
	}
	if payload.Active {
		t.Fatal("launch reported active=true, want false")
	}
	if payload.ID != "booksmith-studio" {
		t.Fatalf("id = %q, want booksmith-studio", payload.ID)
	}
}

func TestMalformedAppJSONFailsHonestly(t *testing.T) {
	root := t.TempDir()
	writeApp(t, root, "registry/apps/broken.json", `{"id":`)

	_, err := discoverApps(root)
	if err == nil {
		t.Fatal("discoverApps succeeded, want malformed JSON error")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("malformed app manifest")) {
		t.Fatalf("error = %q, want malformed app manifest", err.Error())
	}
}

func writeApp(t *testing.T, root, rel, content string) {
	t.Helper()

	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
