package services

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/state"
)

type Service struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

func Defaults() []Service {
	return []Service{
		{Name: "api", Status: "available", Description: "Local HTTP API"},
		{Name: "scheduler", Status: "available", Description: "Federation scheduler"},
		{Name: "events", Status: "available", Description: "Event log and bus"},
		{Name: "providers", Status: "available", Description: "Provider registry"},
		{Name: "registry", Status: "available", Description: "Repository registry generator"},
		{Name: "reports", Status: "available", Description: "Dashboard and dependency reports"},
	}
}

func List(cfg config.Config) error {
	st, _ := state.Load(cfg)

	fmt.Printf("%-16s %-14s %s\n", "SERVICE", "STATUS", "DESCRIPTION")
	for _, svc := range Defaults() {
		status := svc.Status
		if st.Services != nil && st.Services[svc.Name] != "" {
			status = st.Services[svc.Name]
		}
		fmt.Printf("%-16s %-14s %s\n", svc.Name, status, svc.Description)
	}

	return nil
}

func Mark(cfg config.Config, name, statusValue string) error {
	st, _ := state.Load(cfg)
	if st.Services == nil {
		st.Services = map[string]string{}
	}
	st.Services[name] = statusValue
	if err := state.Save(cfg, st); err != nil {
		return err
	}
	return events.Emit(cfg, "service."+statusValue, name, "service state changed", map[string]string{"service": name, "status": statusValue})
}
