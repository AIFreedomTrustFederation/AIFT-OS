package schedulerplan

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
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/jsonfile"
)

// Job statuses for the scheduler plan.
const (
	StatusPending   = "pending"
	StatusRunnable  = "runnable"
	StatusBlocked   = "blocked"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
	StatusSkipped   = "skipped"
)

// ValidStatuses lists all valid job statuses.
var ValidStatuses = []string{
	StatusPending, StatusRunnable, StatusBlocked,
	StatusRunning, StatusSucceeded, StatusFailed, StatusSkipped,
}

// Job represents a single schedulable unit of work.
type Job struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Kind         string   `json:"kind"`
	Status       string   `json:"status"`
	Command      string   `json:"command,omitempty"`
	Script       string   `json:"script,omitempty"`
	DependsOn    []string `json:"dependsOn,omitempty"`
	BlockedBy    []string `json:"blockedBy,omitempty"`
	Evidence     string   `json:"evidence"`
	Source       string   `json:"source"`
	ReadinessRef string   `json:"readinessRef,omitempty"`
}

// Plan holds the full scheduler plan.
type Plan struct {
	GeneratedAt string  `json:"generatedAt"`
	Jobs        []Job   `json:"jobs"`
	Summary     Summary `json:"summary"`
}

// Summary counts jobs by status.
type Summary struct {
	Total         int            `json:"total"`
	ByStatus      map[string]int `json:"byStatus"`
	ByKind        map[string]int `json:"byKind"`
	RunnableCount int            `json:"runnableCount"`
	BlockedCount  int            `json:"blockedCount"`
}

// IsValid returns true if the status is recognized.
func IsValid(status string) bool {
	for _, s := range ValidStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// readinessObject mirrors the readiness registry JSON structure.
type readinessObject struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
	Source   string `json:"source"`
}

type readinessRegistry struct {
	Objects []readinessObject `json:"objects"`
}

// archCommand mirrors the architecture registry command entry.
type archCommand struct {
	Name       string `json:"name"`
	HasHandler bool   `json:"has_handler"`
	HasHelp    bool   `json:"has_help"`
	Status     string `json:"status"`
}

type archRegistry struct {
	Commands []archCommand `json:"commands"`
}

// BuildPlan generates the scheduler plan from readiness and architecture data.
func BuildPlan(cfg config.Config) (Plan, error) {
	readiness, err := loadReadiness(cfg)
	if err != nil {
		return Plan{}, fmt.Errorf("load readiness: %w", err)
	}

	arch, err := loadArchitecture(cfg)
	if err != nil {
		// Architecture is optional; proceed without it
		arch = archRegistry{}
	}

	jobs := buildJobs(readiness, arch)
	resolveDependencies(jobs)

	plan := Plan{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Jobs:        jobs,
		Summary:     summarize(jobs),
	}

	return plan, nil
}

// GeneratePlan builds the plan and writes registry + report.
func GeneratePlan(cfg config.Config) error {
	plan, err := BuildPlan(cfg)
	if err != nil {
		return err
	}

	if err := writePlanRegistry(cfg, plan); err != nil {
		return err
	}
	if err := writePlanReport(cfg, plan); err != nil {
		return err
	}

	return events.Emit(cfg, "scheduler.plan", "schedulerplan",
		fmt.Sprintf("scheduler plan: %d jobs (%d runnable, %d blocked)",
			plan.Summary.Total, plan.Summary.RunnableCount, plan.Summary.BlockedCount),
		map[string]string{"total": fmt.Sprint(plan.Summary.Total)})
}

