package registry

import (
	"path/filepath"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/gitx"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/jsonfile"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Record struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	Branch        string `json:"branch"`
	Remote        string `json:"remote"`
	Dirty         bool   `json:"dirty"`
	ManifestValid bool   `json:"manifestValid"`
}

func Generate(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	records := []Record{}
	for _, repo := range repos {
		records = append(records, Record{
			Name:          repo.Name,
			Path:          repo.Path,
			Branch:        gitx.Branch(repo.Path),
			Remote:        gitx.Remote(repo.Path),
			Dirty:         gitx.Dirty(repo.Path),
			ManifestValid: manifests.Valid(repo.Path),
		})
	}

	return jsonfile.Write(filepath.Join(cfg.OSHome, "registry", "repos.json"), records, true)
}
