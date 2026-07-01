package main

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/readiness"
)

func runRuntime(cfg config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: aift runtime scan|status|ready|blocked|report")
	}

	switch args[0] {
	case "scan":
		return readiness.Scan(cfg)
	case "status":
		return readiness.Status(cfg)
	case "ready":
		return readiness.Ready(cfg)
	case "blocked":
		return readiness.Blocked(cfg)
	case "report":
		return readiness.Report(cfg)
	default:
		return fmt.Errorf("usage: aift runtime scan|status|ready|blocked|report")
	}
}
