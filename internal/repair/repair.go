package repair

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type ReportData struct {
	Name      string   `json:"name"`
	Time      string   `json:"time"`
	Mode      string   `json:"mode"`
	Verified  bool     `json:"verified"`
	Actions   []string `json:"actions"`
	Blocked   []string `json:"blocked"`
	Workspace string   `json:"workspace"`
}

func Run(cfg config.Config, safe bool) error {
	r := ReportData{
		Name:      "AIFT Self-Repair AI Kernel",
		Time:      time.Now().Format(time.RFC3339),
		Mode:      "safe",
		Workspace: cfg.Root,
		Actions:   []string{},
		Blocked:   []string{},
	}

	if !safe {
		r.Mode = "blocked"
		r.Blocked = append(r.Blocked, "unsafe repair mode is not implemented")
		return writeReport(cfg, r)
	}

	for _, dir := range []string{
		"registry/repairs",
		"registry/tests",
		"reports",
		"runtime/logs",
		"runtime/cache",
		"runtime/events",
	} {
		path := filepath.Join(cfg.OSHome, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			r.Blocked = append(r.Blocked, "mkdir failed: "+dir+": "+err.Error())
		} else {
			r.Actions = append(r.Actions, "ensured directory: "+dir)
		}
	}

	restoreGenerated(cfg, &r)
	ensureFile(filepath.Join(cfg.OSHome, "registry/repairs/repairs.json"), "[]\n", &r)
	ensureFile(filepath.Join(cfg.OSHome, "registry/tests/tests.json"), "[]\n", &r)

	if err := run(cfg.OSHome, "gofmt", "-w", "internal/ai/ai.go", "internal/repair/repair.go"); err != nil {
		r.Blocked = append(r.Blocked, "gofmt blocked: "+err.Error())
	}

	if err := run(cfg.OSHome, "go", "test", "./..."); err != nil {
		r.Verified = false
		r.Blocked = append(r.Blocked, "go test blocked: "+err.Error())
	} else {
		r.Verified = true
		r.Actions = append(r.Actions, "go test passed")
	}

	return writeReport(cfg, r)
}

func Verify(cfg config.Config) error {
	if err := run(cfg.OSHome, "go", "test", "./..."); err != nil {
		return err
	}
	fmt.Println("OK: self-repair verification passed")
	return nil
}

func Report(cfg config.Config) error {
	data, err := os.ReadFile(filepath.Join(cfg.OSHome, "reports/self-repair.md"))
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

func restoreGenerated(cfg config.Config, r *ReportData) {
	files := []string{
		".aift/capabilities.json",
		".aift/providers.json",
		".aift/repos.json",
		".aift/workflows.json",
		"var/events/events.jsonl",
	}
	for _, file := range files {
		_ = run(cfg.OSHome, "git", "restore", file)
		r.Actions = append(r.Actions, "restored generated file if tracked: "+file)
	}
}

func ensureFile(path string, content string, r *ReportData) {
	if _, err := os.Stat(path); err == nil {
		return
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		r.Blocked = append(r.Blocked, "write failed: "+path+": "+err.Error())
		return
	}
	r.Actions = append(r.Actions, "created scaffold: "+path)
}

func writeReport(cfg config.Config, r ReportData) error {
	jsonPath := filepath.Join(cfg.OSHome, "registry/repairs/self-repair-report.json")
	mdPath := filepath.Join(cfg.OSHome, "reports/self-repair.md")

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(jsonPath, append(data, '\n'), 0644); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# AIFT Self-Repair Report\n\n")
	b.WriteString("Mode: " + r.Mode + "\n\n")
	if r.Verified {
		b.WriteString("Verified: true\n\n")
	} else {
		b.WriteString("Verified: false\n\n")
	}
	b.WriteString("## Actions\n\n")
	for _, action := range r.Actions {
		b.WriteString("- " + action + "\n")
	}
	b.WriteString("\n## Blocked\n\n")
	if len(r.Blocked) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, blocked := range r.Blocked {
			b.WriteString("- " + blocked + "\n")
		}
	}

	if err := os.WriteFile(mdPath, []byte(b.String()), 0644); err != nil {
		return err
	}

	fmt.Println("Wrote registry/repairs/self-repair-report.json")
	fmt.Println("Wrote reports/self-repair.md")
	return nil
}

func run(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %s", name, strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return nil
}
