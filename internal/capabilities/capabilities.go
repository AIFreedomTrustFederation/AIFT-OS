package capabilities

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/fsutil"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/jsonfile"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

const (
	StatusPlanned  = "planned"
	StatusDetected = "detected"
	StatusReady    = "ready"
	StatusV1       = "v1"
	StatusBroken   = "broken"
	StatusMissing  = "missing"
)

type Capability struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Version     int    `json:"version"`
	Command     string `json:"command,omitempty"`
	Evidence    string `json:"evidence,omitempty"`
	Description string `json:"description,omitempty"`
	LastChecked string `json:"lastChecked"`
}

type RepoCapabilities struct {
	Repo         string       `json:"repo"`
	Path         string       `json:"path"`
	Capabilities []Capability `json:"capabilities"`
}

type FederationCapabilities struct {
	GeneratedAt string             `json:"generatedAt"`
	Repos       []RepoCapabilities `json:"repos"`
}

func capabilityNames() []string {
	return []string{
		"status",
		"verify",
		"test",
		"build",
		"start",
		"stop",
		"health",
		"deploy",
		"sync",
		"docs",
	}
}

func Scan(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	all := FederationCapabilities{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Repos:       []RepoCapabilities{},
	}

	for _, r := range repos {
		rc, err := ScanRepo(cfg, r)
		if err != nil {
			return err
		}
		all.Repos = append(all.Repos, rc)
	}

	if err := writeGlobal(cfg, all); err != nil {
		return err
	}

	if err := writeReport(cfg, all); err != nil {
		return err
	}

	return events.Emit(cfg, "capabilities.scan", "capabilities", "federation capability scan complete", map[string]string{
		"repos": fmt.Sprint(len(all.Repos)),
	})
}

func ScanRepo(cfg config.Config, r workspace.Repo) (RepoCapabilities, error) {
	now := time.Now().Format(time.RFC3339)
	old := readExisting(r.Path)

	rc := RepoCapabilities{
		Repo:         r.Name,
		Path:         r.Path,
		Capabilities: []Capability{},
	}

	for _, name := range capabilityNames() {
		prev := old[name]
		cap := detectCapability(r.Path, name)
		cap.LastChecked = now

		if prev.Status == StatusV1 {
			if cap.Status == StatusReady || cap.Status == StatusDetected {
				if cap.Command != "" && commandPasses(r.Path, cap.Command) {
					cap.Status = StatusV1
					cap.Version = 1
					cap.Evidence = "previously promoted to v1 and verification still passes"
				} else if cap.Command != "" {
					cap.Status = StatusBroken
					cap.Version = 1
					cap.Evidence = "was v1, but current verification failed"
				}
			}
		}

		rc.Capabilities = append(rc.Capabilities, cap)
	}

	if err := writeRepo(r.Path, rc); err != nil {
		return rc, err
	}

	return rc, nil
}

func detectCapability(repoPath, name string) Capability {
	c := Capability{
		Name:        name,
		Status:      StatusPlanned,
		Version:     0,
		Description: description(name),
	}

	cmdPath := filepath.Join(repoPath, ".aift", "commands", name+".sh")
	if fsutil.FileExists(cmdPath) {
		c.Command = ".aift/commands/" + name + ".sh"
		if commandPasses(repoPath, cmdPath) {
			c.Status = StatusReady
			c.Evidence = "command exists and passes local verification"
		} else {
			c.Status = StatusBroken
			c.Evidence = "command exists but failed local verification"
		}
		return c
	}

	switch name {
	case "test":
		if fsutil.FileExists(filepath.Join(repoPath, "package.json")) {
			c.Status = StatusDetected
			c.Evidence = "package.json detected; test capability may exist but no .aift command is proven"
			return c
		}
		if fsutil.FileExists(filepath.Join(repoPath, "go.mod")) {
			c.Status = StatusDetected
			c.Evidence = "go.mod detected; Go tests may exist but no .aift command is proven"
			return c
		}
	case "build":
		if fsutil.FileExists(filepath.Join(repoPath, "package.json")) || fsutil.FileExists(filepath.Join(repoPath, "go.mod")) || fsutil.FileExists(filepath.Join(repoPath, "Makefile")) {
			c.Status = StatusDetected
			c.Evidence = "build-related project file detected but no .aift build command is proven"
			return c
		}
	case "docs":
		if fsutil.FileExists(filepath.Join(repoPath, "README.md")) || fsutil.DirExists(filepath.Join(repoPath, "docs")) {
			c.Status = StatusDetected
			c.Evidence = "README/docs detected"
			return c
		}
	case "status":
		if fsutil.DirExists(filepath.Join(repoPath, ".git")) {
			c.Status = StatusReady
			c.Command = "git status --short"
			c.Evidence = "git repository detected; built-in status capability is ready"
			return c
		}
	case "sync":
		if fsutil.DirExists(filepath.Join(repoPath, ".git")) {
			c.Status = StatusReady
			c.Command = "git remote/status"
			c.Evidence = "git repository detected; safe sync can inspect this repo"
			return c
		}
	}

	c.Evidence = "not proven yet"
	return c
}

