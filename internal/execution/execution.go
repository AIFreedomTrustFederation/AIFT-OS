package execution

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Plan struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Steps   []Step `json:"steps"`
}

type Step struct {
	Repo    string   `json:"repo"`
	Path    string   `json:"path"`
	Kind    string   `json:"kind"`
	Manager string   `json:"manager"`
	Build   []string `json:"build"`
	Test    []string `json:"test"`
	Start   []string `json:"start"`
	Status  string   `json:"status"`
}

func Build(cfg config.Config) (Plan, error) {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return Plan{}, err
	}

	plan := Plan{
		Name:    "AIFT-OS Runtime Execution Plan",
		Version: "phase6",
		Steps:   []Step{},
	}

	for _, repo := range repos {
		step := inspect(repo.Name, repo.Path)
		if step.Kind != "" {
			plan.Steps = append(plan.Steps, step)
		}
	}

	sort.Slice(plan.Steps, func(i int, j int) bool {
		return plan.Steps[i].Repo < plan.Steps[j].Repo
	})

	return plan, nil
}

func Write(cfg config.Config) error {
	plan, err := Build(cfg)
	if err != nil {
		return err
	}

	out := filepath.Join(cfg.OSHome, "registry", "execution-plan.json")

	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(out, append(data, '\n'), 0644); err != nil {
		return err
	}

	fmt.Println("Wrote", out)
	return nil
}

func Print(cfg config.Config) error {
	plan, err := Build(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("%-28s %-12s %-14s %-30s %-30s %-30s\n", "REPO", "KIND", "MANAGER", "BUILD", "TEST", "START")

	for _, step := range plan.Steps {
		fmt.Printf(
			"%-28s %-12s %-14s %-30v %-30v %-30v\n",
			step.Repo,
			step.Kind,
			step.Manager,
			step.Build,
			step.Test,
			step.Start,
		)
	}

	return nil
}

func inspect(name string, path string) Step {
	step := Step{
		Repo:   name,
		Path:   path,
		Status: "planned",
	}

	if exists(path, "pnpm-workspace.yaml") {
		step.Kind = "workspace"
		step.Manager = "pnpm"
		step.Build = []string{"pnpm", "run", "build"}
		step.Test = []string{"pnpm", "test"}
		step.Start = []string{"pnpm", "run", "dev"}
		return step
	}

	if exists(path, "package.json") {
		step.Kind = "package"
		step.Manager = "node"
		step.Build = []string{"npm", "run", "build"}
		step.Test = []string{"npm", "test"}
		step.Start = []string{"npm", "run", "dev"}
		return step
	}

	if exists(path, "go.mod") {
		step.Kind = "module"
		step.Manager = "go"
		step.Build = []string{"go", "build", "./..."}
		step.Test = []string{"go", "test", "./..."}
		step.Start = []string{"go", "run", "./cmd/aift"}
		return step
	}

	if exists(path, "Cargo.toml") {
		step.Kind = "crate"
		step.Manager = "cargo"
		step.Build = []string{"cargo", "build"}
		step.Test = []string{"cargo", "test"}
		step.Start = []string{"cargo", "run"}
		return step
	}

	if exists(path, "pyproject.toml") {
		step.Kind = "project"
		step.Manager = "python"
		step.Build = []string{"python", "-m", "build"}
		step.Test = []string{"python", "-m", "pytest"}
		step.Start = []string{"python", "-m", name}
		return step
	}

	if exists(path, "Makefile") {
		step.Kind = "make"
		step.Manager = "make"
		step.Build = []string{"make"}
		step.Test = []string{"make", "test"}
		step.Start = []string{"make", "run"}
		return step
	}

	return step
}

func exists(root string, name string) bool {
	_, err := os.Stat(filepath.Join(root, name))
	return err == nil
}
