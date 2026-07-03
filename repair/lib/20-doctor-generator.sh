#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail
cd "${AIFT_ROOT:-$HOME/AIFT}/AIFT-OS"

cat > docs/architecture/PHASE7-AIFT-DOCTOR-GIT-HOUSEKEEPING.md <<'DOC'
# AIFT-OS Phase 7: Doctor and Git Housekeeping

AIFT Doctor inspects the real local federation workspace.

It repairs safe generated state, checks git status, verifies the native Go CLI, and reports what still needs human review.

AIFT-OS remains the central runtime. Other repositories remain source packages.
DOC

cat > internal/doctor/housekeeping.go <<'GO'
package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

var generatedStatePaths = []string{
	".aift/capabilities.json",
	".aift/providers.json",
	".aift/workflows.json",
	".aift/repos.json",
	"var/events/events.jsonl",
}

func Git(cfg config.Config) error {
	repos, err := gitRepos(cfg.Root)
	if err != nil {
		return err
	}

	for _, repo := range repos {
		fmt.Fprintf(os.Stdout, "== %s ==\n", repo)
		cmd := exec.Command("git", "-C", repo, "status", "--short")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func Repair(cfg config.Config) error {
	fmt.Fprintln(os.Stdout, "Repairing generated runtime state")

	repos, err := gitRepos(cfg.Root)
	if err != nil {
		return err
	}

	for _, repo := range repos {
		args := append([]string{"-C", repo, "restore"}, generatedStatePaths...)
		cmd := exec.Command("git", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()

		modulePath := filepath.Join(repo, ".aift", "module.json")
		if err := os.Remove(modulePath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

func Full(cfg config.Config) error {
	if err := Repair(cfg); err != nil {
		return err
	}
	return Run(cfg)
}

func gitRepos(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var repos []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repo := filepath.Join(root, entry.Name())
		gitDir := filepath.Join(repo, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			repos = append(repos, repo)
		}
	}

	sort.Strings(repos)
	return repos, nil
}
GO
