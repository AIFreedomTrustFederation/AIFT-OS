package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestBuiltinsAreSortedAndUnique(t *testing.T) {
	commands := Builtins()
	if len(commands) == 0 {
		t.Fatal("no built-in commands")
	}
	names := Names()
	if !sort.StringsAreSorted(names) {
		t.Fatalf("names are not sorted: %#v", names)
	}
	seen := map[string]bool{}
	for _, name := range names {
		if seen[name] {
			t.Fatalf("duplicate command: %s", name)
		}
		seen[name] = true
	}
}

func TestWritePersistsCLIRegistryAndReport(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}
	report := Report{Name: "test", Verified: true, Commands: []Command{{Name: "status", Description: "status"}}}
	if err := Write(cfg, report); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "registry", "cli", "cli-registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	var decoded Report
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Commands[0].Name != "status" {
		t.Fatalf("decoded report = %#v", decoded)
	}
	if _, err := os.Stat(filepath.Join(dir, "reports", "cli-registry.md")); err != nil {
		t.Fatal(err)
	}
}
