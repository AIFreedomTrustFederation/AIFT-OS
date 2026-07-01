package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/readiness"
)

func runOperator(cfg config.Config, args []string) error {
	if len(args) == 0 || args[0] != "check" {
		return fmt.Errorf("usage: aift operator check")
	}

	return operatorCheck(cfg)
}

func operatorCheck(cfg config.Config) error {
	type step struct {
		name string
		fn   func() error
	}

	steps := []step{
		{"verify", func() error { return verify(cfg) }},
		{"architecture", func() error { return runArchitectureCheck(cfg) }},
		{"runtime scan", func() error { return readiness.Scan(cfg) }},
	}

	var failed []string

	for _, s := range steps {
		fmt.Printf("--- %s ---\n", s.name)
		if err := s.fn(); err != nil {
			fmt.Printf("FAIL: %s: %v\n", s.name, err)
			failed = append(failed, s.name)
		} else {
			fmt.Printf("OK: %s\n", s.name)
		}
		fmt.Println()
	}

	// Print readiness summary
	fmt.Println("--- readiness summary ---")
	if err := readiness.Status(cfg); err != nil {
		fmt.Printf("FAIL: readiness summary: %v\n", err)
		failed = append(failed, "readiness summary")
	}
	fmt.Println()

	if len(failed) > 0 {
		return fmt.Errorf("operator check failed: %s", strings.Join(failed, ", "))
	}

	fmt.Println("OK: operator check passed")
	return nil
}

func runArchitectureCheck(cfg config.Config) error {
	cmd := exec.Command("go", "run", "./tools/architecture", "--ci")
	cmd.Dir = cfg.OSHome
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("architecture check failed: %s", string(out))
	}
	fmt.Print(string(out))
	return nil
}
