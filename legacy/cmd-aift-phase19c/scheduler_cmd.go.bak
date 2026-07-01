package main

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/schedulerplan"
)

func runScheduler(cfg config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: aift scheduler plan|ready|blocked|report")
	}

	switch args[0] {
	case "plan":
		return schedulerplan.GeneratePlan(cfg)
	case "ready":
		return schedulerplan.PrintReady(cfg)
	case "blocked":
		return schedulerplan.PrintBlocked(cfg)
	case "report":
		return schedulerplan.PrintReport(cfg)
	default:
		return fmt.Errorf("usage: aift scheduler plan|ready|blocked|report")
	}
}
