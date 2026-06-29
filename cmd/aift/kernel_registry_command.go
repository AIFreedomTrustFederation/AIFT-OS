package main

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernelregistry"
)

func runKernelRegistry(cfg config.Config, args []string) error {
	if len(args) == 0 || args[0] == "scan" {
		return kernelregistry.Scan(cfg)
	}

	switch args[0] {
	case "scan":
		return kernelregistry.Scan(cfg)
	case "list":
		return kernelregistry.List(cfg)
	case "object":
		if len(args) < 2 {
			return fmt.Errorf("usage: aift kernel-registry object <id-or-name>")
		}
		return kernelregistry.ObjectInfo(cfg, args[1])
	case "report":
		return kernelregistry.Report(cfg)
	default:
		return fmt.Errorf("usage: aift kernel-registry scan|list|object|report")
	}
}
