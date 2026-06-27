package supervisor

import (
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/jobs"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/services"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/state"
)

type Supervisor struct {
	Config config.Config
}

func New(cfg config.Config) Supervisor {
	return Supervisor{Config: cfg}
}

func (s Supervisor) Boot() error {
	st := state.New()
	for _, svc := range services.Defaults() {
		st.Services[svc.Name] = "running"
	}
	if err := state.Save(s.Config, st); err != nil {
		return err
	}
	if err := events.Emit(s.Config, "supervisor.boot", "supervisor", "runtime supervisor booted", nil); err != nil {
		return err
	}
	return jobs.RunAll(s.Config)
}

func (s Supervisor) Tick() error {
	if err := events.Emit(s.Config, "supervisor.tick", "supervisor", "runtime supervisor tick", nil); err != nil {
		return err
	}
	return jobs.RunAll(s.Config)
}

func (s Supervisor) Loop(interval time.Duration) error {
	if err := s.Boot(); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		_ = s.Tick()
	}

	return nil
}
