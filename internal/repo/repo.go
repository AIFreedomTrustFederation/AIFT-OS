package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/gitx"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Info struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	Branch        string `json:"branch"`
	Remote        string `json:"remote"`
	Dirty         bool   `json:"dirty"`
	ManifestPath  string `json:"manifestPath"`
	ManifestValid bool   `json:"manifestValid"`
	CommandsPath  string `json:"commandsPath"`
	WorkflowsPath string `json:"workflowsPath"`
}

func List(cfg config.Config) ([]Info, error) {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return nil, err
	}

	out := make([]Info, 0, len(repos))
	for _, r := range repos {
		out = append(out, InspectRepo(r))
	}

	return out, nil
}

func Inspect(cfg config.Config, name string) (Info, error) {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return Info{}, err
	}

	for _, r := range repos {
		if r.Name == name || filepath.Base(r.Path) == name {
			return InspectRepo(r), nil
		}
	}

	return Info{}, fmt.Errorf("repository not found: %s", name)
}

func InspectRepo(r workspace.Repo) Info {
	return Info{
		Name:          r.Name,
		Path:          r.Path,
		Branch:        gitx.Branch(r.Path),
		Remote:        gitx.Remote(r.Path),
		Dirty:         gitx.Dirty(r.Path),
		ManifestPath:  manifests.Path(r.Path),
		ManifestValid: manifests.Valid(r.Path),
		CommandsPath:  filepath.Join(r.Path, ".aift", "commands"),
		WorkflowsPath: filepath.Join(r.Path, ".aift", "workflows.json"),
	}
}

func PrintList(cfg config.Config) error {
	repos, err := List(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("%-32s %-12s %-8s %-8s %s\n", "REPOSITORY", "BRANCH", "STATE", "MANIFEST", "REMOTE")
	for _, r := range repos {
		state := "clean"
		if r.Dirty {
			state = "dirty"
		}
		manifest := "valid"
		if !r.ManifestValid {
			manifest = "missing"
		}
		fmt.Printf("%-32s %-12s %-8s %-8s %s\n", r.Name, r.Branch, state, manifest, r.Remote)
	}

	return nil
}

func PrintInspect(cfg config.Config, name string) error {
	r, err := Inspect(cfg, name)
	if err != nil {
		return err
	}

	fmt.Println("Repository:", r.Name)
	fmt.Println("Path:", r.Path)
	fmt.Println("Branch:", r.Branch)
	fmt.Println("Remote:", r.Remote)
	fmt.Println("Dirty:", r.Dirty)
	fmt.Println("Manifest:", r.ManifestPath)
	fmt.Println("Manifest valid:", r.ManifestValid)
	fmt.Println("Commands:", r.CommandsPath)
	fmt.Println("Workflows:", r.WorkflowsPath)

	return nil
}

func RunCommand(cfg config.Config, name string, commandName string, args []string) error {
	r, err := Inspect(cfg, name)
	if err != nil {
		return err
	}

	script := filepath.Join(r.CommandsPath, commandName+".sh")
	if _, err := os.Stat(script); err != nil {
		return fmt.Errorf("repo command not found: %s", script)
	}

	cmd := exec.Command("sh", append([]string{script}, args...)...)
	cmd.Dir = r.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(),
		"AIFT_REPO_NAME="+r.Name,
		"AIFT_REPO_PATH="+r.Path,
	)

	return cmd.Run()
}

func EnsureExampleCommand(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	for _, r := range repos {
		dir := filepath.Join(r.Path, ".aift", "commands")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		script := filepath.Join(dir, "status.sh")
		if _, err := os.Stat(script); err == nil {
			continue
		}
		content := "#!/usr/bin/env sh\nset -eu\necho \"AIFT repo: ${AIFT_REPO_NAME:-unknown}\"\ngit status --short\n"
		if err := os.WriteFile(script, []byte(content), 0755); err != nil {
			return err
		}
	}

	return nil
}

func NormalizeName(s string) string {
	return strings.TrimSpace(s)
}
