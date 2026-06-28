package jsonfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Write marshals v as indented JSON and writes it to path, creating parent
// directories as needed. If announce is true, prints "Wrote <path>" to stdout.
func Write(path string, v any, announce bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return err
	}
	if announce {
		fmt.Println("Wrote", path)
	}
	return nil
}

// ReadPackageCommands reads package.json from repoPath and populates the
// commands map with "npm:<script>" -> "npm run <script>" entries.
func ReadPackageCommands(repoPath string, commands map[string]string) {
	data, err := os.ReadFile(filepath.Join(repoPath, "package.json"))
	if err != nil {
		return
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return
	}
	for name := range pkg.Scripts {
		commands["npm:"+name] = "npm run " + name
	}
}

// ReadNamedList reads a JSON file from .aift/<fileName> and returns the "name"
// values from the array at the given field key.
func ReadNamedList(repoPath, fileName, field string) []string {
	data, err := os.ReadFile(filepath.Join(repoPath, ".aift", fileName))
	if err != nil {
		return []string{}
	}
	var raw map[string][]map[string]string
	if json.Unmarshal(data, &raw) != nil {
		return []string{}
	}
	out := []string{}
	for _, item := range raw[field] {
		if item["name"] != "" {
			out = append(out, item["name"])
		}
	}
	return out
}
