package runtime

import (
	"fmt"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/jobs"
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
	return jobs.RunAll(cfg)
}

func Loop(cfg config.Config) error {
	fmt.Println("AIFT-OS runtime loop started")
	return supervisor.New(cfg).Loop(5 * time.Minute)
}
