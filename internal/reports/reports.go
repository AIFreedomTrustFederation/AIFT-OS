package reports

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/gitx"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

func Dashboard(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	out := filepath.Join(cfg.OSHome, "reports", "dashboard.md")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	total, dirty, valid := 0, 0, 0
	for _, repo := range repos {
		total++
		if gitx.Dirty(repo.Path) {
			dirty++
		}
		if manifests.Valid(repo.Path) {
			valid++
		}
	}

	fmt.Fprintln(f, "# AIFT-OS Federation Dashboard")
	fmt.Fprintln(f)
	fmt.Fprintf(f, "- Total repositories: %d\n", total)
	fmt.Fprintf(f, "- Clean repositories: %d\n", total-dirty)
	fmt.Fprintf(f, "- Dirty repositories: %d\n", dirty)
	fmt.Fprintf(f, "- Valid manifests: %d\n", valid)
	fmt.Fprintf(f, "- Missing/invalid manifests: %d\n\n", total-valid)
	fmt.Fprintln(f, "| Repository | Branch | State | Manifest |")
	fmt.Fprintln(f, "|---|---|---|---|")

	for _, repo := range repos {
		state := "clean"
		if gitx.Dirty(repo.Path) {
			state = "dirty"
		}
		manifest := "valid"
		if !manifests.Valid(repo.Path) {
			manifest = "missing/invalid"
		}
		fmt.Fprintf(f, "| `%s` | `%s` | `%s` | `%s` |\n", repo.Name, gitx.Branch(repo.Path), state, manifest)
	}

	fmt.Println("Wrote", out)
	return nil
}

func Deps(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	out := filepath.Join(cfg.OSHome, "reports", "dependency-graph.md")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "# AIFT Dependency Graph")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "| Repository | Dependencies |")
	fmt.Fprintln(f, "|---|---|")

	for _, repo := range repos {
		fmt.Fprintf(f, "| `%s` | `[]` |\n", repo.Name)
	}

	fmt.Println("Wrote", out)
	return nil
}
