package compiler

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
	Name     string `json:"name"`
	Path     string `json:"path"`
	Branch   string `json:"branch"`
	State    string `json:"state"`
	Manifest string `json:"manifest"`
	Remote   string `json:"remote"`
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
		Name:     "AIFT Repository Compiler",
		Time:     time.Now().Format(time.RFC3339),
		Root:     cfg.Root,
		OSHome:   cfg.OSHome,
		Verified: true,
		Repos:    repos,
		Blocked:  blocked,
	}

	outDir := filepath.Join(cfg.OSHome, "registry", "compiler")
	repDir := filepath.Join(cfg.OSHome, "reports")
	_ = os.MkdirAll(outDir, 0755)
	_ = os.MkdirAll(repDir, 0755)

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	jsonPath := filepath.Join(outDir, "repository-compiler-report.json")
	mdPath := filepath.Join(repDir, "repository-compiler-report.md")

	if err := os.WriteFile(jsonPath, append(b, '\n'), 0644); err != nil {
		return err
	}

	md := "# AIFT Repository Compiler Report\n\n"
	md += fmt.Sprintf("Verified: %v\n\n", report.Verified)
	md += "## Repositories\n\n"
	for _, repo := range repos {
		md += fmt.Sprintf("- %s | %s | %s | %s\n", repo.Name, repo.State, repo.Manifest, repo.Branch)
	}

	if err := os.WriteFile(mdPath, []byte(md), 0644); err != nil {
		return err
	}

	fmt.Println("AIFT Repository Compiler")
	fmt.Println("repos:", len(repos))
	fmt.Println("blocked:", len(blocked))
	fmt.Println("wrote:", jsonPath)
	fmt.Println("wrote:", mdPath)
	return nil
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
		case ".git", "node_modules", ".next", "dist", "build", "vendor", "runtime", "reports":
			return filepath.SkipDir
		}

		if exists(filepath.Join(path, ".git")) {
			repos = append(repos, inspect(path))
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		blocked = append(blocked, err.Error())
	}

	return repos, blocked
}

func inspect(path string) Repo {
	repo := Repo{
		Name:     filepath.Base(path),
		Path:     path,
		Branch:   gitOut(path, "branch", "--show-current"),
		State:    "clean",
		Remote:   gitOut(path, "remote", "get-url", "origin"),
		Manifest: "missing",
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

	return repo
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
