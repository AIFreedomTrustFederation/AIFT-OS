package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExistsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0644)

	if !Exists(path) {
		t.Error("Exists should return true for existing file")
	}
}

func TestExistsDir(t *testing.T) {
	dir := t.TempDir()
	if !Exists(dir) {
		t.Error("Exists should return true for existing directory")
	}
}

func TestExistsMissing(t *testing.T) {
	if Exists("/nonexistent/path/xyz") {
		t.Error("Exists should return false for nonexistent path")
	}
}

func TestFileExistsTrue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0644)

	if !FileExists(path) {
		t.Error("FileExists should return true for regular file")
	}
}

func TestFileExistsDir(t *testing.T) {
	dir := t.TempDir()
	if FileExists(dir) {
		t.Error("FileExists should return false for directory")
	}
}

func TestFileExistsMissing(t *testing.T) {
	if FileExists("/nonexistent/path/xyz") {
		t.Error("FileExists should return false for nonexistent path")
	}
}

func TestDirExistsTrue(t *testing.T) {
	dir := t.TempDir()
	if !DirExists(dir) {
		t.Error("DirExists should return true for directory")
	}
}

func TestDirExistsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0644)

	if DirExists(path) {
		t.Error("DirExists should return false for regular file")
	}
}

func TestDirExistsMissing(t *testing.T) {
	if DirExists("/nonexistent/path/xyz") {
		t.Error("DirExists should return false for nonexistent path")
	}
}
