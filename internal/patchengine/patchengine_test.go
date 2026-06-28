package patchengine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"main.go", "go-source"},
		{"internal/graph/graph.go", "go-source"},
		{"deploy.sh", "shell-script"},
		{"config/settings.json", "json-document"},
		{"README.md", "markdown-document"},
		{"unknown.txt", "file"},
		{"no-ext", "file"},
	}
	for _, tt := range tests {
		got := classify(tt.path)
		if got != tt.want {
			t.Errorf("classify(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestSafeID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"go-source.internal/graph/graph.go", "go-source.internal.graph.graph.go"},
		{"Hello World", "hello-world"},
		{"under_score", "under-score"},
		{"colon:value", "colon-value"},
		{".leading-dot", "leading-dot"},
		{"trailing-dot.", "trailing-dot"},
		{"back\\slash", "back.slash"},
		{"", ""},
	}
	for _, tt := range tests {
		got := safeID(tt.input)
		if got != tt.want {
			t.Errorf("safeID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRel(t *testing.T) {
	tests := []struct {
		root string
		path string
		want string
	}{
		{"/home/user/project", "/home/user/project/internal/main.go", "internal/main.go"},
		{"/home/user/project", "/home/user/project/go.mod", "go.mod"},
		{"/home/user/project", "/home/user/project", "."},
	}
	for _, tt := range tests {
		got := rel(tt.root, tt.path)
		if got != tt.want {
			t.Errorf("rel(%q, %q) = %q, want %q", tt.root, tt.path, got, tt.want)
		}
	}
}

func TestDiscoverFiles(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "script.sh"), []byte("#!/bin/sh"), 0644)
	os.WriteFile(filepath.Join(dir, "config.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# readme"), 0644)
	os.WriteFile(filepath.Join(dir, "data.txt"), []byte("text"), 0644)

	files := discoverFiles(dir, []string{".go", ".sh", ".json", ".md"})

	if len(files) != 4 {
		t.Errorf("expected 4 files, got %d: %v", len(files), files)
	}

	foundGo := false
	foundSh := false
	foundJson := false
	foundMd := false
	for _, f := range files {
		switch filepath.Base(f) {
		case "main.go":
			foundGo = true
		case "script.sh":
			foundSh = true
		case "config.json":
			foundJson = true
		case "README.md":
			foundMd = true
		}
	}
	if !foundGo || !foundSh || !foundJson || !foundMd {
		t.Errorf("missing expected files: go=%v sh=%v json=%v md=%v", foundGo, foundSh, foundJson, foundMd)
	}
}

func TestDiscoverFilesSkipsDirs(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("git config"), 0644)
	os.MkdirAll(filepath.Join(dir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(dir, "node_modules", "pkg.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)

	files := discoverFiles(dir, []string{".go", ".json"})

	for _, f := range files {
		base := filepath.Base(filepath.Dir(f))
		if base == ".git" || base == "node_modules" {
			t.Errorf("should skip %q directory, found %q", base, f)
		}
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(files), files)
	}
}

func TestDiscoverFilesEmpty(t *testing.T) {
	dir := t.TempDir()
	files := discoverFiles(dir, []string{".go"})
	if len(files) != 0 {
		t.Errorf("expected 0 files in empty dir, got %d", len(files))
	}
}

func TestDiscoverFilesNested(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "internal", "pkg")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(sub, "handler.go"), []byte("package pkg"), 0644)

	files := discoverFiles(dir, []string{".go"})
	if len(files) != 1 {
		t.Errorf("expected 1 nested file, got %d", len(files))
	}
}

func TestBuildPlan(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "deploy.sh"), []byte("#!/bin/sh"), 0644)

	cfg := struct {
		OSHome string
	}{OSHome: dir}

	_ = cfg

	files := discoverFiles(dir, []string{".go", ".sh", ".json", ".md"})
	if len(files) < 2 {
		t.Errorf("expected at least 2 files, got %d", len(files))
	}

	for _, path := range files {
		kind := classify(path)
		if kind == "" {
			t.Errorf("classify(%q) returned empty string", path)
		}
		id := safeID(kind + "." + rel(dir, path))
		if id == "" {
			t.Errorf("safeID produced empty ID for %q", path)
		}
	}
}
