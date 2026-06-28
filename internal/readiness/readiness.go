package readiness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/fsutil"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/jsonfile"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

// Status values for readiness objects.
const (
	StatusPlanned    = "planned"
	StatusDetected   = "detected"
	StatusReady      = "ready"
	StatusActive     = "active"
	StatusBlocked    = "blocked"
	StatusDeprecated = "deprecated"
	StatusRemoved    = "removed"
)

// ValidStatuses lists all valid status values.
var ValidStatuses = []string{
	StatusPlanned, StatusDetected, StatusReady,
	StatusActive, StatusBlocked, StatusDeprecated, StatusRemoved,
}

// Object represents a single evaluated object in the federation.
type Object struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
	Source   string `json:"source"`
}

// Registry holds the full readiness state.
type Registry struct {
	GeneratedAt string   `json:"generatedAt"`
	Objects     []Object `json:"objects"`
	Summary     Summary  `json:"summary"`
}

// Summary counts objects by status.
type Summary struct {
	Total      int            `json:"total"`
	ByStatus   map[string]int `json:"byStatus"`
	ByKind     map[string]int `json:"byKind"`
	ReadyCount int            `json:"readyCount"`
}

// IsValid returns true if the status is a recognized value.
func IsValid(status string) bool {
	for _, s := range ValidStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// Transition validates whether a status change from old to new is allowed.
// Allowed transitions:
//
//	planned   -> detected, ready, removed
//	detected  -> ready, blocked, deprecated, removed
//	ready     -> active, blocked, deprecated
//	active    -> blocked, deprecated
//	blocked   -> detected, ready, removed
//	deprecated -> removed
//	removed   -> (terminal)
func Transition(old, new string) error {
	if !IsValid(old) {
		return fmt.Errorf("invalid current status: %q", old)
	}
	if !IsValid(new) {
		return fmt.Errorf("invalid target status: %q", new)
	}
	if old == new {
		return nil
	}

	allowed := map[string][]string{
		StatusPlanned:    {StatusDetected, StatusReady, StatusRemoved},
		StatusDetected:   {StatusReady, StatusBlocked, StatusDeprecated, StatusRemoved},
		StatusReady:      {StatusActive, StatusBlocked, StatusDeprecated},
		StatusActive:     {StatusBlocked, StatusDeprecated},
		StatusBlocked:    {StatusDetected, StatusReady, StatusRemoved},
		StatusDeprecated: {StatusRemoved},
		StatusRemoved:    {},
	}

	for _, target := range allowed[old] {
		if target == new {
			return nil
		}
	}
	return fmt.Errorf("invalid transition: %s -> %s", old, new)
}

// Scan evaluates all discoverable objects and builds the readiness registry.
func Scan(cfg config.Config) error {
	var objects []Object

	objects = append(objects, scanRepositories(cfg)...)
	objects = append(objects, scanModules(cfg)...)
	objects = append(objects, scanCapabilities(cfg)...)
	objects = append(objects, scanServices(cfg)...)
	objects = append(objects, scanEvents(cfg)...)
	objects = append(objects, scanCommands(cfg)...)
	objects = append(objects, scanScripts(cfg)...)

	sort.Slice(objects, func(i, j int) bool {
		if objects[i].Kind != objects[j].Kind {
			return objects[i].Kind < objects[j].Kind
		}
		return objects[i].ID < objects[j].ID
	})

	reg := Registry{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Objects:     objects,
		Summary:     summarize(objects),
	}

	if err := writeRegistry(cfg, reg); err != nil {
		return err
	}
	if err := writeReport(cfg, reg); err != nil {
		return err
	}

	return events.Emit(cfg, "runtime.readiness.scan", "readiness",
		fmt.Sprintf("readiness scan: %d objects", len(objects)),
		map[string]string{"total": fmt.Sprint(len(objects))})
}

// Status prints a summary of all objects grouped by status.
func Status(cfg config.Config) error {
	reg, err := loadOrScan(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("%-14s %-24s %-12s %s\n", "KIND", "NAME", "STATUS", "EVIDENCE")
	for _, obj := range reg.Objects {
		fmt.Printf("%-14s %-24s %-12s %s\n", obj.Kind, obj.Name, obj.Status, obj.Evidence)
	}
	return nil
}

// Ready prints only objects with ready or active status.
func Ready(cfg config.Config) error {
	reg, err := loadOrScan(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("%-14s %-24s %-12s %s\n", "KIND", "NAME", "STATUS", "EVIDENCE")
	for _, obj := range reg.Objects {
		if obj.Status == StatusReady || obj.Status == StatusActive {
			fmt.Printf("%-14s %-24s %-12s %s\n", obj.Kind, obj.Name, obj.Status, obj.Evidence)
		}
	}
	return nil
}

// Blocked prints only objects with blocked status.
func Blocked(cfg config.Config) error {
	reg, err := loadOrScan(cfg)
	if err != nil {
		return err
	}

	count := 0
	fmt.Printf("%-14s %-24s %-12s %s\n", "KIND", "NAME", "STATUS", "EVIDENCE")
	for _, obj := range reg.Objects {
		if obj.Status == StatusBlocked {
			fmt.Printf("%-14s %-24s %-12s %s\n", obj.Kind, obj.Name, obj.Status, obj.Evidence)
			count++
		}
	}
	if count == 0 {
		fmt.Println("No blocked objects.")
	}
	return nil
}

// Report prints the readiness report markdown.
func Report(cfg config.Config) error {
	path := filepath.Join(cfg.OSHome, "reports", "runtime-readiness.md")
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

// ── Scanners ────────────────────────────────────────────────────────

func scanRepositories(cfg config.Config) []Object {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return nil
	}

	var objects []Object
	for _, r := range repos {
		status := StatusDetected
		evidence := ".git directory"

		if fsutil.Exists(filepath.Join(r.Path, ".aift", "repo.json")) {
			status = StatusReady
			evidence = ".aift/repo.json manifest"
		} else if fsutil.Exists(filepath.Join(r.Path, "go.mod")) || fsutil.Exists(filepath.Join(r.Path, "package.json")) {
			status = StatusDetected
			evidence = "build file (go.mod or package.json)"
		}

		objects = append(objects, Object{
			ID:       "repo:" + r.Name,
			Kind:     "repository",
			Name:     r.Name,
			Status:   status,
			Evidence: evidence,
			Source:   r.Path,
		})
	}
	return objects
}

func scanModules(cfg config.Config) []Object {
	path := filepath.Join(cfg.OSHome, "registry", "modules.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var reg struct {
		Modules []struct {
			Repo   string `json:"repo"`
			Name   string `json:"name"`
			Kind   string `json:"kind"`
			Status string `json:"status"`
		} `json:"modules"`
	}
	if json.Unmarshal(data, &reg) != nil {
		return nil
	}

	var objects []Object
	for _, m := range reg.Modules {
		status := mapStatus(m.Status)
		objects = append(objects, Object{
			ID:       "module:" + m.Name,
			Kind:     "module",
			Name:     m.Name,
			Status:   status,
			Evidence: "registry/modules.json",
			Source:   m.Repo,
		})
	}
	return objects
}

func scanCapabilities(cfg config.Config) []Object {
	path := filepath.Join(cfg.OSHome, "registry", "capabilities.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var reg struct {
		Repos []struct {
			Repo         string `json:"repo"`
			Capabilities []struct {
				Name   string `json:"name"`
				Status string `json:"status"`
			} `json:"capabilities"`
		} `json:"repos"`
	}
	if json.Unmarshal(data, &reg) != nil {
		return nil
	}

	var objects []Object
	for _, r := range reg.Repos {
		for _, c := range r.Capabilities {
			status := mapStatus(c.Status)
			objects = append(objects, Object{
				ID:       "capability:" + r.Repo + ":" + c.Name,
				Kind:     "capability",
				Name:     r.Repo + "/" + c.Name,
				Status:   status,
				Evidence: "registry/capabilities.json",
				Source:   r.Repo,
			})
		}
	}
	return objects
}

func scanServices(cfg config.Config) []Object {
	path := filepath.Join(cfg.OSHome, "registry", "service-contracts.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var reg struct {
		Services []struct {
			Repo     string `json:"repo"`
			Name     string `json:"name"`
			Status   string `json:"status"`
			Evidence string `json:"evidence"`
		} `json:"services"`
	}
	if json.Unmarshal(data, &reg) != nil {
		return nil
	}

	var objects []Object
	for _, s := range reg.Services {
		status := mapStatus(s.Status)
		evidence := s.Evidence
		if evidence == "" {
			evidence = "registry/service-contracts.json"
		}
		objects = append(objects, Object{
			ID:       "service:" + s.Name,
			Kind:     "service",
			Name:     s.Name,
			Status:   status,
			Evidence: evidence,
			Source:   s.Repo,
		})
	}
	return objects
}

func scanEvents(cfg config.Config) []Object {
	path := filepath.Join(cfg.OSHome, "var", "events", "event-bus.jsonl")
	if !fsutil.Exists(path) {
		return nil
	}

	// Just check if the event bus exists and has content
	info, err := os.Stat(path)
	if err != nil || info.Size() == 0 {
		return []Object{{
			ID:       "event:event-bus",
			Kind:     "event",
			Name:     "event-bus",
			Status:   StatusPlanned,
			Evidence: "event-bus.jsonl is empty",
			Source:   "var/events",
		}}
	}

	return []Object{{
		ID:       "event:event-bus",
		Kind:     "event",
		Name:     "event-bus",
		Status:   StatusActive,
		Evidence: fmt.Sprintf("event-bus.jsonl (%d bytes)", info.Size()),
		Source:   "var/events",
	}}
}

func scanCommands(cfg config.Config) []Object {
	// Read architecture.json to get command data
	path := filepath.Join(cfg.OSHome, "registry", "architecture.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var arch struct {
		Commands []struct {
			Name       string `json:"name"`
			HasHandler bool   `json:"has_handler"`
			HasHelp    bool   `json:"has_help"`
			Status     string `json:"status"`
		} `json:"commands"`
	}
	if json.Unmarshal(data, &arch) != nil {
		return nil
	}

	var objects []Object
	for _, c := range arch.Commands {
		status := StatusReady
		evidence := "case in main.go switch + help entry"
		if c.Status == "planned" {
			status = StatusPlanned
			evidence = "planned command stub"
		} else if !c.HasHandler {
			status = StatusBlocked
			evidence = "listed in help but no handler"
		} else if !c.HasHelp {
			status = StatusDetected
			evidence = "handler exists but not in help"
		}

		objects = append(objects, Object{
			ID:       "command:" + c.Name,
			Kind:     "command",
			Name:     c.Name,
			Status:   status,
			Evidence: evidence,
			Source:   "cmd/aift",
		})
	}
	return objects
}

func scanScripts(cfg config.Config) []Object {
	scriptsDir := filepath.Join(cfg.OSHome, "scripts")
	if !fsutil.Exists(scriptsDir) {
		return nil
	}

	var objects []Object
	filepath.Walk(scriptsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".sh") {
			return nil
		}

		rel, _ := filepath.Rel(cfg.OSHome, path)
		name := filepath.Base(path)
		status := StatusDetected
		evidence := "shell script exists"

		// Check if it sources the harness
		data, err := os.ReadFile(path)
		if err == nil {
			content := string(data)
			if strings.Contains(content, "aift-run.sh") {
				status = StatusReady
				evidence = "sources aift-run.sh harness"
			} else if strings.Contains(content, "# no-harness:") || strings.Contains(content, "# harness-exempt:") {
				status = StatusReady
				evidence = "explicitly exempt from harness"
			}
		}

		objects = append(objects, Object{
			ID:       "script:" + name,
			Kind:     "script",
			Name:     name,
			Status:   status,
			Evidence: evidence,
			Source:   rel,
		})
		return nil
	})
	return objects
}

// ── Helpers ─────────────────────────────────────────────────────────

func mapStatus(s string) string {
	switch s {
	case "ready", "v1":
		return StatusReady
	case "detected":
		return StatusDetected
	case "planned":
		return StatusPlanned
	case "broken":
		return StatusBlocked
	case "active":
		return StatusActive
	case "deprecated":
		return StatusDeprecated
	case "removed":
		return StatusRemoved
	default:
		if s == "" {
			return StatusPlanned
		}
		return StatusDetected
	}
}

func summarize(objects []Object) Summary {
	s := Summary{
		Total:    len(objects),
		ByStatus: map[string]int{},
		ByKind:   map[string]int{},
	}
	for _, obj := range objects {
		s.ByStatus[obj.Status]++
		s.ByKind[obj.Kind]++
		if obj.Status == StatusReady || obj.Status == StatusActive {
			s.ReadyCount++
		}
	}
	return s
}

func writeRegistry(cfg config.Config, reg Registry) error {
	return jsonfile.Write(filepath.Join(cfg.OSHome, "registry", "runtime-readiness.json"), reg, true)
}

func writeReport(cfg config.Config, reg Registry) error {
	out := filepath.Join(cfg.OSHome, "reports", "runtime-readiness.md")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# Runtime Readiness Report\n\n")
	b.WriteString(fmt.Sprintf("Generated: %s\n\n", reg.GeneratedAt))

	// Summary
	b.WriteString("## Summary\n\n")
	b.WriteString(fmt.Sprintf("- **Total objects**: %d\n", reg.Summary.Total))
	b.WriteString(fmt.Sprintf("- **Ready/Active**: %d\n", reg.Summary.ReadyCount))
	b.WriteString("\n")

	// By status
	b.WriteString("### By Status\n\n")
	for _, status := range ValidStatuses {
		if count, ok := reg.Summary.ByStatus[status]; ok {
			b.WriteString(fmt.Sprintf("- **%s**: %d\n", status, count))
		}
	}
	b.WriteString("\n")

	// By kind
	b.WriteString("### By Kind\n\n")
	kinds := sortedKeys(reg.Summary.ByKind)
	for _, kind := range kinds {
		b.WriteString(fmt.Sprintf("- **%s**: %d\n", kind, reg.Summary.ByKind[kind]))
	}
	b.WriteString("\n")

	// Full table
	b.WriteString("## All Objects\n\n")
	b.WriteString("| Kind | Name | Status | Evidence | Source |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, obj := range reg.Objects {
		b.WriteString(fmt.Sprintf("| %s | `%s` | %s | %s | %s |\n",
			obj.Kind, obj.Name, obj.Status, obj.Evidence, obj.Source))
	}

	return os.WriteFile(out, []byte(b.String()), 0644)
}

func loadOrScan(cfg config.Config) (Registry, error) {
	path := filepath.Join(cfg.OSHome, "registry", "runtime-readiness.json")
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

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
