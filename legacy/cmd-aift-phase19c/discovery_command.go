package main

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/discoveryengine"
)

func runDiscovery(cfg config.Config, args []string) error {
	if len(args) == 0 || args[0] == "scan" {
		return discoveryengine.Scan(cfg)
	}

	switch args[0] {
	case "scan":
		return discoveryengine.Scan(cfg)
	case "list":
		return discoveryengine.List(cfg)
	case "object":
		if len(args) < 2 {
			return fmt.Errorf("usage: aift discovery object <id-or-name>")
		}
		return discoveryengine.ObjectInfo(cfg, args[1])
	case "report":
		return discoveryengine.Report(cfg)
	default:
		return fmt.Errorf("usage: aift discovery scan|list|object|report")
	}
}
