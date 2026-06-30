package runtime

import (
	"fmt"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/jobs"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/services"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/state"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/supervisor"
)

func StartOnce(cfg config.Config) error {
	if err := events.Emit(cfg, "runtime.start", "runtime", "runtime one-shot start", nil); err != nil {
		return err
	}
	if err := supervisor.New(cfg).Boot(); err != nil {
		return err
	}
	fmt.Println("AIFT-OS runtime completed one-shot start")
	return events.Emit(cfg, "runtime.complete", "runtime", "runtime one-shot complete", nil)
}

func Tick(cfg config.Config) error {
	if err := services.Reconcile(cfg); err != nil {
		return err
	}
	return jobs.RunAll(cfg)
}

func Loop(cfg config.Config) error {
	fmt.Println("AIFT-OS runtime loop started")
	return supervisor.New(cfg).Loop(5 * time.Minute)
}

func Status(cfg config.Config) error {
	st, _ := state.Load(cfg)
	fmt.Printf("runtime: %s\n", st.Status)
	fmt.Printf("started: %s\n", st.StartedAt)
	fmt.Printf("updated: %s\n", st.UpdatedAt)
	fmt.Printf("services: %d\n", len(st.Catalog))
	return nil
}

func Ready(cfg config.Config) error {
	st, _ := state.Load(cfg)
	for _, svc := range st.Catalog {
		if svc.Status != "ready" && svc.Status != "running" {
			fmt.Printf("not-ready: %s %s %s\n", svc.Name, svc.Repo, svc.Status)
		}
	}
	return nil
}

func Blocked(cfg config.Config) error {
	st, _ := state.Load(cfg)
	for _, svc := range st.Catalog {
		if svc.Status == "blocked" || svc.Status == "error" || svc.Status == "failed" {
			fmt.Printf("blocked: %s %s %s\n", svc.Name, svc.Repo, svc.Status)
		}
	}
	return nil
}

func Report(cfg config.Config) error {
	return services.List(cfg)
}
