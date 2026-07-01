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
