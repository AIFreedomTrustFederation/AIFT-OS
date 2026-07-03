package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type Command struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Usage       string               `json:"usage"`
	Aliases     []string             `json:"aliases,omitempty"`
	Status      string               `json:"status"`
	Handler     func([]string) error `json:"-"`
}

type Check struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type AIFTApp struct {
	ID          string         `json:"id"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Status      string         `json:"status,omitempty"`
	Runtime     string         `json:"runtime,omitempty"`
	Command     []string       `json:"command,omitempty"`
	WorkingDir  string         `json:"working_dir,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Source      string         `json:"source,omitempty"`
	Duplicate   bool           `json:"duplicate,omitempty"`
}

type appDiscoveryResult struct {
	Apps         []AIFTApp `json:"apps"`
	DuplicateIDs []string  `json:"duplicate_ids,omitempty"`
}

func main() {
	code := run(os.Args[1:])
	if code != 0 {
		os.Exit(code)
	}
}

func run(args []string) int {
	args = stripArgv0(args)
	cmds := commands()

	if len(args) == 0 {
		printHelp(cmds)
		return 0
	}

	if args[0] == "--" {
		args = args[1:]
	}

	if len(args) == 0 {
		printHelp(cmds)
		return 0
	}

	name := args[0]
	commandArgs := args[1:]

	cmd, ok := resolve(cmds, name)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", name)
		printHelp(cmds)
		return 2
	}

	if err := cmd.Handler(commandArgs); err != nil {
		fmt.Fprintf(os.Stderr, "command failed: %v\n", err)
		return 1
	}

	return 0
}

func stripArgv0(args []string) []string {
	if len(args) == 0 {
		return args
	}

	first := args[0]
	base := filepath.Base(first)

	if base == "aift" || base == "aift.exe" {
		return args[1:]
	}

	if strings.Contains(first, string(os.PathSeparator)) {
		if _, err := os.Stat(first); err == nil {
			return args[1:]
		}
	}

	return args
}

func commands() []Command {
	cmds := []Command{
		{"apps", "Discover and inspect truthful AIFT app manifests.", "aift apps <list|inspect|launch>", nil, "active", runApps},
		{"help", "Show available commands.", "aift help", []string{"--help", "-h"}, "active", runHelp},
		{"status", "Inspect the real local repository status.", "aift status", []string{"doctor"}, "active", runStatus},
		{"verify", "Run real local bootstrap verification checks.", "aift verify", []string{"check"}, "active", runVerify},
		{"registry", "Print command registry as JSON.", "aift registry", []string{"commands"}, "active", runRegistry},
		{"bootstrap", "Print federation bootstrap discovery JSON.", "aift bootstrap", nil, "active", runBootstrap},
		{"federation", "Federation command group; planned until real APIs are proven.", "aift federation", []string{"fed"}, "planned", planned("federation")},
		{"repo", "Repository command group; planned until real APIs are proven.", "aift repo", []string{"repos"}, "planned", planned("repo")},
		{"workflow", "Workflow command group; planned until real APIs are proven.", "aift workflow", []string{"flows"}, "planned", planned("workflow")},
	}

	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })
	return cmds
}

func resolve(cmds []Command, name string) (Command, bool) {
	for _, c := range cmds {
		if c.Name == name {
			return c, true
		}
		for _, a := range c.Aliases {
			if a == name {
				return c, true
			}
		}
	}
	return Command{}, false
}

func runHelp(args []string) error {
	printHelp(commands())
	return nil
}

func printHelp(cmds []Command) {
	fmt.Println("AIFT-OS CLI")
	fmt.Println()
	fmt.Println("Truthful local-first federation operator CLI.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  aift <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	for _, c := range cmds {
		fmt.Printf("  %-12s %-8s %s\n", c.Name, c.Status, c.Description)
	}
}

func runRegistry(args []string) error {
	return printJSON(map[string]any{
		"generated_at": time.Now().Format(time.RFC3339),
		"runtime":      runtime.Version(),
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
		"commands":     commands(),
	})
}

