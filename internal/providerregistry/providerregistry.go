package providerregistry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type Provider struct {
	Name                 string   `json:"name"`
	Runtime              string   `json:"runtime"`
	DetectionFiles       []string `json:"detection_files"`
	RequiredCapabilities []string `json:"required_capabilities"`
	BuildCommand         string   `json:"build_command"`
	TestCommand          string   `json:"test_command"`
	SupportsSync         bool     `json:"supports_sync"`
	SupportsAsync        bool     `json:"supports_async"`
}

type Report struct {
	Name      string     `json:"name"`
	Time      string     `json:"time"`
	Verified  bool       `json:"verified"`
	Providers []Provider `json:"providers"`
}

func Builtins() []Provider {
	return []Provider{
		{
			Name:                 "go",
			Runtime:              "go",
			DetectionFiles:       []string{"go.mod"},
			RequiredCapabilities: []string{"go"},
			BuildCommand:         "go build ./...",
			TestCommand:          "go test ./...",
			SupportsSync:         true,
			SupportsAsync:        true,
		},
		{
			Name:                 "node-pnpm",
			Runtime:              "node",
			DetectionFiles:       []string{"pnpm-lock.yaml"},
			RequiredCapabilities: []string{"node", "pnpm"},
			BuildCommand:         "pnpm run build",
			TestCommand:          "pnpm test",
			SupportsSync:         true,
			SupportsAsync:        true,
		},
		{
			Name:                 "node-npm",
			Runtime:              "node",
			DetectionFiles:       []string{"package.json"},
			RequiredCapabilities: []string{"node", "npm"},
			BuildCommand:         "npm run build",
			TestCommand:          "npm test",
			SupportsSync:         true,
			SupportsAsync:        true,
		},
		{
			Name:                 "python",
			Runtime:              "python",
			DetectionFiles:       []string{"pyproject.toml", "requirements.txt"},
			RequiredCapabilities: []string{"python"},
			BuildCommand:         "python -m compileall .",
			TestCommand:          "python -m pytest",
			SupportsSync:         true,
			SupportsAsync:        true,
		},
		{
			Name:                 "rust",
			Runtime:              "rust",
			DetectionFiles:       []string{"Cargo.toml"},
			RequiredCapabilities: []string{"cargo"},
			BuildCommand:         "cargo build",
			TestCommand:          "cargo test",
			SupportsSync:         true,
			SupportsAsync:        true,
		},
		{
			Name:                 "make",
			Runtime:              "make",
			DetectionFiles:       []string{"Makefile"},
			RequiredCapabilities: []string{"make"},
			BuildCommand:         "make",
			TestCommand:          "make test",
			SupportsSync:         true,
			SupportsAsync:        true,
		},
	}
}

func Match(path string) (Provider, bool) {
	for _, provider := range Builtins() {
		for _, file := range provider.DetectionFiles {
			if exists(filepath.Join(path, file)) {
				return provider, true
			}
		}
	}
	return Provider{}, false
}

func Run(cfg config.Config) error {
	report := Report{
		Name:      "AIFT Provider Registry",
		Time:      time.Now().Format(time.RFC3339),
		Verified:  true,
		Providers: Builtins(),
	}

	if err := Write(cfg, report); err != nil {
		return err
	}

	fmt.Println("AIFT Provider Registry")
	fmt.Println("providers:", len(report.Providers))
	return nil
}

func Write(cfg config.Config, report Report) error {
	outDir := filepath.Join(cfg.OSHome, "registry", "providers")
	reportDir := filepath.Join(cfg.OSHome, "reports")

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	jsonPath := filepath.Join(outDir, "provider-registry.json")
	mdPath := filepath.Join(reportDir, "provider-registry.md")

	if err := os.WriteFile(jsonPath, append(b, '\n'), 0644); err != nil {
		return err
	}

	md := "# AIFT Provider Registry Report\n\n"
	for _, provider := range report.Providers {
		md += "- " + provider.Name + " | " + provider.Runtime + "\n"
	}

	return os.WriteFile(mdPath, []byte(md), 0644)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
