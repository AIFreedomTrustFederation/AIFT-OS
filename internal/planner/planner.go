package planner

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

type Capability struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Command  string `json:"command,omitempty"`
	Evidence string `json:"evidence,omitempty"`
}

type Service struct {
	Name     string   `json:"name"`
	Kind     string   `json:"kind"`
	Status   string   `json:"status"`
	Version  string   `json:"version"`
	Provides []string `json:"provides"`
	Requires []string `json:"requires"`
	Events   []string `json:"events"`
	Health   string   `json:"health,omitempty"`
	Start    string   `json:"start,omitempty"`
	Stop     string   `json:"stop,omitempty"`
	Evidence string   `json:"evidence"`
}

type RepoPlan struct {
	Repo          string       `json:"repo"`
	Path          string       `json:"path"`
	State         string       `json:"state"`
	Reason        string       `json:"reason"`
	Ready         []string     `json:"ready"`
	V1            []string     `json:"v1"`
	Detected      []string     `json:"detected"`
	Planned       []string     `json:"planned"`
	Broken        []string     `json:"broken"`
	Blocked       []string     `json:"blocked"`
	Services      []Service    `json:"services"`
	Capabilities  []Capability `json:"capabilities"`
	Recommended   []string     `json:"recommended"`
	LastEvaluated string       `json:"lastEvaluated"`
}

type Plan struct {
	GeneratedAt string     `json:"generatedAt"`
	Summary     Summary    `json:"summary"`
	Repos       []RepoPlan `json:"repos"`
}

type Summary struct {
	Repos    int `json:"repos"`
	Ready    int `json:"ready"`
	Planned  int `json:"planned"`
	Blocked  int `json:"blocked"`
	Broken   int `json:"broken"`
	Detected int `json:"detected"`
}

func Build(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	plan := Plan{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Repos:       []RepoPlan{},
	}

	for _, r := range repos {
		rp := evaluateRepo(r.Name, r.Path)
		plan.Repos = append(plan.Repos, rp)
		plan.Summary.Repos++

		switch rp.State {
		case "READY":
			plan.Summary.Ready++
		case "PLANNED":
			plan.Summary.Planned++
		case "BLOCKED":
			plan.Summary.Blocked++
		case "BROKEN":
			plan.Summary.Broken++
		case "DETECTED":
			plan.Summary.Detected++
		}
	}

	sort.Slice(plan.Repos, func(i, j int) bool {
		return plan.Repos[i].Repo < plan.Repos[j].Repo
	})

	if err := writeRegistry(cfg, plan); err != nil {
		return err
	}
	if err := writeReports(cfg, plan); err != nil {
		return err
	}

	return events.Emit(cfg, "planner.build", "planner", "execution plan built", map[string]string{
		"repos":   fmt.Sprint(plan.Summary.Repos),
		"ready":   fmt.Sprint(plan.Summary.Ready),
		"blocked": fmt.Sprint(plan.Summary.Blocked),
	})
}

