package kernelruntime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/discoveryengine"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/eventbus"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/jsonfile"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernelregistry"
)

type BootStep struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
	StartedAt   string `json:"startedAt"`
	FinishedAt  string `json:"finishedAt"`
}

type BootReport struct {
	SchemaVersion string     `json:"schemaVersion"`
	GeneratedAt   string     `json:"generatedAt"`
	Status        string     `json:"status"`
	OSHome        string     `json:"osHome"`
	Root          string     `json:"root"`
	Steps         []BootStep `json:"steps"`
	Summary       Summary    `json:"summary"`
}

type Summary struct {
	DiscoveryObjects int `json:"discoveryObjects"`
	RegistryObjects  int `json:"registryObjects"`
	EventCount       int `json:"eventCount"`
}

func Boot(cfg config.Config) error {
	report := BootReport{
		SchemaVersion: "aift.kernel.boot.v1",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Status:        "booting",
		OSHome:        cfg.OSHome,
		Root:          cfg.Root,
		Steps:         []BootStep{},
	}

	fmt.Println("AIFT-OS Kernel v0.1")
	fmt.Println()

	if err := runStep(&report, "configuration", "Configuration loaded from runtime config.", func() error {
		return nil
	}); err != nil {
		return finish(cfg, report, err)
	}

	if err := runStep(&report, "discovery", "Discovering federation repositories and runtime evidence.", func() error {
		return discoveryengine.Scan(cfg)
	}); err != nil {
		return finish(cfg, report, err)
	}

	if err := runStep(&report, "kernel-registry", "Building kernel registry from discovered evidence.", func() error {
		return kernelregistry.Scan(cfg)
	}); err != nil {
		return finish(cfg, report, err)
	}

	if err := runStep(&report, "event-bus", "Publishing kernel boot event.", func() error {
		return eventbus.Publish(cfg, "kernel.started", "kernel", "kernelruntime", "AIFT kernel boot sequence completed", map[string]string{
			"osHome": cfg.OSHome,
			"root":   cfg.Root,
		})
	}); err != nil {
		return finish(cfg, report, err)
	}

	discovery, _ := discoveryengine.LoadOrBuild(cfg)
	registry, _ := kernelregistry.LoadOrBuild(cfg)
	events, _ := eventbus.Load(cfg)

	report.Summary = Summary{
		DiscoveryObjects: len(discovery.Objects),
		RegistryObjects:  len(registry.Objects),
		EventCount:       len(events),
	}
	report.Status = "ready"

	if err := WriteReport(cfg, report); err != nil {
		return err
	}
	if err := WriteJSON(cfg, report); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Kernel Ready.")
	fmt.Printf("Discovery objects: %d\n", report.Summary.DiscoveryObjects)
	fmt.Printf("Registry objects:  %d\n", report.Summary.RegistryObjects)
	fmt.Printf("Events:            %d\n", report.Summary.EventCount)

	return nil
}

func Status(cfg config.Config) error {
	path := filepath.Join(cfg.OSHome, "registry", "kernel-boot.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("kernel has not booted yet: %w", err)
	}
	fmt.Print(string(data))
	return nil
}

func Report(cfg config.Config) error {
	path := filepath.Join(cfg.OSHome, "reports", "kernel-boot.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("kernel boot report missing: %w", err)
	}
	fmt.Print(string(data))
	return nil
}

func runStep(report *BootReport, name string, description string, fn func() error) error {
	start := time.Now().UTC().Format(time.RFC3339)
	step := BootStep{
		Name:        name,
		Status:      "running",
		Description: description,
		StartedAt:   start,
	}

	fmt.Printf("→ %s...\n", name)

	if err := fn(); err != nil {
		step.Status = "failed"
		step.FinishedAt = time.Now().UTC().Format(time.RFC3339)
		report.Steps = append(report.Steps, step)
		fmt.Printf("✗ %s failed\n", name)
		return err
	}

	step.Status = "ready"
	step.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	report.Steps = append(report.Steps, step)
	fmt.Printf("✓ %s ready\n", name)
	return nil
}

func finish(cfg config.Config, report BootReport, err error) error {
	report.Status = "failed"
	_ = WriteReport(cfg, report)
	_ = WriteJSON(cfg, report)
	return err
}

func WriteJSON(cfg config.Config, report BootReport) error {
	return jsonfile.Write(filepath.Join(cfg.OSHome, "registry", "kernel-boot.json"), report, false)
}

func WriteReport(cfg config.Config, report BootReport) error {
	out := filepath.Join(cfg.OSHome, "reports", "kernel-boot.md")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# AIFT Kernel Boot Report\n\n")
	b.WriteString(fmt.Sprintf("- Status: `%s`\n", report.Status))
	b.WriteString(fmt.Sprintf("- Generated: `%s`\n", report.GeneratedAt))
	b.WriteString(fmt.Sprintf("- OS Home: `%s`\n", report.OSHome))
	b.WriteString(fmt.Sprintf("- Root: `%s`\n\n", report.Root))

	b.WriteString("## Steps\n\n")
	b.WriteString("| Step | Status | Description |\n")
	b.WriteString("|---|---|---|\n")
	for _, step := range report.Steps {
		b.WriteString(fmt.Sprintf("| `%s` | `%s` | %s |\n", step.Name, step.Status, escape(step.Description)))
	}

	b.WriteString("\n## Summary\n\n")
	b.WriteString(fmt.Sprintf("- Discovery objects: `%d`\n", report.Summary.DiscoveryObjects))
	b.WriteString(fmt.Sprintf("- Registry objects: `%d`\n", report.Summary.RegistryObjects))
	b.WriteString(fmt.Sprintf("- Events: `%d`\n", report.Summary.EventCount))

	return os.WriteFile(out, []byte(b.String()), 0644)
}

func escape(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}
