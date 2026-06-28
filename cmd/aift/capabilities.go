package main

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/capabilityregistry"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func runCapabilities(cfg config.Config, args []string) error {
	if len(args) == 0 || args[0] == "scan" {
		return capabilityregistry.Scan(cfg)
	}

	switch args[0] {
	case "scan":
		return capabilityregistry.Scan(cfg)
	case "list":
		return capabilityregistry.List(cfg)
	case "info":
		if len(args) < 2 {
			return fmt.Errorf("usage: aift capabilities info <id-or-name>")
		}
		return capabilityregistry.Info(cfg, args[1])
	case "report":
		return capabilityregistry.Report(cfg)
	default:
		return fmt.Errorf("usage: aift capabilities scan|list|info|report")
	}
}