func runStatus(args []string) error {
	root, _ := os.Getwd()
	checks := collectChecks()

	return printJSON(map[string]any{
		"status":       aggregate(checks),
		"generated_at": time.Now().Format(time.RFC3339),
		"root":         root,
		"runtime":      runtime.Version(),
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
		"checks":       checks,
	})
}

func runVerify(args []string) error {
	checks := collectChecks()
	status := aggregate(checks)

	if err := printJSON(map[string]any{
		"status": status,
		"checks": checks,
	}); err != nil {
		return err
	}

	if status == "fail" {
		return fmt.Errorf("verification failed")
	}

	return nil
}

func runBootstrap(args []string) error {
	root, _ := os.Getwd()

	return printJSON(map[string]any{
		"generated_at": time.Now().Format(time.RFC3339),
		"root":         root,
		"discovery": map[string]any{
			"git":       exists(".git"),
			"go_mod":    exists("go.mod"),
			"package":   exists("package.json"),
			"registry":  exists("registry"),
			"apps":      exists("registry/apps"),
			"internal":  exists("internal"),
			"cmd_aift":  exists("cmd/aift/main.go"),
			"reports":   exists("reports"),
			"scripts":   exists("scripts"),
			"manifests": exists("manifests"),
			"workflows": exists(".github/workflows"),
		},
		"commands": commands(),
	})
}

func runApps(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: aift apps <list|inspect|launch>")
	}

	switch args[0] {
	case "list":
		if len(args) != 1 {
			return fmt.Errorf("usage: aift apps list")
		}
		return runAppsList()
	case "inspect":
		if len(args) != 2 {
			return fmt.Errorf("usage: aift apps inspect <id>")
		}
		return runAppsInspect(args[1])
	case "launch":
		return runAppsLaunch(args[1:])
	default:
		return fmt.Errorf("unknown apps subcommand: %s", args[0])
	}
}

func runAppsList() error {
	root, _ := os.Getwd()
	result, err := discoverApps(root)
	if err != nil {
		return err
	}
	return printJSON(map[string]any{
		"status":        "pass",
		"generated_at":  time.Now().Format(time.RFC3339),
		"root":          root,
		"apps":          result.Apps,
		"duplicate_ids": result.DuplicateIDs,
	})
}

func runAppsInspect(id string) error {
	root, _ := os.Getwd()
	result, err := discoverApps(root)
	if err != nil {
		return err
	}

	matches := appsByID(result.Apps, id)
	if len(matches) == 0 {
		return fmt.Errorf("app not found: %s", id)
	}
	if len(matches) > 1 {
		return printJSON(map[string]any{
			"status":  "fail",
			"reason":  "duplicate_app_id",
			"id":      id,
			"matches": matches,
		})
	}

	return printJSON(map[string]any{
		"status": "pass",
		"app":    matches[0],
	})
}

func runAppsLaunch(args []string) error {
	if len(args) != 2 || args[1] != "--plan" {
		return fmt.Errorf("usage: aift apps launch <id> --plan")
	}

	id := args[0]
	root, _ := os.Getwd()
	result, err := discoverApps(root)
	if err != nil {
		return err
	}

	matches := appsByID(result.Apps, id)
	if len(matches) == 0 {
		return fmt.Errorf("app not found: %s", id)
	}
	if len(matches) > 1 {
		return printJSON(map[string]any{
			"status":  "planned",
			"active":  false,
			"reason":  "duplicate_app_id",
			"id":      id,
			"matches": matches,
		})
	}

	app := matches[0]
	return printJSON(map[string]any{
		"status":       "planned",
		"active":       false,
		"id":           app.ID,
		"app":          app,
		"launch_plan":  "local app launch is intentionally not active until local verification is implemented",
		"verification": "planned",
	})
}

