package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func Run(cfg config.Config) error {
	fmt.Println("AIFT-OS Doctor")
	fmt.Println("root:", cfg.Root)
	fmt.Println("os:  ", cfg.OSHome)

	required := []string{
		"cmd/aift",
		"internal/config",
		"internal/workspace",
		"internal/gitx",
		"internal/doctor",
		"internal/registry",
		"internal/manifests",
		"internal/reports",
		"internal/plugins",
		"internal/sync",
		"internal/kernel",
		"install",
		"tests",
		"docs",
		"schemas",
		"registry",
		"reports",
		"bin",
	}

	for _, dir := range required {
		full := filepath.Join(cfg.OSHome, dir)
		info, err := os.Stat(full)
		if err != nil || !info.IsDir() {
			return fmt.Errorf("missing directory: %s", dir)
		}
	}

	fmt.Println("OK: AIFT-OS Go kernel layout healthy")
	return nil
}
