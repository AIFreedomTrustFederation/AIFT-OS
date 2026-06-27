package sync

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/gitx"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

func Safe(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	fmt.Println("AIFT safe sync: pulls clean repos only; dirty repos are skipped.")

	for _, repo := range repos {
		if gitx.Remote(repo.Path) == "" {
			fmt.Println(repo.Name + ": skip, no origin")
			continue
		}
		if gitx.Dirty(repo.Path) {
			fmt.Println(repo.Name + ": skip, dirty")
			continue
		}
		branch := gitx.Branch(repo.Path)
		fmt.Printf("%s: pull --rebase origin %s\n", repo.Name, branch)
		_, _ = gitx.Run(repo.Path, "pull", "--rebase", "origin", branch)
	}

	return nil
}
