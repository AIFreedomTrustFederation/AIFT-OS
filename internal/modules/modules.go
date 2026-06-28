package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type ModuleManifest struct {
	ID             string            `json:"id"`
	Repo           string            `json:"repo"`
	Name           string            `json:"name"`
	Version        string            `json:"version"`
	Status         string            `json:"status"`
	Kind           string            `json:"kind"`
	Description    string            `json:"description"`
	DependsOn      []string          `json:"dependsOn"`
	Provides       []string          `json:"provides"`
	Consumes       []string          `json:"consumes"`
	Publishes      []string          `json:"publishes"`
	Commands       map[string]string `json:"commands"`
	Services       []string          `json:"services"`
	Capabilities   []string          `json:"capabilities"`
	Docs           []string          `json:"docs"`
	HealthChecks   []string          `json:"healthChecks"`
	MigrationLevel string            `json:"migrationLevel"`
	Evidence       []string          `json:"evidence"`
	GeneratedAt    string            `json:"generatedAt"`
}

type Registry struct {
	GeneratedAt string           `json:"generatedAt"`
	Modules     []ModuleManifest `json:"modules"`
}

func InitAll(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}
	for _, repo := range repos {
		if err := InitRepo(repo.Name, repo.Path); err != nil {
			return err
		}
	}
	return Scan(cfg)
}

func InitRepo(name, repoPath string) error {
	dir := filepath.Join(repoPath, ".aift")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, "module.json")
	if fileExists(path) {
		return nil
	}
	manifest := BuildRepoManifest(name, repoPath)
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func Scan(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	reg := Registry{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Modules:     []ModuleManifest{},
	}

	for _, repo := range repos {
		manifest := BuildRepoManifest(repo.Name, repo.Path)
		reg.Modules = append(reg.Modules, manifest)

		data, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(repo.Path, ".aift"), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(repo.Path, ".aift", "module.json"), append(data, '\n'), 0644); err != nil {
			return err
		}
	}

	sort.Slice(reg.Modules, func(i, j int) bool {
		return reg.Modules[i].ID < reg.Modules[j].ID
	})

	if err := writeRegistry(cfg, reg); err != nil {
		return err
	}
	if err := writeReport(cfg, reg); err != nil {
		return err
	}

	return events.Emit(cfg, "modules.scan", "modules", "kernel modules scanned", map[string]string{
		"count": fmt.Sprint(len(reg.Modules)),
	})
}

func BuildRepoManifest(name, repoPath string) ModuleManifest {
	commands := map[string]string{}
	evidence := []string{".git"}
	provides := []string{}
	consumes := []string{}
	publishes := []string{}
	docs := []string{}
	health := []string{}
	capabilities := []string{}
	services := []string{}

	if fileExists(filepath.Join(repoPath, "README.md")) {
		docs = append(docs, "README.md")
		evidence = append(evidence, "README.md")
	}
	if dirExists(filepath.Join(repoPath, "docs")) {
		docs = append(docs, "docs/")
		evidence = append(evidence, "docs/")
	}
	if fileExists(filepath.Join(repoPath, "package.json")) {
		evidence = append(evidence, "package.json")
		readPackageCommands(repoPath, commands)
		provides = append(provides, "node.package")
	}
	if fileExists(filepath.Join(repoPath, "go.mod")) {
		evidence = append(evidence, "go.mod")
		commands["go:test"] = "go test ./..."
		commands["go:build"] = "go build ./..."
		provides = append(provides, "go.module")
	}
	if fileExists(filepath.Join(repoPath, "Cargo.toml")) {
		evidence = append(evidence, "Cargo.toml")
		commands["cargo:test"] = "cargo test"
		commands["cargo:build"] = "cargo build"
		provides = append(provides, "rust.crate")
	}
	if fileExists(filepath.Join(repoPath, ".aift", "manual.json")) {
		provides = append(provides, "manual.contract")
		docs = append(docs, ".aift/manual.json")
		evidence = append(evidence, ".aift/manual.json")
	}
	if fileExists(filepath.Join(repoPath, ".aift", "capabilities.json")) {
		caps := readNamedList(repoPath, "capabilities.json", "capabilities")
		capabilities = append(capabilities, caps...)
		provides = append(provides, caps...)
		evidence = append(evidence, ".aift/capabilities.json")
	}
	if fileExists(filepath.Join(repoPath, ".aift", "services.json")) {
		services = append(services, readNamedList(repoPath, "services.json", "services")...)
		provides = append(provides, "service.contract")
		evidence = append(evidence, ".aift/services.json")
	}
	if fileExists(filepath.Join(repoPath, ".aift", "commands", "verify.sh")) {
		commands["aift:verify"] = "sh .aift/commands/verify.sh"
		health = append(health, ".aift/commands/verify.sh")
	}

	status := "detected"
	if len(capabilities) == 0 && len(services) == 0 && len(commands) == 0 {
		status = "planned"
	}
	if contains(capabilities, "verify") || commands["aift:verify"] != "" {
		status = "ready"
	}

	return ModuleManifest{
		ID:             "repo." + name,
		Repo:           name,
		Name:           name,
		Version:        "0.1.0",
		Status:         status,
		Kind:           inferKind(name, repoPath),
		Description:    "Auto-discovered federation kernel module for " + name,
		DependsOn:      []string{},
		Provides:       unique(provides),
		Consumes:       unique(consumes),
		Publishes:      unique(publishes),
		Commands:       commands,
		Services:       unique(services),
		Capabilities:   unique(capabilities),
		Docs:           unique(docs),
		HealthChecks:   unique(health),
		MigrationLevel: "phase-17",
		Evidence:       unique(evidence),
		GeneratedAt:    time.Now().Format(time.RFC3339),
	}
}

