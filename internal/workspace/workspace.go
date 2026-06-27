package workspace

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type Repo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func FindRepos(cfg config.Config) ([]Repo, error) {
	var repos []Repo

	err := filepath.WalkDir(cfg.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}

		if d.Name() == ".git" {
			repoPath := filepath.Dir(path)
			repos = append(repos, Repo{Name: filepath.Base(repoPath), Path: repoPath})
			return filepath.SkipDir
		}

		if d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == ".next" {
			return filepath.SkipDir
		}

		return nil
	})

	sort.Slice(repos, func(i, j int) bool { return repos[i].Name < repos[j].Name })
	return repos, err
}
