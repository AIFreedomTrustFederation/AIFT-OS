package integration

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHelpListsEachRegistryCommandOnce(t *testing.T) {
	helpOut, err := run(t, "help")
	if err != nil {
		t.Fatalf("help failed: %v\n%s", err, helpOut)
	}

	registryOut, err := run(t, "registry")
	if err != nil {
		t.Fatalf("registry failed: %v\n%s", err, registryOut)
	}

	var registry struct {
		Commands []struct {
			Name string `json:"name"`
		} `json:"commands"`
	}
	if err := json.Unmarshal([]byte(registryOut), &registry); err != nil {
		t.Fatalf("registry should be JSON: %v", err)
	}

	counts := map[string]int{}
	for _, line := range strings.Split(helpOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "Commands:" || strings.HasPrefix(line, "AIFT-OS") || strings.HasPrefix(line, "Truthful") || strings.HasPrefix(line, "Usage:") || strings.HasPrefix(line, "aift ") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) > 0 {
			counts[parts[0]]++
		}
	}

	for _, command := range registry.Commands {
		if counts[command.Name] != 1 {
			t.Fatalf("help lists %q %d times, want once", command.Name, counts[command.Name])
		}
	}
}
