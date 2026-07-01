package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type Command struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Report struct {
	Name     string    `json:"name"`
	Time     string    `json:"time"`
	Verified bool      `json:"verified"`
	Commands []Command `json:"commands"`
}

func Builtins() []Command {
	names := []string{
		"help",
		"version",
		"doctor",
		"status",
		"manifest",
		"registry",
		"dashboard",
		"deps",
		"plugins",
		"providers",
		"events",
		"services",
		"start",
		"tick",
		"serve",
		"daemon",
		"sync",
		"federation",
		"repo",
		"workflow",
		"intelligence",
		"manual",
		"graph",
		"service-contracts",
		"plan",
		"modules",
		"kernel-registry",
		"discovery",
		"event-bus",
		"patch-engine",
		"kernel",
		"runtime",
		"capabilities",
		"operator",
		"scheduler",
		"ai",
		"compile",
		"compiler",
		"provider-registry",
		"capability",
		"lifecycle",
		"federation-build",
		"build",
		"verify",
	}

	seen := map[string]bool{}
	var commands []Command

	for _, name := range names {
		if seen[name] {
			continue
		}
		seen[name] = true
		commands = append(commands, Command{Name: name, Description: "registered AIFT command"})
	}

	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	return commands
}

func Names() []string {
	var names []string
	for _, command := range Builtins() {
		names = append(names, command.Name)
	}
	return names
}

func PrintHelp() {
	fmt.Println("AIFT-OS Federation Control Plane")
	fmt.Println("")
	fmt.Println("Commands:")

	for _, command := range Builtins() {
		fmt.Println("  " + command.Name)
	}
}

func Run(cfg config.Config) error {
	report := Report{
		Name:     "AIFT CLI Registry",
		Time:     time.Now().Format(time.RFC3339),
		Verified: true,
		Commands: Builtins(),
	}

	return Write(cfg, report)
}

func Write(cfg config.Config, report Report) error {
	outDir := filepath.Join(cfg.OSHome, "registry", "cli")
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

	if err := os.WriteFile(filepath.Join(outDir, "cli-registry.json"), append(b, '\n'), 0644); err != nil {
		return err
	}

	md := "# AIFT CLI Registry Report\n\n"
	for _, command := range report.Commands {
		md += "- " + command.Name + "\n"
	}

	return os.WriteFile(filepath.Join(reportDir, "cli-registry.md"), []byte(md), 0644)
}
