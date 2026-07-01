package providers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestDefaultsIncludePlannedAIProvidersAsPlanned(t *testing.T) {
	found := false
	for _, provider := range Defaults() {
		if provider.Name == "ollama" {
			found = true
			if provider.Status != "planned" {
				t.Fatalf("ollama status = %q, want planned", provider.Status)
			}
		}
	}
	if !found {
		t.Fatal("ollama provider not found")
	}
}

func TestWriteRegistryWritesProviders(t *testing.T) {
	dir := t.TempDir()
	if err := WriteRegistry(config.Config{Root: dir, OSHome: dir}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "registry", "providers.json"))
	if err != nil {
		t.Fatal(err)
	}
	var providers []Provider
	if err := json.Unmarshal(data, &providers); err != nil {
		t.Fatal(err)
	}
	if len(providers) != len(Defaults()) {
		t.Fatalf("provider count = %d, want %d", len(providers), len(Defaults()))
	}
}