// PrintStatus prints all jobs in a table.
func PrintStatus(cfg config.Config) error {
	plan, err := loadOrBuild(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("%-14s %-30s %-12s %s\n", "KIND", "NAME", "STATUS", "EVIDENCE")
	for _, j := range plan.Jobs {
		fmt.Printf("%-14s %-30s %-12s %s\n", j.Kind, j.Name, j.Status, j.Evidence)
	}
	return nil
}

// PrintReady prints only runnable jobs.
func PrintReady(cfg config.Config) error {
	plan, err := loadOrBuild(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("%-14s %-30s %-12s %s\n", "KIND", "NAME", "STATUS", "EVIDENCE")
	for _, j := range plan.Jobs {
		if j.Status == StatusRunnable {
			fmt.Printf("%-14s %-30s %-12s %s\n", j.Kind, j.Name, j.Status, j.Evidence)
		}
	}
	return nil
}

// PrintBlocked prints only blocked jobs.
func PrintBlocked(cfg config.Config) error {
	plan, err := loadOrBuild(cfg)
	if err != nil {
		return err
	}

	count := 0
	fmt.Printf("%-14s %-30s %-12s %s\n", "KIND", "NAME", "STATUS", "BLOCKED BY")
	for _, j := range plan.Jobs {
		if j.Status == StatusBlocked {
			fmt.Printf("%-14s %-30s %-12s %s\n", j.Kind, j.Name, j.Status, strings.Join(j.BlockedBy, ", "))
			count++
		}
	}
	if count == 0 {
		fmt.Println("No blocked jobs.")
	}
	return nil
}

// PrintReport prints the scheduler plan report.
func PrintReport(cfg config.Config) error {
	path := filepath.Join(cfg.OSHome, "reports", "scheduler-plan.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if err := GeneratePlan(cfg); err != nil {
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

// ── Job Building ────────────────────────────────────────────────────

func buildJobs(rr readinessRegistry, ar archRegistry) []Job {
	var jobs []Job

	// Build command index for quick lookup
	cmdIndex := map[string]archCommand{}
	for _, c := range ar.Commands {
		cmdIndex[c.Name] = c
	}

	for _, obj := range rr.Objects {
		job := objectToJob(obj, cmdIndex)
		jobs = append(jobs, job)
	}

	sort.Slice(jobs, func(i, j int) bool {
		if jobs[i].Kind != jobs[j].Kind {
			return jobs[i].Kind < jobs[j].Kind
		}
		return jobs[i].ID < jobs[j].ID
	})

	return jobs
}

func objectToJob(obj readinessObject, cmdIndex map[string]archCommand) Job {
	job := Job{
		ID:           "job:" + obj.ID,
		Name:         obj.Name,
		Kind:         obj.Kind,
		Evidence:     obj.Evidence,
		Source:       obj.Source,
		ReadinessRef: obj.ID,
	}

	switch obj.Status {
	case "ready", "active":
		job.Status = StatusRunnable
		job.Evidence = "readiness: " + obj.Status
	case "planned":
		job.Status = StatusPending
		job.Evidence = "readiness: planned"
	case "blocked":
		job.Status = StatusBlocked
		job.Evidence = "readiness: blocked"
		job.BlockedBy = []string{"prerequisite not ready"}
	case "deprecated", "removed":
		job.Status = StatusSkipped
		job.Evidence = "readiness: " + obj.Status
	default:
		job.Status = StatusPending
		job.Evidence = "readiness: " + obj.Status
	}

	// Attach command/script if discoverable
	switch obj.Kind {
	case "command":
		if ac, ok := cmdIndex[obj.Name]; ok {
			if ac.HasHandler {
				job.Command = "aift " + obj.Name
			}
			if ac.Status == "planned" {
				job.Status = StatusPending
				job.Evidence = "command is planned (not implemented)"
				job.Command = ""
			}
		}
	case "script":
		job.Script = obj.Source
	}

	return job
}

// resolveDependencies sets inter-job dependencies based on kind ordering.
// Commands depend on their prerequisite readiness objects being runnable.
func resolveDependencies(jobs []Job) {
	// Build index for lookup
	byRef := map[string]*Job{}
	for i := range jobs {
		byRef[jobs[i].ReadinessRef] = &jobs[i]
	}

	// Commands that run verify-like operations depend on repositories being ready
	repoReady := false
	for _, j := range jobs {
		if j.Kind == "repository" && j.Status == StatusRunnable {
			repoReady = true
			break
		}
	}

	for i := range jobs {
		j := &jobs[i]

		// If there are no runnable repos, command/script jobs are blocked
		if !repoReady && (j.Kind == "command" || j.Kind == "script") && j.Status == StatusRunnable {
			j.Status = StatusBlocked
			j.BlockedBy = []string{"no runnable repositories"}
			j.Evidence = "blocked: no repository is ready"
		}

		// Capability jobs depend on their source repo being ready
		if j.Kind == "capability" && j.Status == StatusRunnable {
			repoRef := "repo:" + j.Source
			if repo, ok := byRef[repoRef]; ok {
				if repo.Status != StatusRunnable {
					j.Status = StatusBlocked
					j.BlockedBy = []string{repoRef}
					j.DependsOn = []string{repoRef}
					j.Evidence = "blocked: source repository not ready"
				}
			}
		}

		// Service jobs depend on their source repo being ready
		if j.Kind == "service" && j.Status == StatusRunnable {
			repoRef := "repo:" + j.Source
			if repo, ok := byRef[repoRef]; ok {
				if repo.Status != StatusRunnable {
					j.Status = StatusBlocked
					j.BlockedBy = []string{repoRef}
					j.DependsOn = []string{repoRef}
					j.Evidence = "blocked: source repository not ready"
				}
			}
		}

		// Module jobs depend on their source repo
		if j.Kind == "module" && j.Status == StatusRunnable {
			repoRef := "repo:" + j.Source
			if repo, ok := byRef[repoRef]; ok {
				if repo.Status != StatusRunnable {
					j.Status = StatusBlocked
					j.BlockedBy = []string{repoRef}
					j.DependsOn = []string{repoRef}
					j.Evidence = "blocked: source repository not ready"
				}
			}
		}
	}
}

// ── Helpers ─────────────────────────────────────────────────────────

func summarize(jobs []Job) Summary {
	s := Summary{
		Total:    len(jobs),
		ByStatus: map[string]int{},
		ByKind:   map[string]int{},
	}
	for _, j := range jobs {
		s.ByStatus[j.Status]++
		s.ByKind[j.Kind]++
		if j.Status == StatusRunnable {
			s.RunnableCount++
		}
		if j.Status == StatusBlocked {
			s.BlockedCount++
		}
	}
	return s
}

func loadReadiness(cfg config.Config) (readinessRegistry, error) {
	path := filepath.Join(cfg.OSHome, "registry", "runtime-readiness.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return readinessRegistry{}, fmt.Errorf("runtime-readiness.json not found: run 'aift runtime scan' first")
	}

	var rr readinessRegistry
	if err := json.Unmarshal(data, &rr); err != nil {
		return readinessRegistry{}, err
	}
	return rr, nil
}

func loadArchitecture(cfg config.Config) (archRegistry, error) {
	path := filepath.Join(cfg.OSHome, "registry", "architecture.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return archRegistry{}, err
	}

	var ar archRegistry
	if err := json.Unmarshal(data, &ar); err != nil {
		return archRegistry{}, err
	}
	return ar, nil
}

func writePlanRegistry(cfg config.Config, plan Plan) error {
	return jsonfile.Write(filepath.Join(cfg.OSHome, "registry", "scheduler-plan.json"), plan, true)
}

func writePlanReport(cfg config.Config, plan Plan) error {
	out := filepath.Join(cfg.OSHome, "reports", "scheduler-plan.md")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# Scheduler Plan\n\n")
	b.WriteString(fmt.Sprintf("Generated: %s\n\n", plan.GeneratedAt))

	// Summary
	b.WriteString("## Summary\n\n")
	b.WriteString(fmt.Sprintf("- **Total jobs**: %d\n", plan.Summary.Total))
	b.WriteString(fmt.Sprintf("- **Runnable**: %d\n", plan.Summary.RunnableCount))
	b.WriteString(fmt.Sprintf("- **Blocked**: %d\n", plan.Summary.BlockedCount))
	b.WriteString("\n")

	// By status
	b.WriteString("### By Status\n\n")
	for _, status := range ValidStatuses {
		if count, ok := plan.Summary.ByStatus[status]; ok {
			b.WriteString(fmt.Sprintf("- **%s**: %d\n", status, count))
		}
	}
	b.WriteString("\n")

	// By kind
	b.WriteString("### By Kind\n\n")
	kinds := sortedKeys(plan.Summary.ByKind)
	for _, kind := range kinds {
		b.WriteString(fmt.Sprintf("- **%s**: %d\n", kind, plan.Summary.ByKind[kind]))
	}
	b.WriteString("\n")

	// Runnable jobs
	b.WriteString("## Runnable Jobs\n\n")
	b.WriteString("| Kind | Name | Command/Script | Evidence |\n")
	b.WriteString("|---|---|---|---|\n")
	runnableCount := 0
	for _, j := range plan.Jobs {
		if j.Status == StatusRunnable {
			cmd := j.Command
			if cmd == "" {
				cmd = j.Script
			}
			if cmd == "" {
				cmd = "-"
			}
			b.WriteString(fmt.Sprintf("| %s | `%s` | `%s` | %s |\n", j.Kind, j.Name, cmd, j.Evidence))
			runnableCount++
		}
	}
	if runnableCount == 0 {
		b.WriteString("| - | No runnable jobs | - | - |\n")
	}
	b.WriteString("\n")

	// Blocked jobs
	b.WriteString("## Blocked Jobs\n\n")
	b.WriteString("| Kind | Name | Blocked By | Evidence |\n")
	b.WriteString("|---|---|---|---|\n")
	blockedCount := 0
	for _, j := range plan.Jobs {
		if j.Status == StatusBlocked {
			b.WriteString(fmt.Sprintf("| %s | `%s` | %s | %s |\n",
				j.Kind, j.Name, strings.Join(j.BlockedBy, ", "), j.Evidence))
			blockedCount++
		}
	}
	if blockedCount == 0 {
		b.WriteString("| - | No blocked jobs | - | - |\n")
	}
	b.WriteString("\n")

	// Full table
	b.WriteString("## All Jobs\n\n")
	b.WriteString("| Kind | Name | Status | Evidence | Source |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, j := range plan.Jobs {
		b.WriteString(fmt.Sprintf("| %s | `%s` | %s | %s | %s |\n",
			j.Kind, j.Name, j.Status, j.Evidence, j.Source))
	}

	return os.WriteFile(out, []byte(b.String()), 0644)
}

func loadOrBuild(cfg config.Config) (Plan, error) {
	path := filepath.Join(cfg.OSHome, "registry", "scheduler-plan.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if err := GeneratePlan(cfg); err != nil {
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

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
