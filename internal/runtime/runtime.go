package runtime

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/providers"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports"
)

func StartOnce(cfg config.Config) error {
	if err := events.Emit(cfg, "runtime.start", "runtime", "runtime one-shot start", nil); err != nil {
		return err
	}

	if err := providers.WriteRegistry(cfg); err != nil {
		return err
	}

	if err := registry.Generate(cfg); err != nil {
		return err
	}

	if err := reports.Dashboard(cfg); err != nil {
		return err
	}

	if err := reports.Deps(cfg); err != nil {
		return err
	}

	fmt.Println("AIFT-OS runtime completed one-shot start")
	return events.Emit(cfg, "runtime.complete", "runtime", "runtime one-shot complete", nil)
}
