package scheduler

import (
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports"
)

type Scheduler struct {
	Config config.Config
}

func New(cfg config.Config) Scheduler {
	return Scheduler{Config: cfg}
}

func (s Scheduler) RunOnce() error {
	if err := events.Emit(s.Config, "scheduler.tick", "scheduler", "scheduler tick started", nil); err != nil {
		return err
	}

	if err := registry.Generate(s.Config); err != nil {
		return err
	}

	if err := reports.Dashboard(s.Config); err != nil {
		return err
	}

	if err := reports.Deps(s.Config); err != nil {
		return err
	}

	return events.Emit(s.Config, "scheduler.tick.complete", "scheduler", "scheduler tick completed", map[string]string{
		"interval": "manual",
	})
}

func (s Scheduler) Loop(interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if err := s.RunOnce(); err != nil {
		return err
	}

	for range ticker.C {
		if err := s.RunOnce(); err != nil {
			_ = events.Emit(s.Config, "scheduler.error", "scheduler", err.Error(), nil)
		}
	}

	return nil
}