func discoverApps(root string) (appDiscoveryResult, error) {
	patterns := []string{
		filepath.Join(root, "registry", "apps", "*.json"),
		filepath.Join(root, ".aift", "apps", "*.json"),
		filepath.Join(root, "..", "*", ".aift", "apps", "*.json"),
	}

	seenPaths := map[string]bool{}
	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return appDiscoveryResult{}, fmt.Errorf("invalid app discovery pattern %q: %w", pattern, err)
		}
		sort.Strings(matches)
		for _, match := range matches {
			clean := filepath.Clean(match)
			if seenPaths[clean] {
				continue
			}
			seenPaths[clean] = true
			files = append(files, clean)
		}
	}

	var apps []AIFTApp
	idCounts := map[string]int{}
	for _, file := range files {
		app, err := readAppManifest(file)
		if err != nil {
			return appDiscoveryResult{}, err
		}
		app.Source = relativeOrClean(root, file)
		if app.Status == "" {
			app.Status = "discovered"
		}
		apps = append(apps, app)
		idCounts[app.ID]++
	}

	var duplicateIDs []string
	for id, count := range idCounts {
		if count > 1 {
			duplicateIDs = append(duplicateIDs, id)
		}
	}
	sort.Strings(duplicateIDs)

	if len(duplicateIDs) > 0 {
		duplicates := map[string]bool{}
		for _, id := range duplicateIDs {
			duplicates[id] = true
		}
		for i := range apps {
			if duplicates[apps[i].ID] {
				apps[i].Duplicate = true
			}
		}
	}

	sort.Slice(apps, func(i, j int) bool {
		if apps[i].ID == apps[j].ID {
			return apps[i].Source < apps[j].Source
		}
		return apps[i].ID < apps[j].ID
	})

	return appDiscoveryResult{Apps: apps, DuplicateIDs: duplicateIDs}, nil
}

func readAppManifest(path string) (AIFTApp, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AIFTApp{}, fmt.Errorf("read app manifest %s: %w", path, err)
	}

	var app AIFTApp
	if err := json.Unmarshal(data, &app); err != nil {
		return AIFTApp{}, fmt.Errorf("malformed app manifest %s: %w", path, err)
	}
	if strings.TrimSpace(app.ID) == "" {
		return AIFTApp{}, fmt.Errorf("invalid app manifest %s: missing id", path)
	}
	return app, nil
}

func appsByID(apps []AIFTApp, id string) []AIFTApp {
	var matches []AIFTApp
	for _, app := range apps {
		if app.ID == id {
			matches = append(matches, app)
		}
	}
	return matches
}

func relativeOrClean(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err == nil && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".." {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(filepath.Clean(path))
}

func collectChecks() []Check {
	checks := []Check{
		fileCheck("git-repository", ".git", "Local Git repository exists."),
		fileCheck("go-module", "go.mod", "Go module manifest exists."),
		fileCheck("cli-entrypoint", "cmd/aift/main.go", "AIFT CLI entrypoint exists."),
		fileCheck("registry-directory", "registry", "Federation registry directory exists."),
		fileCheck("apps-registry-directory", "registry/apps", "AIFT app registry directory exists."),
		fileCheck("internal-directory", "internal", "Internal package directory exists."),
		fileCheck("reports-directory", "reports", "Reports directory exists."),
		toolCheck("git-binary", "git"),
		toolCheck("go-binary", "go"),
	}

	checks = append(checks, commandCheck("go-build-cmd-aift", "go", "build", "./cmd/aift"))
	return checks
}

func fileCheck(name, path, detail string) Check {
	if exists(path) {
		return Check{name, "pass", detail}
	}
	return Check{name, "planned", "Missing: " + path}
}

func toolCheck(name, tool string) Check {
	if _, err := exec.LookPath(tool); err == nil {
		return Check{name, "pass", tool + " is available."}
	}
	return Check{name, "fail", tool + " is not available."}
}

func commandCheck(name string, cmd string, args ...string) Check {
	c := exec.Command(cmd, args...)
	out, err := c.CombinedOutput()
	if err != nil {
		return Check{name, "fail", strings.TrimSpace(string(out))}
	}
	return Check{name, "pass", strings.TrimSpace(string(out))}
}

func aggregate(checks []Check) string {
	status := "pass"
	for _, c := range checks {
		if c.Status == "fail" {
			return "fail"
		}
		if c.Status == "planned" {
			status = "partial"
		}
	}
	return status
}

func planned(name string) func([]string) error {
	return func(args []string) error {
		return printJSON(map[string]any{
			"command": name,
			"status":  "planned",
			"message": "Registered honestly but not yet wired to a proven internal implementation.",
		})
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
