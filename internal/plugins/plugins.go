package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

func List(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	fmt.Println("AIFT plugin commands:")
	for _, repo := range repos {
		dir := filepath.Join(repo.Path, ".aift", "commands")
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".sh") {
				fmt.Printf("%s :: %s\n", repo.Name, strings.TrimSuffix(file.Name(), ".sh"))
			}
		}
	}

	return nil
}
