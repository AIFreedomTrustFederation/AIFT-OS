package main

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/patchengine"
)

func runPatchEngine(cfg config.Config, args []string) error {
	if len(args) == 0 || args[0] == "inspect" {
		return patchengine.Inspect(cfg)
	}

	switch args[0] {
	case "inspect":
		return patchengine.Inspect(cfg)
	case "plan":
		return patchengine.PlanCommand(cfg)
	case "validate":
		return patchengine.Validate(cfg)
	default:
		return fmt.Errorf("usage: aift patch-engine inspect|plan|validate")
	}
}