func SummaryReport(cfg config.Config) error {
	plan, err := loadOrBuild(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("%-30s %-10s %s\n", "REPOSITORY", "STATE", "REASON")
	for _, r := range plan.Repos {
		fmt.Printf("%-30s %-10s %s\n", r.Repo, r.State, r.Reason)
	}
	return nil
}

func Repo(cfg config.Config, name string) error {
	plan, err := loadOrBuild(cfg)
	if err != nil {
		return err
	}

	for _, r := range plan.Repos {
		if r.Repo != name {
			continue
		}

		fmt.Println("Repository:", r.Repo)
		fmt.Println("State:", r.State)
		fmt.Println("Reason:", r.Reason)
		fmt.Println("Ready:", strings.Join(r.Ready, ", "))
		fmt.Println("V1:", strings.Join(r.V1, ", "))
		fmt.Println("Detected:", strings.Join(r.Detected, ", "))
		fmt.Println("Planned:", strings.Join(r.Planned, ", "))
		fmt.Println("Broken:", strings.Join(r.Broken, ", "))
		fmt.Println("Blocked:", strings.Join(r.Blocked, ", "))
		fmt.Println()
		fmt.Println("Recommended:")
		for _, rec := range r.Recommended {
			fmt.Println("-", rec)
		}
		return nil
	}

	return fmt.Errorf("repo not found in execution plan: %s", name)
}

func Ready(cfg config.Config) error {
	plan, err := loadOrBuild(cfg)
	if err != nil {
		return err
	}

	for _, r := range plan.Repos {
		if r.State == "READY" {
			fmt.Println(r.Repo)
		}
	}
	return nil
}

func Blocked(cfg config.Config) error {
	plan, err := loadOrBuild(cfg)
	if err != nil {
		return err
	}

	for _, r := range plan.Repos {
		if r.State == "BLOCKED" || r.State == "BROKEN" {
			fmt.Printf("%-30s %-10s %s\n", r.Repo, r.State, r.Reason)
		}
	}
	return nil
}

func Report(cfg config.Config) error {
	path := filepath.Join(cfg.OSHome, "reports", "execution-plan.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if err := Build(cfg); err != nil {
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

func evaluateRepo(name string, repoPath string) RepoPlan {
	rp := RepoPlan{
		Repo:          name,
		Path:          repoPath,
		LastEvaluated: time.Now().Format(time.RFC3339),
	}

	rp.Capabilities = loadCapabilities(repoPath)
	rp.Services = loadServices(repoPath)

	for _, c := range rp.Capabilities {
		switch c.Status {
		case "ready":
			rp.Ready = append(rp.Ready, c.Name)
		case "v1":
			rp.V1 = append(rp.V1, c.Name)
		case "detected":
			rp.Detected = append(rp.Detected, c.Name)
		case "planned":
			rp.Planned = append(rp.Planned, c.Name)
		case "broken":
			rp.Broken = append(rp.Broken, c.Name)
		}
	}

	sort.Strings(rp.Ready)
	sort.Strings(rp.V1)
	sort.Strings(rp.Detected)
	sort.Strings(rp.Planned)
	sort.Strings(rp.Broken)

	serviceBlocked := []string{}
	for _, s := range rp.Services {
		if s.Status == "broken" {
			serviceBlocked = append(serviceBlocked, s.Name+": service broken")
		}
		if s.Status == "planned" {
			serviceBlocked = append(serviceBlocked, s.Name+": service planned")
		}
		for _, req := range s.Requires {
			if !hasCapability(rp, req) {
				serviceBlocked = append(serviceBlocked, s.Name+": missing required capability "+req)
			}
		}
	}

	rp.Blocked = append(rp.Blocked, serviceBlocked...)
	sort.Strings(rp.Blocked)

	switch {
	case len(rp.Broken) > 0:
		rp.State = "BROKEN"
		rp.Reason = "one or more capabilities are broken"
	case len(rp.Blocked) > 0:
		rp.State = "BLOCKED"
		rp.Reason = "service contract has unmet planned or missing requirements"
	case len(rp.Ready)+len(rp.V1) > 0:
		rp.State = "READY"
		rp.Reason = "has at least one ready or v1 capability and no blocking service requirements"
	case len(rp.Detected) > 0:
		rp.State = "DETECTED"
		rp.Reason = "repo has detected capabilities but none are proven executable"
	default:
		rp.State = "PLANNED"
		rp.Reason = "repo is known but has no ready/v1 capabilities"
	}

	rp.Recommended = recommendations(rp)
	return rp
}

func loadCapabilities(repoPath string) []Capability {
	path := filepath.Join(repoPath, ".aift", "capabilities.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return []Capability{}
	}

	var rc struct {
		Capabilities []Capability `json:"capabilities"`
	}
	if json.Unmarshal(data, &rc) != nil {
		return []Capability{}
	}
	return rc.Capabilities
}

func loadServices(repoPath string) []Service {
	path := filepath.Join(repoPath, ".aift", "services.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return []Service{}
	}

	var rc struct {
		Services []Service `json:"services"`
	}
	if json.Unmarshal(data, &rc) != nil {
		return []Service{}
	}
	return rc.Services
}

func hasCapability(rp RepoPlan, name string) bool {
	for _, c := range rp.Capabilities {
		if c.Name == name && (c.Status == "ready" || c.Status == "v1") {
			return true
		}
	}
	return false
}

func recommendations(rp RepoPlan) []string {
	var out []string

	if len(rp.Broken) > 0 {
		out = append(out, "Fix broken capabilities before any runtime execution.")
	}

	if !contains(rp.Ready, "verify") && !contains(rp.V1, "verify") {
		out = append(out, "Add and prove `.aift/commands/verify.sh`.")
	}

	if contains(rp.Detected, "build") {
		out = append(out, "Convert detected build support into `.aift/commands/build.sh`.")
	}

	if contains(rp.Detected, "test") {
		out = append(out, "Convert detected test support into `.aift/commands/test.sh`.")
	}

	if contains(rp.Planned, "start") {
		out = append(out, "Keep start orchestration disabled until a real start command and health check exist.")
	}

	if len(rp.Services) == 0 {
		out = append(out, "Add `.aift/services.json` service contract.")
	}

	if len(out) == 0 {
		out = append(out, "No immediate planner blockers detected.")
	}

	return out
}

func writeRegistry(cfg config.Config, plan Plan) error {
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

func writeReports(cfg config.Config, plan Plan) error {
	if err := os.MkdirAll(filepath.Join(cfg.OSHome, "reports"), 0755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# Federation Execution Plan\n\n")
	b.WriteString(fmt.Sprintf("- Repos: %d\n", plan.Summary.Repos))
	b.WriteString(fmt.Sprintf("- Ready: %d\n", plan.Summary.Ready))
	b.WriteString(fmt.Sprintf("- Detected: %d\n", plan.Summary.Detected))
	b.WriteString(fmt.Sprintf("- Planned: %d\n", plan.Summary.Planned))
	b.WriteString(fmt.Sprintf("- Blocked: %d\n", plan.Summary.Blocked))
	b.WriteString(fmt.Sprintf("- Broken: %d\n\n", plan.Summary.Broken))

	b.WriteString("| Repository | State | Reason |\n")
	b.WriteString("|---|---|---|\n")
	for _, r := range plan.Repos {
		b.WriteString(fmt.Sprintf("| `%s` | `%s` | %s |\n", r.Repo, r.State, r.Reason))
	}

	b.WriteString("\n## Repository Detail\n\n")
	for _, r := range plan.Repos {
		b.WriteString("### " + r.Repo + "\n\n")
		b.WriteString(fmt.Sprintf("- State: `%s`\n", r.State))
		b.WriteString(fmt.Sprintf("- Reason: %s\n", r.Reason))
		b.WriteString(fmt.Sprintf("- Ready: `%s`\n", strings.Join(r.Ready, ", ")))
		b.WriteString(fmt.Sprintf("- V1: `%s`\n", strings.Join(r.V1, ", ")))
		b.WriteString(fmt.Sprintf("- Detected: `%s`\n", strings.Join(r.Detected, ", ")))
		b.WriteString(fmt.Sprintf("- Planned: `%s`\n", strings.Join(r.Planned, ", ")))
		b.WriteString(fmt.Sprintf("- Broken: `%s`\n", strings.Join(r.Broken, ", ")))
		if len(r.Blocked) > 0 {
			b.WriteString("\nBlockers:\n")
			for _, blocker := range r.Blocked {
				b.WriteString("- " + blocker + "\n")
			}
		}
		b.WriteString("\nRecommendations:\n")
		for _, rec := range r.Recommended {
			b.WriteString("- " + rec + "\n")
		}
		b.WriteString("\n")
	}

	if err := os.WriteFile(filepath.Join(cfg.OSHome, "reports", "execution-plan.md"), []byte(b.String()), 0644); err != nil {
		return err
	}

	blockers := "# Execution Blockers\n\n"
	for _, r := range plan.Repos {
		if r.State == "BLOCKED" || r.State == "BROKEN" {
			blockers += "## " + r.Repo + "\n\n"
			blockers += "- State: `" + r.State + "`\n"
			blockers += "- Reason: " + r.Reason + "\n"
			for _, b := range r.Blocked {
				blockers += "- " + b + "\n"
			}
			blockers += "\n"
		}
	}

	return os.WriteFile(filepath.Join(cfg.OSHome, "reports", "execution-blockers.md"), []byte(blockers), 0644)
}

func loadOrBuild(cfg config.Config) (Plan, error) {
	path := filepath.Join(cfg.OSHome, "registry", "execution-plan.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if err := Build(cfg); err != nil {
			return Plan{}, err
		}
		data, err = os.ReadFile(path)
		if err != nil {
			return Plan{}, err
		}
	}

	var plan Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func contains(items []string, wanted string) bool {
	for _, item := range items {
		if item == wanted {
			return true
		}
	}
	return false
}
