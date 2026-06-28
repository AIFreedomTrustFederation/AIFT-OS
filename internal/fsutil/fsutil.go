package fsutil

import "os"

// Exists returns true if the path exists (file or directory).
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FileExists returns true if path exists and is a regular file.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// DirExists returns true if path exists and is a directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
