package jobs

import (
	"fmt"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/providers"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports"
)

type Job struct {
	Name string
	Run  func(config.Config) error
}

func Defaults() []Job {
	return []Job{
		{Name: "providers", Run: providers.WriteRegistry},
		{Name: "registry", Run: registry.Generate},
		{Name: "dashboard", Run: reports.Dashboard},
		{Name: "deps", Run: reports.Deps},
	}
}

func RunAll(cfg config.Config) error {
	start := time.Now()
	if err := events.Emit(cfg, "jobs.start", "jobs", "job batch started", nil); err != nil {
		return err
	}

	for _, job := range Defaults() {
		if err := events.Emit(cfg, "job.start", job.Name, "job started", nil); err != nil {
			return err
		}
		if err := job.Run(cfg); err != nil {
			_ = events.Emit(cfg, "job.error", job.Name, err.Error(), nil)
			return err
		}
		if err := events.Emit(cfg, "job.complete", job.Name, "job completed", nil); err != nil {
			return err
		}
	}

	return events.Emit(cfg, "jobs.complete", "jobs", fmt.Sprintf("job batch completed in %s", time.Since(start)), nil)
}
