package daemon

import (
	"fmt"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/api"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/providers"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/scheduler"
)

func Start(cfg config.Config, addr string) error {
	if err := events.Emit(cfg, "daemon.start", "daemon", "AIFT-OS daemon starting", map[string]string{"addr": addr}); err != nil {
		return err
	}

	if err := providers.WriteRegistry(cfg); err != nil {
		return err
	}

	s := scheduler.New(cfg)
	if err := s.RunOnce(); err != nil {
		return err
	}

	go func() {
		_ = s.Loop(5 * time.Minute)
	}()

	fmt.Println("AIFT-OS daemon started")
	fmt.Println("API:", addr)
	fmt.Println("Press CTRL+C to stop.")

	return api.New(cfg, addr).Serve()
}
