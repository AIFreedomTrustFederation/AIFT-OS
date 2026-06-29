package jsonfile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "test.json")

	data := map[string]string{"key": "value"}
	err := Write(path, data, false)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}

	var decoded map[string]string
	if err := json.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded["key"] != "value" {
		t.Errorf("key = %q, want value", decoded["key"])
	}
}

func TestWriteIndented(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	data := map[string]int{"a": 1}
	Write(path, data, false)

	content, _ := os.ReadFile(path)
	s := string(content)
	if s[0] != '{' {
		t.Error("should start with {")
	}
	if s[len(s)-1] != '\n' {
		t.Error("should end with newline")
	}
}

func TestWriteCreatesDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "test.json")

	err := Write(path, "hello", false)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Error("parent directories not created")
	}
}

func TestWriteOverwrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	Write(path, map[string]int{"v": 1}, false)
	Write(path, map[string]int{"v": 2}, false)

	content, _ := os.ReadFile(path)
	var decoded map[string]int
	json.Unmarshal(content, &decoded)
	if decoded["v"] != 2 {
		t.Errorf("v = %d, want 2 (overwritten)", decoded["v"])
	}
}

func TestReadPackageCommands(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
		"scripts": {
			"build": "next build",
			"dev": "next dev",
			"test": "jest"
		}
	}`), 0644)

	commands := map[string]string{}
	ReadPackageCommands(dir, commands)

	if commands["npm:build"] != "npm run build" {
		t.Errorf("npm:build = %q, want 'npm run build'", commands["npm:build"])
	}
	if commands["npm:dev"] != "npm run dev" {
		t.Errorf("npm:dev = %q, want 'npm run dev'", commands["npm:dev"])
	}
	if commands["npm:test"] != "npm run test" {
		t.Errorf("npm:test = %q, want 'npm run test'", commands["npm:test"])
	}
}

func TestReadPackageCommandsMissing(t *testing.T) {
	dir := t.TempDir()
	commands := map[string]string{}
	ReadPackageCommands(dir, commands)

	if len(commands) != 0 {
		t.Errorf("expected empty commands for missing package.json, got %d", len(commands))
	}
}

func TestReadPackageCommandsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("not json"), 0644)

	commands := map[string]string{}
	ReadPackageCommands(dir, commands)

	if len(commands) != 0 {
		t.Errorf("expected empty commands for invalid JSON, got %d", len(commands))
	}
}

func TestReadPackageCommandsNoScripts(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test"}`), 0644)

	commands := map[string]string{}
	ReadPackageCommands(dir, commands)

	if len(commands) != 0 {
		t.Errorf("expected empty commands for no scripts, got %d", len(commands))
	}
}

func TestReadNamedList(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "capabilities.json"), []byte(`{
		"capabilities": [
			{"name": "build", "status": "ready"},
			{"name": "test", "status": "v1"},
			{"name": "", "status": "empty"}
		]
	}`), 0644)

	names := ReadNamedList(dir, "capabilities.json", "capabilities")
	if len(names) != 2 {
		t.Fatalf("expected 2 names (empty name skipped), got %d: %v", len(names), names)
	}
	if names[0] != "build" || names[1] != "test" {
		t.Errorf("names = %v, want [build, test]", names)
	}
}

func TestReadNamedListMissing(t *testing.T) {
	dir := t.TempDir()
	names := ReadNamedList(dir, "capabilities.json", "capabilities")
	if len(names) != 0 {
		t.Errorf("expected empty for missing file, got %v", names)
	}
}

func TestReadNamedListInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "test.json"), []byte("not json"), 0644)

	names := ReadNamedList(dir, "test.json", "items")
	if len(names) != 0 {
		t.Errorf("expected empty for invalid JSON, got %v", names)
	}
}

func TestReadNamedListWrongField(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "test.json"), []byte(`{
		"capabilities": [{"name": "build"}]
	}`), 0644)

	names := ReadNamedList(dir, "test.json", "services")
	if len(names) != 0 {
		t.Errorf("expected empty for wrong field, got %v", names)
	}
}
