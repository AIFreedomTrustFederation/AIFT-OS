package builder

import (
	"encoding/json"
	"fmt"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/lifecycle"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/scheduler"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/compiler"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type Step struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type Report struct {
	Name     string   `json:"name"`
	Time     string   `json:"time"`
	Root     string   `json:"root"`
	OSHome   string   `json:"os_home"`
	Verified bool     `json:"verified"`
	Steps    []Step   `json:"steps"`
	Blocked  []string `json:"blocked"`
}

func Run(cfg config.Config) error {
	report := Report{
		Name:     "AIFT Build Orchestrator",
		Time:     time.Now().Format(time.RFC3339),
		Root:     cfg.Root,
		OSHome:   cfg.OSHome,
		Verified: true,
	}

	add := func(name string, err error) {
		status := "pass"
		if err != nil {
			status = "blocked"
			report.Verified = false
			report.Blocked = append(report.Blocked, name+": "+err.Error())
		}
		report.Steps = append(report.Steps, Step{Name: name, Status: status})
	}

	add("repository compiler", compiler.Run(cfg))
	add("federation lifecycle", lifecycle.Run(cfg))
	add("federation scheduler", scheduler.Run(cfg))
	add("doctor", run(cfg.OSHome, "aift", "doctor"))
	add("verify", run(cfg.OSHome, "aift", "verify"))
	add("go test", run(cfg.OSHome, "go", "test", "./..."))
	add("go build", run(cfg.OSHome, "go", "build", "-o", filepath.Join(os.Getenv("HOME"), ".local", "bin", "aift"), "./cmd/aift"))

	if err := writeReport(cfg, report); err != nil {
		return err
	}

	fmt.Println("AIFT Build Orchestrator")
	fmt.Println("verified:", report.Verified)

	for _, step := range report.Steps {
		fmt.Printf("%-25s %s\n", step.Name, step.Status)
	}

	if !report.Verified {
		return fmt.Errorf("build completed with blocked work")
	}

	return nil
}

func run(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func writeReport(cfg config.Config, report Report) error {
	buildDir := filepath.Join(cfg.OSHome, "registry", "builds")
	reportDir := filepath.Join(cfg.OSHome, "reports")

	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	jsonPath := filepath.Join(buildDir, "build-report.json")
	mdPath := filepath.Join(reportDir, "build-report.md")

	if err := os.WriteFile(jsonPath, append(b, '\n'), 0644); err != nil {
		return err
	}

	md := "# AIFT Build Report\n\n"
	md += fmt.Sprintf("Verified: %v\n\n", report.Verified)
	md += "## Steps\n\n"

	for _, step := range report.Steps {
		md += "- " + step.Name + ": " + step.Status + "\n"
	}

	if len(report.Blocked) > 0 {
		md += "\n## Blocked\n\n"
		for _, item := range report.Blocked {
			md += "- " + item + "\n"
		}
	}

	return os.WriteFile(mdPath, []byte(md), 0644)
}
