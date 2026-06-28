package eventmesh

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Topic struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Evidence    string `json:"evidence"`
}

type Subscriber struct {
	Repo     string `json:"repo"`
	Topic    string `json:"topic"`
	Status   string `json:"status"`
	Handler  string `json:"handler,omitempty"`
	Evidence string `json:"evidence"`
}

type EventContract struct {
	Repo        string       `json:"repo"`
	Publishes   []Topic      `json:"publishes"`
	Subscribes  []Subscriber `json:"subscribes"`
	GeneratedAt string       `json:"generatedAt"`
}

type Registry struct {
	GeneratedAt string          `json:"generatedAt"`
	Topics      []Topic         `json:"topics"`
	Subscribers []Subscriber    `json:"subscribers"`
	Contracts   []EventContract `json:"contracts"`
}

func InitAll(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	for _, r := range repos {
		if err := InitRepo(r.Name, r.Path); err != nil {
			return err
		}
	}

	return Scan(cfg)
}

func InitRepo(name, repoPath string) error {
	dir := filepath.Join(repoPath, ".aift")
	if err := os.MkdirAll(filepath.Join(dir, "events", "handlers"), 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, "events.json")
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	contract := EventContract{
		Repo: name,
		Publishes: []Topic{
			{Name: "repo.changed", Status: "planned", Description: "Repository content changed", Evidence: "default event contract"},
			{Name: "capability.changed", Status: "planned", Description: "Capability status changed", Evidence: "default event contract"},
			{Name: "manual.changed", Status: "planned", Description: "Manual source changed", Evidence: "default event contract"},
		},
		Subscribes:  []Subscriber{},
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0644)
}

func Scan(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	topicMap := map[string]Topic{}
	var subscribers []Subscriber
	var contracts []EventContract

	for _, r := range repos {
		c, ok := readContract(r.Name, r.Path)
		if !ok {
			continue
		}
		contracts = append(contracts, c)

		for _, t := range c.Publishes {
			if t.Status == "" {
				t.Status = "planned"
			}
			if t.Evidence == "" {
				t.Evidence = ".aift/events.json"
			}
			topicMap[t.Name] = t
		}

		for _, s := range c.Subscribes {
			if s.Status == "" {
				s.Status = "planned"
			}
			if s.Evidence == "" {
				s.Evidence = ".aift/events.json"
			}
			subscribers = append(subscribers, s)
		}
	}

	topics := make([]Topic, 0, len(topicMap))
	for _, t := range topicMap {
		topics = append(topics, t)
	}
	sort.Slice(topics, func(i, j int) bool { return topics[i].Name < topics[j].Name })

	sort.Slice(subscribers, func(i, j int) bool {
		if subscribers[i].Topic == subscribers[j].Topic {
			return subscribers[i].Repo < subscribers[j].Repo
		}
		return subscribers[i].Topic < subscribers[j].Topic
	})

	reg := Registry{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Topics:      topics,
		Subscribers: subscribers,
		Contracts:   contracts,
	}

	if err := writeRegistry(cfg, reg); err != nil {
		return err
	}
	if err := writeReport(cfg, reg); err != nil {
		return err
	}

	return events.Emit(cfg, "eventmesh.scan", "eventmesh", "event mesh scanned", map[string]string{
		"topics":      fmt.Sprint(len(topics)),
		"subscribers": fmt.Sprint(len(subscribers)),
	})
}

func readContract(name, repoPath string) (EventContract, bool) {
	path := filepath.Join(repoPath, ".aift", "events.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return EventContract{}, false
	}

	var c EventContract
	if json.Unmarshal(data, &c) != nil {
		return EventContract{}, false
	}
	if c.Repo == "" {
		c.Repo = name
	}
	return c, true
}

func Publish(cfg config.Config, topic string, source string, message string) error {
	if topic == "" {
		return fmt.Errorf("topic is required")
	}
	if source == "" {
		source = "manual"
	}
	if message == "" {
		message = topic
	}

	return events.Emit(cfg, topic, source, message, map[string]string{
		"eventMesh": "true",
	})
}

func Tail(cfg config.Config, n int) error {
	return events.Tail(cfg, n)
}

func Topics(cfg config.Config) error {
	reg, err := loadOrScan(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("%-32s %-12s %s\n", "TOPIC", "STATUS", "DESCRIPTION")
	for _, t := range reg.Topics {
		fmt.Printf("%-32s %-12s %s\n", t.Name, t.Status, t.Description)
	}
	return nil
}

func Subscribers(cfg config.Config) error {
	reg, err := loadOrScan(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("%-32s %-28s %-12s %s\n", "TOPIC", "REPO", "STATUS", "HANDLER")
	for _, s := range reg.Subscribers {
		fmt.Printf("%-32s %-28s %-12s %s\n", s.Topic, s.Repo, s.Status, s.Handler)
	}
	return nil
}

func Replay(cfg config.Config, topic string) error {
	path := filepath.Join(cfg.OSHome, "var", "events", "events.jsonl")
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if topic == "" || strings.Contains(line, `"type":"`+topic+`"`) || strings.Contains(line, `"topic":"`+topic+`"`) {
			fmt.Println(line)
		}
	}
	return scanner.Err()
}

func Report(cfg config.Config) error {
	path := filepath.Join(cfg.OSHome, "reports", "event-mesh.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if err := Scan(cfg); err != nil {
			return err
		}
		data, err = os.ReadFile(path)
		if err != nil {
			return err
		}
	}
	fmt.Print(string(data))
	return nil
}

func loadOrScan(cfg config.Config) (Registry, error) {
	path := filepath.Join(cfg.OSHome, "registry", "event-mesh.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if err := Scan(cfg); err != nil {
			return Registry{}, err
		}
		data, err = os.ReadFile(path)
		if err != nil {
			return Registry{}, err
		}
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return Registry{}, err
	}
	return reg, nil
}

func writeRegistry(cfg config.Config, reg Registry) error {
	out := filepath.Join(cfg.OSHome, "registry", "event-mesh.json")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(out, append(data, '\n'), 0644); err != nil {
		return err
	}
	fmt.Println("Wrote", out)
	return nil
}

func writeReport(cfg config.Config, reg Registry) error {
	out := filepath.Join(cfg.OSHome, "reports", "event-mesh.md")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# Federation Event Mesh\n\n")
	b.WriteString("The event mesh describes asynchronous federation coordination.\n\n")
	b.WriteString("## Topics\n\n")
	b.WriteString("| Topic | Status | Description | Evidence |\n")
	b.WriteString("|---|---|---|---|\n")
	for _, t := range reg.Topics {
		b.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s |\n", t.Name, t.Status, t.Description, t.Evidence))
	}

	b.WriteString("\n## Subscribers\n\n")
	b.WriteString("| Topic | Repository | Status | Handler | Evidence |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, s := range reg.Subscribers {
		b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | %s |\n", s.Topic, s.Repo, s.Status, s.Handler, s.Evidence))
	}

	return os.WriteFile(out, []byte(b.String()), 0644)
}
