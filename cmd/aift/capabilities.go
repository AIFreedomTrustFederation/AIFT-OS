package main

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/capabilities"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func runCapabilities(cfg config.Config, args []string) error {
	if len(args) == 0 || args[0] == "scan" {
		return capabilities.Scan(cfg)
	}

	switch args[0] {
	case "scan":
		return capabilities.Scan(cfg)
	case "list":
		return capabilities.Scan(cfg)
	case "info":
		if len(args) < 2 {
			return fmt.Errorf("usage: aift capabilities info <id-or-name>")
		}
		return capabilities.PrintRepo(cfg, args[1])
	case "report":
		return capabilities.Report(cfg)
	default:
		return fmt.Errorf("usage: aift capabilities scan|list|info|report")
	}
}