func List(cfg config.Config) error {
	reg, err := loadOrScan(cfg)
	if err != nil {
		return err
	}
	fmt.Printf("%-36s %-12s %-18s %s\n", "MODULE", "STATUS", "KIND", "REPO")
	for _, module := range reg.Modules {
		fmt.Printf("%-36s %-12s %-18s %s\n", module.ID, module.Status, module.Kind, module.Repo)
	}
	return nil
}

func Repo(cfg config.Config, name string) error {
	reg, err := loadOrScan(cfg)
	if err != nil {
		return err
	}
	for _, module := range reg.Modules {
		if module.Repo == name || module.ID == name {
			data, err := json.MarshalIndent(module, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}
	}
	return fmt.Errorf("module not found: %s", name)
}

func Report(cfg config.Config) error {
	path := filepath.Join(cfg.OSHome, "reports", "modules.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if err := Scan(cfg); err != nil {
			return err
		}
		data, err = os.ReadFile(path)
		if err != nil {
			return err
		}
	}
	fmt.Print(string(data))
	return nil
}

func writeRegistry(cfg config.Config, reg Registry) error {
	out := filepath.Join(cfg.OSHome, "registry", "modules.json")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(out, append(data, '\n'), 0644)
}

func writeReport(cfg config.Config, reg Registry) error {
	out := filepath.Join(cfg.OSHome, "reports", "modules.md")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# Federation Kernel Modules\n\n")
	b.WriteString("Every repository is discoverable as a kernel module when evidence exists on disk.\n\n")
	b.WriteString("| Module | Repo | Kind | Status | Provides | Commands |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, module := range reg.Modules {
		b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | `%s` | `%d` |\n",
			module.ID, module.Repo, module.Kind, module.Status, strings.Join(module.Provides, ", "), len(module.Commands)))
	}
	return os.WriteFile(out, []byte(b.String()), 0644)
}

func loadOrScan(cfg config.Config) (Registry, error) {
	path := filepath.Join(cfg.OSHome, "registry", "modules.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if err := Scan(cfg); err != nil {
			return Registry{}, err
		}
		data, err = os.ReadFile(path)
		if err != nil {
			return Registry{}, err
		}
	}
	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return Registry{}, err
	}
	return reg, nil
}

func readPackageCommands(repoPath string, commands map[string]string) {
	data, err := os.ReadFile(filepath.Join(repoPath, "package.json"))
	if err != nil {
		return
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return
	}
	for name := range pkg.Scripts {
		commands["npm:"+name] = "npm run " + name
	}
}

func readNamedList(repoPath, fileName, field string) []string {
	data, err := os.ReadFile(filepath.Join(repoPath, ".aift", fileName))
	if err != nil {
		return []string{}
	}
	var raw map[string][]map[string]string
	if json.Unmarshal(data, &raw) != nil {
		return []string{}
	}
	out := []string{}
	for _, item := range raw[field] {
		if item["name"] != "" {
			out = append(out, item["name"])
		}
	}
	return out
}

func inferKind(name, repoPath string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "os"):
		return "kernel"
	case strings.Contains(lower, "forge"):
		return "forge"
	case strings.Contains(lower, "book"):
		return "publishing"
	case strings.Contains(lower, "www") || strings.Contains(lower, "github.io"):
		return "website"
	case fileExists(filepath.Join(repoPath, "package.json")):
		return "node-app"
	case fileExists(filepath.Join(repoPath, "go.mod")):
		return "go-module"
	case fileExists(filepath.Join(repoPath, "Cargo.toml")):
		return "rust-crate"
	default:
		return "repository"
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func contains(items []string, wanted string) bool {
	for _, item := range items {
		if item == wanted {
			return true
		}
	}
	return false
}

func unique(items []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, item := range items {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}
