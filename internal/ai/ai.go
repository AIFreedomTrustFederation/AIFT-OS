package ai

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/repair"
)

func Run(cfg config.Config, args []string) error {
	if len(args) == 0 || args[0] == "doctor" {
		return Doctor(cfg)
	}

	switch args[0] {
	case "repair":
		safe := false
		for _, arg := range args[1:] {
			if arg == "--safe" || arg == "safe" {
				safe = true
			}
		}
		return repair.Run(cfg, safe)
	case "verify":
		return repair.Verify(cfg)
	case "report":
		return repair.Report(cfg)
	case "help", "-h", "--help":
		return Doctor(cfg)
	default:
		return fmt.Errorf("unknown ai command: %s", args[0])
	}
}

func Doctor(cfg config.Config) error {
	fmt.Println("AIFT Self-Repair AI Kernel")
	fmt.Println("mode: conservative")
	fmt.Println("truth: inspect first")
	fmt.Println("commands:")
	fmt.Println("  aift ai doctor")
	fmt.Println("  aift ai repair --safe")
	fmt.Println("  aift ai verify")
	fmt.Println("  aift ai report")
	return nil
}
