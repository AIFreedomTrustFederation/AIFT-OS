package supervisor

import (
	"fmt"
	"os"
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
	st, _ := state.Load(s.Config)
	st.Name = "AIFT-OS"
	st.Status = "running"

	if err := services.Reconcile(s.Config); err != nil {
		return err
	}

	discovered, err := services.Discover(s.Config)
	if err != nil {
		return err
	}

	st, _ = state.Load(s.Config)
	for _, svc := range discovered {
		if st.Services[svc.Name] == "" || st.Services[svc.Name] == "available" || st.Services[svc.Name] == "discovered" {
			st.Services[svc.Name] = "ready"
		}
	}

	if err := state.Save(s.Config, st); err != nil {
		return err
	}

	if err := events.Emit(s.Config, "supervisor.boot", "supervisor", "runtime supervisor booted from discovered service catalog", map[string]string{
		"services": fmt.Sprintf("%d", len(discovered)),
	}); err != nil {
		return err
	}

	return jobs.RunAll(s.Config)
}

func (s Supervisor) Tick() error {
	if err := services.Reconcile(s.Config); err != nil {
		return err
	}
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
		if err := s.Tick(); err != nil {
			fmt.Fprintf(os.Stderr, "supervisor tick error: %v\n", err)
		}
	}

	return nil
}
