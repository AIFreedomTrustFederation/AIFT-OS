package capability

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type Capability struct {
	Name      string `json:"name"`
	Command   string `json:"command"`
	Installed bool   `json:"installed"`
	Path      string `json:"path"`
}

type Report struct {
	Name         string       `json:"name"`
	Time         string       `json:"time"`
	Root         string       `json:"root"`
	OSHome       string       `json:"os_home"`
	Verified     bool         `json:"verified"`
	Capabilities []Capability `json:"capabilities"`
}

func Discover(cfg config.Config) Report {
	names := []string{
		"go",
		"git",
		"node",
		"npm",
		"pnpm",
		"bun",
		"python",
		"python3",
		"pip",
		"pip3",
		"cargo",
		"rustc",
		"make",
		"docker",
		"java",
		"mvn",
		"gradle",
		"zig",
	}

	report := Report{
		Name:     "AIFT Capability Discovery",
		Time:     time.Now().Format(time.RFC3339),
		Root:     cfg.Root,
		OSHome:   cfg.OSHome,
		Verified: true,
	}

	for _, name := range names {
		path, err := exec.LookPath(name)
		capability := Capability{
			Name:      name,
			Command:   name,
			Installed: err == nil,
			Path:      path,
		}
		report.Capabilities = append(report.Capabilities, capability)
	}

	return report
}

func Run(cfg config.Config) error {
	report := Discover(cfg)
	if err := Write(cfg, report); err != nil {
		return err
	}

	fmt.Println("AIFT Capability Discovery")
	for _, cap := range report.Capabilities {
		status := "missing"
		if cap.Installed {
			status = "installed"
		}
		fmt.Printf("%-10s %s\n", cap.Name, status)
	}
	return nil
}

func Write(cfg config.Config, report Report) error {
	outDir := filepath.Join(cfg.OSHome, "registry", "capabilities")
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

	jsonPath := filepath.Join(outDir, "capabilities.json")
	mdPath := filepath.Join(reportDir, "capabilities.md")

	if err := os.WriteFile(jsonPath, append(b, '\n'), 0644); err != nil {
		return err
	}

	md := "# AIFT Capability Discovery Report\n\n"
	for _, cap := range report.Capabilities {
		status := "missing"
		if cap.Installed {
			status = "installed"
		}
		md += "- " + cap.Name + ": " + status + "\n"
	}

	return os.WriteFile(mdPath, []byte(md), 0644)
}

func Has(report Report, name string) bool {
	for _, cap := range report.Capabilities {
		if cap.Name == name && cap.Installed {
			return true
		}
	}
	return false
}
