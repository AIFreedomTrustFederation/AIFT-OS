package main

import (
	"fmt"
	"strings"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/eventbus"
)

func runEventBus(cfg config.Config, args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return eventbus.List(cfg)
	}

	switch args[0] {
	case "publish":
		if len(args) < 3 {
			return fmt.Errorf("usage: aift event-bus publish <topic> <message> [key=value...]")
		}
		payload := map[string]string{}
		for _, item := range args[3:] {
			parts := strings.SplitN(item, "=", 2)
			if len(parts) == 2 {
				payload[parts[0]] = parts[1]
			}
		}
		return eventbus.Publish(cfg, args[1], "manual", "aiftd", args[2], payload)
	case "list":
		return eventbus.List(cfg)
	case "replay":
		topic := ""
		if len(args) > 1 {
			topic = args[1]
		}
		return eventbus.Replay(cfg, topic)
	case "report":
		return eventbus.Report(cfg)
	default:
		return fmt.Errorf("usage: aift event-bus publish|list|replay|report")
	}
}
