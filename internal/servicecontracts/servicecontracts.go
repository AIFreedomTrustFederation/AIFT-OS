package servicecontracts

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
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/fsutil"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/jsonfile"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Service struct {
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	Status      string   `json:"status"`
	Version     string   `json:"version"`
	Owner       string   `json:"owner"`
	Provides    []string `json:"provides"`
	Requires    []string `json:"requires"`
	Events      []string `json:"events"`
	Health      string   `json:"health,omitempty"`
	Start       string   `json:"start,omitempty"`
	Stop        string   `json:"stop,omitempty"`
	Evidence    string   `json:"evidence"`
	Description string   `json:"description"`
}

type Contract struct {
	Repo        string    `json:"repo"`
	Services    []Service `json:"services"`
	GeneratedAt string    `json:"generatedAt"`
}

type Registry struct {
	GeneratedAt string          `json:"generatedAt"`
	Contracts   []Contract      `json:"contracts"`
	Services    []ServiceRecord `json:"services"`
}

type ServiceRecord struct {
	Repo     string `json:"repo"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Status   string `json:"status"`
	Version  string `json:"version"`
	Evidence string `json:"evidence"`
}

func InitAll(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	for _, r := range repos {
		if err := InitRepo(r.Name, r.Path); err != nil {
			return err
		}
	}

	return Scan(cfg)
}

func InitRepo(name string, repoPath string) error {
	dir := filepath.Join(repoPath, ".aift")
	if err := os.MkdirAll(filepath.Join(dir, "services"), 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, "services.json")
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	contract := Contract{
		Repo: name,
		Services: []Service{
			{
				Name:        name + ".service",
				Kind:        inferKind(name, repoPath),
				Status:      "planned",
				Version:     "0.1.0",
				Owner:       name,
				Provides:    []string{},
				Requires:    []string{},
				Events:      []string{"repo.changed", "capability.changed", "manual.changed"},
				Evidence:    "default service contract",
				Description: "Default planned federation service contract for " + name,
			},
		},
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(contract, "", "  ")
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

	var contracts []Contract
	var services []ServiceRecord

	for _, r := range repos {
		c, ok := readContract(r.Name, r.Path)
		if !ok {
			continue
		}

		for i := range c.Services {
			if c.Services[i].Status == "" {
				c.Services[i].Status = "planned"
			}
			if c.Services[i].Version == "" {
				c.Services[i].Version = "0.1.0"
			}
			if c.Services[i].Owner == "" {
				c.Services[i].Owner = c.Repo
			}
			if c.Services[i].Evidence == "" {
				c.Services[i].Evidence = ".aift/services.json"
			}

			services = append(services, ServiceRecord{
				Repo:     c.Repo,
				Name:     c.Services[i].Name,
				Kind:     c.Services[i].Kind,
				Status:   c.Services[i].Status,
				Version:  c.Services[i].Version,
				Evidence: c.Services[i].Evidence,
			})
		}

		contracts = append(contracts, c)
	}

	sort.Slice(services, func(i, j int) bool {
		if services[i].Repo == services[j].Repo {
			return services[i].Name < services[j].Name
		}
		return services[i].Repo < services[j].Repo
	})

	reg := Registry{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Contracts:   contracts,
		Services:    services,
	}

	if err := writeRegistry(cfg, reg); err != nil {
		return err
	}
	if err := writeReport(cfg, reg); err != nil {
		return err
	}

	return events.Emit(cfg, "services.scan", "servicecontracts", "service contracts scanned", map[string]string{
		"services": fmt.Sprint(len(services)),
	})
}

func List(cfg config.Config) error {
	reg, err := loadOrScan(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("%-30s %-34s %-16s %-12s %s\n", "REPO", "SERVICE", "KIND", "STATUS", "VERSION")
	for _, s := range reg.Services {
		fmt.Printf("%-30s %-34s %-16s %-12s %s\n", s.Repo, s.Name, s.Kind, s.Status, s.Version)
	}
	return nil
}

func Repo(cfg config.Config, name string) error {
	reg, err := loadOrScan(cfg)
	if err != nil {
		return err
	}

	found := false
	for _, c := range reg.Contracts {
		if c.Repo != name {
			continue
		}
		found = true
		fmt.Println("Repository:", c.Repo)
		for _, svc := range c.Services {
			fmt.Println()
			fmt.Println("Service:", svc.Name)
			fmt.Println("Kind:", svc.Kind)
			fmt.Println("Status:", svc.Status)
			fmt.Println("Version:", svc.Version)
			fmt.Println("Owner:", svc.Owner)
			fmt.Println("Provides:", strings.Join(svc.Provides, ", "))
			fmt.Println("Requires:", strings.Join(svc.Requires, ", "))
			fmt.Println("Events:", strings.Join(svc.Events, ", "))
			fmt.Println("Health:", svc.Health)
			fmt.Println("Start:", svc.Start)
			fmt.Println("Stop:", svc.Stop)
			fmt.Println("Evidence:", svc.Evidence)
		}
	}

	if !found {
		return fmt.Errorf("repository not found or no service contract: %s", name)
	}
	return nil
}

func Report(cfg config.Config) error {
	path := filepath.Join(cfg.OSHome, "reports", "service-contracts.md")
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

func readContract(name string, repoPath string) (Contract, bool) {
	path := filepath.Join(repoPath, ".aift", "services.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, false
	}

	var c Contract
	if json.Unmarshal(data, &c) != nil {
		return Contract{}, false
	}
	if c.Repo == "" {
		c.Repo = name
	}
	return c, true
}

func writeRegistry(cfg config.Config, reg Registry) error {
	return jsonfile.Write(filepath.Join(cfg.OSHome, "registry", "service-contracts.json"), reg, true)
}

func writeReport(cfg config.Config, reg Registry) error {
	out := filepath.Join(cfg.OSHome, "reports", "service-contracts.md")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# Federation Service Contracts\n\n")
	b.WriteString("Service contracts declare what each repo provides, requires, and can eventually run.\n\n")
	b.WriteString("AIFT-OS records these contracts truthfully. Planned services are not executed.\n\n")
	b.WriteString("| Repository | Service | Kind | Status | Version | Evidence |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, s := range reg.Services {
		b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | `%s` | %s |\n", s.Repo, s.Name, s.Kind, s.Status, s.Version, s.Evidence))
	}

	return os.WriteFile(out, []byte(b.String()), 0644)
}

func loadOrScan(cfg config.Config) (Registry, error) {
	path := filepath.Join(cfg.OSHome, "registry", "service-contracts.json")
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

func inferKind(name string, repoPath string) string {
	lower := strings.ToLower(name)

	switch {
	case strings.Contains(lower, "aift-os"):
		return "control-plane"
	case strings.Contains(lower, "forge"):
		return "forge"
	case strings.Contains(lower, "booksmith"):
		return "publishing"
	case strings.Contains(lower, "vps"):
		return "infrastructure"
	case strings.Contains(lower, "www") || strings.Contains(lower, "github.io"):
		return "website"
	case fsutil.Exists(filepath.Join(repoPath, "package.json")):
		return "web-app"
	case fsutil.Exists(filepath.Join(repoPath, "go.mod")):
		return "go-service"
	default:
		return "repository"
	}
}