func Promote(cfg config.Config, repoName, capName string) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	for _, r := range repos {
		if r.Name != repoName {
			continue
		}

		rc, err := ScanRepo(cfg, r)
		if err != nil {
			return err
		}

		changed := false
		for i := range rc.Capabilities {
			if rc.Capabilities[i].Name != capName {
				continue
			}

			if rc.Capabilities[i].Status != StatusReady {
				return fmt.Errorf("cannot promote %s/%s: status is %s, not ready", repoName, capName, rc.Capabilities[i].Status)
			}

			rc.Capabilities[i].Status = StatusV1
			rc.Capabilities[i].Version = 1
			rc.Capabilities[i].Evidence = "promoted to v1 after passing local verification"
			rc.Capabilities[i].LastChecked = time.Now().Format(time.RFC3339)
			changed = true
		}

		if !changed {
			return fmt.Errorf("capability not found: %s", capName)
		}

		if err := writeRepo(r.Path, rc); err != nil {
			return err
		}

		if err := Scan(cfg); err != nil {
			return err
		}

		return events.Emit(cfg, "capability.promote", "capabilities", "capability promoted to v1", map[string]string{
			"repo":       repoName,
			"capability": capName,
		})
	}

	return fmt.Errorf("repository not found: %s", repoName)
}

func Report(cfg config.Config) error {
	data, err := os.ReadFile(filepath.Join(cfg.OSHome, "reports", "capabilities.md"))
	if err != nil {
		if err := Scan(cfg); err != nil {
			return err
		}
		data, err = os.ReadFile(filepath.Join(cfg.OSHome, "reports", "capabilities.md"))
		if err != nil {
			return err
		}
	}
	fmt.Print(string(data))
	return nil
}

func PrintRepo(cfg config.Config, repoName string) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	for _, r := range repos {
		if r.Name != repoName {
			continue
		}
		rc, err := ScanRepo(cfg, r)
		if err != nil {
			return err
		}
		printRepo(rc)
		return nil
	}

	return fmt.Errorf("repository not found: %s", repoName)
}

func printRepo(rc RepoCapabilities) {
	fmt.Println("Repository:", rc.Repo)
	fmt.Println("Path:", rc.Path)
	fmt.Printf("%-14s %-10s %-8s %s\n", "CAPABILITY", "STATUS", "VERSION", "EVIDENCE")
	for _, c := range rc.Capabilities {
		fmt.Printf("%-14s %-10s %-8d %s\n", c.Name, c.Status, c.Version, c.Evidence)
	}
}

func writeRepo(repoPath string, rc RepoCapabilities) error {
	return jsonfile.Write(filepath.Join(repoPath, ".aift", "capabilities.json"), rc, false)
}

func writeGlobal(cfg config.Config, all FederationCapabilities) error {
	return jsonfile.Write(filepath.Join(cfg.OSHome, "registry", "capabilities.json"), all, true)
}

func writeReport(cfg config.Config, all FederationCapabilities) error {
	out := filepath.Join(cfg.OSHome, "reports", "capabilities.md")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# Federation Capabilities\n\n")
	b.WriteString("Statuses: `planned`, `detected`, `ready`, `v1`, `broken`, `missing`.\n\n")

	for _, repo := range all.Repos {
		b.WriteString("## " + repo.Repo + "\n\n")
		b.WriteString("| Capability | Status | Version | Evidence |\n")
		b.WriteString("|---|---|---:|---|\n")
		for _, c := range repo.Capabilities {
			b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%d` | %s |\n", c.Name, c.Status, c.Version, c.Evidence))
		}
		b.WriteString("\n")
	}

	if err := os.WriteFile(out, []byte(b.String()), 0644); err != nil {
		return err
	}
	fmt.Println("Wrote", out)
	return nil
}

func readExisting(repoPath string) map[string]Capability {
	out := map[string]Capability{}
	data, err := os.ReadFile(filepath.Join(repoPath, ".aift", "capabilities.json"))
	if err != nil {
		return out
	}

	var rc RepoCapabilities
	if json.Unmarshal(data, &rc) != nil {
		return out
	}

	for _, c := range rc.Capabilities {
		out[c.Name] = c
	}
	return out
}

func commandPasses(repoPath, commandPath string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if strings.Contains(commandPath, " ") {
		cmd = exec.CommandContext(ctx, "sh", "-c", commandPath)
	} else {
		cmd = exec.CommandContext(ctx, "sh", commandPath)
	}
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "AIFT_CAPABILITY_CHECK=1")
	return cmd.Run() == nil
}

func description(name string) string {
	switch name {
	case "status":
		return "Report repository state"
	case "verify":
		return "Validate repository health"
	case "test":
		return "Run test suite"
	case "build":
		return "Build project artifacts"
	case "start":
		return "Start local service"
	case "stop":
		return "Stop local service"
	case "health":
		return "Check local service health"
	case "deploy":
		return "Deploy project"
	case "sync":
		return "Synchronize safely"
	case "docs":
		return "Documentation present or generated"
	default:
		return "Capability"
	}
}


