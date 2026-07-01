package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/state"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Service struct {
	Name        string `json:"name"`
	Repo        string `json:"repo"`
	Status      string `json:"status"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
}

func Defaults() []Service {
	return []Service{
		{Name: "api", Repo: "AIFT-OS", Status: "available", Kind: "core", Description: "Local HTTP API"},
		{Name: "scheduler", Repo: "AIFT-OS", Status: "available", Kind: "core", Description: "Federation scheduler"},
		{Name: "events", Repo: "AIFT-OS", Status: "available", Kind: "core", Description: "Event log and bus"},
		{Name: "providers", Repo: "AIFT-OS", Status: "available", Kind: "core", Description: "Provider registry"},
		{Name: "registry", Repo: "AIFT-OS", Status: "available", Kind: "core", Description: "Repository registry generator"},
		{Name: "reports", Repo: "AIFT-OS", Status: "available", Kind: "core", Description: "Dashboard and dependency reports"},
	}
}

func Discover(cfg config.Config) ([]Service, error) {
	seen := map[string]Service{}
	for _, svc := range Defaults() {
		seen[svc.Name] = svc
	}

	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return nil, err
	}

	for _, repo := range repos {
		addRepoServices(seen, repo.Name, repo.Path)
	}

	out := make([]Service, 0, len(seen))
	for _, svc := range seen {
		out = append(out, svc)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Repo == out[j].Repo {
			return out[i].Name < out[j].Name
		}
		return out[i].Repo < out[j].Repo
	})

	return out, nil
}

func addRepoServices(seen map[string]Service, repoName string, repoPath string) {
	markers := []struct {
		Path        string
		Name        string
		Kind        string
		Description string
	}{
		{"package.json", "node", "package", "Node package or app"},
		{"pnpm-workspace.yaml", "pnpm-workspace", "workspace", "PNPM workspace"},
		{"go.mod", "go", "package", "Go module"},
		{"Cargo.toml", "cargo", "package", "Rust crate"},
		{"pyproject.toml", "python", "package", "Python project"},
		{"Makefile", "make", "build", "Make targets"},
		{".github/workflows", "github-actions", "workflow", "GitHub Actions workflows"},
		{"docs", "docs", "documentation", "Documentation tree"},
		{"schemas", "schemas", "schema", "Schema definitions"},
		{"registry", "registry", "registry", "Registry artifacts"},
		{"reports", "reports", "report", "Report artifacts"},
	}

	for _, marker := range markers {
		full := filepath.Join(repoPath, marker.Path)
		if _, err := os.Stat(full); err == nil {
			name := safeName(repoName + "-" + marker.Name)
			seen[name] = Service{
				Name:        name,
				Repo:        repoName,
				Status:      "discovered",
				Kind:        marker.Kind,
				Description: marker.Description,
			}
		}
	}

	filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := d.Name()
			if base == ".git" || base == "node_modules" || base == ".next" || base == "dist" || base == "build" || base == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		base := strings.ToLower(d.Name())
		if strings.HasSuffix(base, ".service") {
			name := safeName(repoName + "-" + strings.TrimSuffix(d.Name(), ".service"))
			seen[name] = Service{Name: name, Repo: repoName, Status: "discovered", Kind: "service-file", Description: "Service definition file"}
		}
		if base == "docker-compose.yml" || base == "docker-compose.yaml" || base == "compose.yml" || base == "compose.yaml" {
			name := safeName(repoName + "-compose")
			seen[name] = Service{Name: name, Repo: repoName, Status: "discovered", Kind: "container", Description: "Container compose definition"}
		}
		return nil
	})
}

func List(cfg config.Config) error {
	services, err := Discover(cfg)
	if err != nil {
		return err
	}

	st, _ := state.Load(cfg)

	fmt.Printf("%-36s %-24s %-14s %-16s %s\n", "SERVICE", "REPO", "STATUS", "KIND", "DESCRIPTION")
	for _, svc := range services {
		status := svc.Status
		if st.Services != nil && st.Services[svc.Name] != "" {
			status = st.Services[svc.Name]
		}
		fmt.Printf("%-36s %-24s %-14s %-16s %s\n", svc.Name, svc.Repo, status, svc.Kind, svc.Description)
	}

	return nil
}

func Reconcile(cfg config.Config) error {
	discovered, err := Discover(cfg)
	if err != nil {
		return err
	}

	st, _ := state.Load(cfg)
	now := time.Now().Format(time.RFC3339)

	for _, svc := range discovered {
		if st.Services[svc.Name] == "" {
			st.Services[svc.Name] = svc.Status
		}
		st.Catalog[svc.Name] = state.ServiceState{
			Name:        svc.Name,
			Repo:        svc.Repo,
			Status:      st.Services[svc.Name],
			Kind:        svc.Kind,
			Description: svc.Description,
			UpdatedAt:   now,
		}
	}

	if err := state.Save(cfg, st); err != nil {
		return err
	}

	return events.Emit(cfg, "services.reconcile", "services", "service catalog reconciled from discovered workspace reality", map[string]string{
		"services": fmt.Sprintf("%d", len(discovered)),
	})
}

func Mark(cfg config.Config, name string, statusValue string) error {
	st, _ := state.Load(cfg)
	if st.Services == nil {
		st.Services = map[string]string{}
	}
	st.Services[name] = statusValue
	if entry, ok := st.Catalog[name]; ok {
		entry.Status = statusValue
		entry.UpdatedAt = time.Now().Format(time.RFC3339)
		st.Catalog[name] = entry
	}
	if err := state.Save(cfg, st); err != nil {
		return err
	}
	return events.Emit(cfg, "service."+statusValue, name, "service state changed", map[string]string{"service": name, "status": statusValue})
}

func safeName(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		valid := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
		if valid {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "unknown"
	}
	return out
}
