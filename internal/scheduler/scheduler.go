package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/capability"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type Task struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Runtime      string   `json:"runtime"`
	PackageTool  string   `json:"package_tool"`
	BuildCommand string   `json:"build_command"`
	TestCommand  string   `json:"test_command"`
	State        string   `json:"state"`
	Reasons      []string `json:"reasons"`
}

type Report struct {
	Name     string  `json:"name"`
	Time     string  `json:"time"`
	Root     string  `json:"root"`
	OSHome   string  `json:"os_home"`
	Verified bool    `json:"verified"`
	Tasks    []Task  `json:"tasks"`
	Summary  Summary `json:"summary"`
}

type Summary struct {
	Active      int `json:"active"`
	Planned     int `json:"planned"`
	Waiting     int `json:"waiting"`
	Blocked     int `json:"blocked"`
	Unsupported int `json:"unsupported"`
}

func Run(cfg config.Config) error {
	caps := capability.Discover(cfg)
	_ = capability.Write(cfg, caps)

	tasks := discover(cfg.Root, caps)
	report := Report{
		Name:     "AIFT Federation Scheduler",
		Time:     time.Now().Format(time.RFC3339),
		Root:     cfg.Root,
		OSHome:   cfg.OSHome,
		Verified: true,
		Tasks:    tasks,
	}

	for _, task := range tasks {
		switch task.State {
		case "ACTIVE":
			report.Summary.Active++
		case "PLANNED":
			report.Summary.Planned++
		case "WAITING":
			report.Summary.Waiting++
		case "BLOCKED":
			report.Summary.Blocked++
			report.Verified = false
		case "UNSUPPORTED":
			report.Summary.Unsupported++
		}
	}

	if err := writeReport(cfg, report); err != nil {
		return err
	}

	fmt.Println("AIFT Federation Scheduler")
	fmt.Println("verified:", report.Verified)
	fmt.Println("active:", report.Summary.Active)
	fmt.Println("planned:", report.Summary.Planned)
	fmt.Println("waiting:", report.Summary.Waiting)
	fmt.Println("blocked:", report.Summary.Blocked)
	fmt.Println("unsupported:", report.Summary.Unsupported)

	if !report.Verified {
		return fmt.Errorf("scheduler found blocked modules")
	}

	return nil
}

func discover(root string, caps capability.Report) []Task {
	var tasks []Task

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			return nil
		}

		base := filepath.Base(path)
		switch base {
		case ".git", "node_modules", ".next", "dist", "build", "vendor", "runtime", "reports", "registry", ".cache":
			return filepath.SkipDir
		}

		if exists(filepath.Join(path, ".git")) {
			tasks = append(tasks, inspect(path, caps))
			return filepath.SkipDir
		}

		return nil
	})

	return tasks
}

func inspect(path string, caps capability.Report) Task {
	task := Task{
		Name:  filepath.Base(path),
		Path:  path,
		State: "ACTIVE",
	}

	switch {
	case exists(filepath.Join(path, "go.mod")):
		task.Runtime = "go"
		task.PackageTool = "go"
		task.BuildCommand = "go build ./..."
		task.TestCommand = "go test ./..."
		require(&task, caps, "go")

	case exists(filepath.Join(path, "pnpm-lock.yaml")):
		task.Runtime = "node"
		task.PackageTool = "pnpm"
		task.BuildCommand = "pnpm run build"
		task.TestCommand = "pnpm test"
		require(&task, caps, "node")
		require(&task, caps, "pnpm")

	case exists(filepath.Join(path, "package-lock.json")):
		task.Runtime = "node"
		task.PackageTool = "npm"
		task.BuildCommand = "npm run build"
		task.TestCommand = "npm test"
		require(&task, caps, "node")
		require(&task, caps, "npm")

	case exists(filepath.Join(path, "package.json")):
		task.Runtime = "node"
		task.PackageTool = "npm"
		task.BuildCommand = "npm run build"
		task.TestCommand = "npm test"
		require(&task, caps, "node")
		require(&task, caps, "npm")

	case exists(filepath.Join(path, "Cargo.toml")):
		task.Runtime = "rust"
		task.PackageTool = "cargo"
		task.BuildCommand = "cargo build"
		task.TestCommand = "cargo test"
		require(&task, caps, "cargo")

	case exists(filepath.Join(path, "pyproject.toml")):
		task.Runtime = "python"
		task.PackageTool = "python"
		task.BuildCommand = "python -m compileall ."
		task.TestCommand = "python -m pytest"
		require(&task, caps, "python")

	case exists(filepath.Join(path, "requirements.txt")):
		task.Runtime = "python"
		task.PackageTool = "python"
		task.BuildCommand = "python -m compileall ."
		task.TestCommand = "python -m pytest"
		require(&task, caps, "python")

	case exists(filepath.Join(path, "Makefile")):
		task.Runtime = "make"
		task.PackageTool = "make"
		task.BuildCommand = "make"
		task.TestCommand = "make test"
		require(&task, caps, "make")

	default:
		task.Runtime = "unknown"
		task.PackageTool = "none"
		task.State = "UNSUPPORTED"
		task.Reasons = append(task.Reasons, "no provider registered for this repository")
	}

	if !exists(filepath.Join(path, "aift.repo.json")) && !exists(filepath.Join(path, ".aift", "module.json")) {
		if task.State == "ACTIVE" {
			task.State = "PLANNED"
		}
		task.Reasons = append(task.Reasons, "missing AIFT manifest")
	}

	return task
}

func require(task *Task, caps capability.Report, name string) {
	if !capability.Has(caps, name) {
		task.State = "WAITING"
		task.Reasons = append(task.Reasons, "waiting for capability: "+name)
	}
}

func writeReport(cfg config.Config, report Report) error {
	outDir := filepath.Join(cfg.OSHome, "registry", "scheduler")
	reportDir := filepath.Join(cfg.OSHome, "reports")

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return err
	}

	jsonPath := filepath.Join(outDir, "federation-scheduler.json")
	mdPath := filepath.Join(reportDir, "federation-scheduler.md")

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, append(b, '\n'), 0644); err != nil {
		return err
	}

	md := "# AIFT Federation Scheduler Report\n\n"
	md += fmt.Sprintf("Verified: %v\n\n", report.Verified)
	md += fmt.Sprintf("- active: %d\n", report.Summary.Active)
	md += fmt.Sprintf("- planned: %d\n", report.Summary.Planned)
	md += fmt.Sprintf("- waiting: %d\n", report.Summary.Waiting)
	md += fmt.Sprintf("- blocked: %d\n", report.Summary.Blocked)
	md += fmt.Sprintf("- unsupported: %d\n\n", report.Summary.Unsupported)

	md += "## Tasks\n\n"
	for _, task := range report.Tasks {
		md += fmt.Sprintf("- %s | %s | %s | %s\n", task.Name, task.Runtime, task.PackageTool, task.State)
		if len(task.Reasons) > 0 {
			md += "  - reasons: " + strings.Join(task.Reasons, ", ") + "\n"
		}
	}

	return os.WriteFile(mdPath, []byte(md), 0644)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func _unusedExecReference() {
	_ = exec.ErrDot
}
