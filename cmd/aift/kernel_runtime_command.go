package main

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernelruntime"
)

func runKernelRuntime(cfg config.Config, args []string) error {
	if len(args) == 0 || args[0] == "boot" {
		return kernelruntime.Boot(cfg)
	}

	switch args[0] {
	case "boot":
		return kernelruntime.Boot(cfg)
	case "status":
		return kernelruntime.Status(cfg)
	case "report":
		return kernelruntime.Report(cfg)
	default:
		return fmt.Errorf("usage: aift kernel boot|status|report")
	}
}
