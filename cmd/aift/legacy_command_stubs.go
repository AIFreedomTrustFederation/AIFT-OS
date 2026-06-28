package main

import (
	"fmt"
	"strings"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func runIntelligence(cfg config.Config, args []string) error {
	return plannedCommand("intelligence", args)
}

func runManual(cfg config.Config, args []string) error {
	return plannedCommand("manual", args)
}

func runMesh(cfg config.Config, args []string) error {
	return plannedCommand("mesh", args)
}

func runServiceContracts(cfg config.Config, args []string) error {
	return plannedCommand("service-contracts", args)
}

func runPlanner(cfg config.Config, args []string) error {
	return plannedCommand("planner", args)
}

func plannedCommand(name string, args []string) error {
	detail := strings.Join(args, " ")
	if detail == "" {
		detail = "no subcommand"
	}
	return fmt.Errorf("%s command is planned but not active yet: %s", name, detail)
}
