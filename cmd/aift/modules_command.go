package main

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/modules"
)

func runModules(cfg config.Config, args []string) error {
	if len(args) == 0 || args[0] == "scan" {
		return modules.Scan(cfg)
	}

	switch args[0] {
	case "init-all":
		return modules.InitAll(cfg)
	case "scan":
		return modules.Scan(cfg)
	case "list":
		return modules.List(cfg)
	case "repo":
		if len(args) < 2 {
			return fmt.Errorf("usage: aift modules repo <repo>")
		}
		return modules.Repo(cfg, args[1])
	case "report":
		return modules.Report(cfg)
	default:
		return fmt.Errorf("usage: aift modules init-all|scan|list|repo|report")
	}
}
