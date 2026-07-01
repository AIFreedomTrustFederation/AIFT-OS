package providerregistry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestMatchDetectsKnownProviderFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	provider, ok := Match(dir)
	if !ok {
		t.Fatal("Match did not detect go provider")
	}
	if provider.Name != "go" || provider.TestCommand != "go test ./..." {
		t.Fatalf("provider = %#v", provider)
	}
}

func TestWriteCreatesJSONAndMarkdownReports(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Root: dir, OSHome: dir}
	report := Report{Name: "test", Verified: true, Providers: Builtins()[:1]}
	if err := Write(cfg, report); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		filepath.Join(dir, "registry", "providers", "provider-registry.json"),
		filepath.Join(dir, "reports", "provider-registry.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s: %v", path, err)
		}
	}
}
