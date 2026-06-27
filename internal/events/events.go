package events

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type Event struct {
	Time    string            `json:"time"`
	Type    string            `json:"type"`
	Source  string            `json:"source"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data,omitempty"`
}

func Emit(cfg config.Config, eventType, source, message string, data map[string]string) error {
	event := Event{
		Time:    time.Now().Format(time.RFC3339),
		Type:    eventType,
		Source:  source,
		Message: message,
		Data:    data,
	}

	dir := filepath.Join(cfg.OSHome, "var", "events")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, "events.jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(f, string(b))
	return err
}

func Tail(cfg config.Config, limit int) error {
	path := filepath.Join(cfg.OSHome, "var", "events", "events.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("No events yet.")
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		fmt.Println("No events yet.")
		return nil
	}

	if limit > 0 && len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}

	for _, line := range lines {
		fmt.Println(line)
	}

	return nil
}
