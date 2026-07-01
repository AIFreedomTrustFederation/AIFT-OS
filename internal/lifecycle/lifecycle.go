package lifecycle

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

type Repo struct {
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	Branch    string   `json:"branch"`
	State     string   `json:"state"`
	Manifest  string   `json:"manifest"`
	Remote    string   `json:"remote"`
	Builds    []string `json:"builds"`
	Tests     []string `json:"tests"`
	Blocked   []string `json:"blocked"`
	Artifacts []string `json:"artifacts"`
}

type Report struct {
	Name     string   `json:"name"`
	Time     string   `json:"time"`
	Root     string   `json:"root"`
	OSHome   string   `json:"os_home"`
	Verified bool     `json:"verified"`
	Repos    []Repo   `json:"repos"`
	Blocked  []string `json:"blocked"`
}

func Run(cfg config.Config) error {
	repos, blocked := discover(cfg.Root)

	report := Report{
		Name:     "AIFT Federation Lifecycle",
		Time:     time.Now().Format(time.RFC3339),
		Root:     cfg.Root,
		OSHome:   cfg.OSHome,
		Verified: len(blocked) == 0,
		Repos:    repos,
		Blocked:  blocked,
	}

	return writeReport(cfg, report)
}

func discover(root string) ([]Repo, []string) {
	var repos []Repo
	var blocked []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			blocked = append(blocked, path+": "+err.Error())
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
			repos = append(repos, inspectRepo(path))
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		blocked = append(blocked, err.Error())
	}

	return repos, blocked
}

func inspectRepo(path string) Repo {
	repo := Repo{
		Name:     filepath.Base(path),
		Path:     path,
		Branch:   gitOut(path, "branch", "--show-current"),
		State:    "clean",
		Manifest: "missing",
		Remote:   gitOut(path, "remote", "get-url", "origin"),
	}

	if repo.Branch == "" {
		repo.Branch = "unknown"
	}

	if gitOut(path, "status", "--short") != "" {
		repo.State = "dirty"
	}

	if exists(filepath.Join(path, "aift.repo.json")) || exists(filepath.Join(path, ".aift", "module.json")) {
		repo.Manifest = "valid"
	}

	if exists(filepath.Join(path, "go.mod")) {
		repo.Builds = append(repo.Builds, "go build ./...")
		repo.Tests = append(repo.Tests, "go test ./...")
	}

	if exists(filepath.Join(path, "package.json")) {
		repo.Builds = append(repo.Builds, "npm run build")
		repo.Tests = append(repo.Tests, "npm test")
	}

	if exists(filepath.Join(path, "Cargo.toml")) {
		repo.Builds = append(repo.Builds, "cargo build")
		repo.Tests = append(repo.Tests, "cargo test")
	}

	if exists(filepath.Join(path, "Makefile")) {
		repo.Builds = append(repo.Builds, "make")
		repo.Tests = append(repo.Tests, "make test")
	}

	if len(repo.Builds) == 0 {
		repo.Blocked = append(repo.Blocked, "no build system detected")
	}

	if repo.Manifest == "missing" {
		repo.Blocked = append(repo.Blocked, "missing AIFT manifest")
	}

	return repo
}

func writeReport(cfg config.Config, report Report) error {
	outDir := filepath.Join(cfg.OSHome, "registry", "lifecycle")
	reportDir := filepath.Join(cfg.OSHome, "reports")

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return err
	}

	jsonPath := filepath.Join(outDir, "federation-lifecycle.json")
	mdPath := filepath.Join(reportDir, "federation-lifecycle.md")

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(jsonPath, append(b, '\n'), 0644); err != nil {
		return err
	}

	md := "# AIFT Federation Lifecycle Report\n\n"
	md += fmt.Sprintf("Verified: %v\n\n", report.Verified)
	md += "## Repositories\n\n"

	for _, repo := range report.Repos {
		md += fmt.Sprintf("- %s | %s | %s | %s\n", repo.Name, repo.State, repo.Manifest, repo.Branch)

		if len(repo.Builds) > 0 {
			md += "  - builds: " + strings.Join(repo.Builds, ", ") + "\n"
		}

		if len(repo.Tests) > 0 {
			md += "  - tests: " + strings.Join(repo.Tests, ", ") + "\n"
		}

		if len(repo.Blocked) > 0 {
			md += "  - blocked: " + strings.Join(repo.Blocked, ", ") + "\n"
		}
	}

	if len(report.Blocked) > 0 {
		md += "\n## Blocked\n\n"
		for _, item := range report.Blocked {
			md += "- " + item + "\n"
		}
	}

	if err := os.WriteFile(mdPath, []byte(md), 0644); err != nil {
		return err
	}

	fmt.Println("AIFT Federation Lifecycle")
	fmt.Println("repos:", len(report.Repos))
	fmt.Println("blocked:", len(report.Blocked))
	fmt.Println("wrote:", jsonPath)
	fmt.Println("wrote:", mdPath)

	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func gitOut(dir string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
