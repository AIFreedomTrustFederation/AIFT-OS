package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Root   string
	OSHome string
}

func Load() Config {
	home, _ := os.UserHomeDir()

	root := os.Getenv("AIFT_ROOT")
	if root == "" {
		root = filepath.Join(home, "AIFT")
	}

	osHome := os.Getenv("AIFT_OS_HOME")
	if osHome == "" {
		osHome = filepath.Join(root, "AIFT-OS")
	}

	return Config{Root: root, OSHome: osHome}
}
