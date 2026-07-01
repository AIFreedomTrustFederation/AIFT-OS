package doctor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

var ErrShellUnavailable = errors.New("shell housekeeping unavailable")

func shellCommand(script string) (*exec.Cmd, error) {
	for _, name := range []string{"bash", "sh"} {
		if path, err := exec.LookPath(name); err == nil {
			return exec.Command(path, "-c", script), nil
		}
	}
	return nil, fmt.Errorf("%w: bash/sh not found on PATH", ErrShellUnavailable)
}

func Git(cfg config.Config) error {
	root := cfg.Root
	cmd, err := shellCommand(`find "` + root + `" -mindepth 2 -maxdepth 2 -type d -name .git | sort | while read gitdir; do repo=$(dirname "$gitdir"); echo "== $repo =="; git -C "$repo" status --short; done`)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Repair(cfg config.Config) error {
	fmt.Fprintln(os.Stdout, "Repairing generated runtime state")
	cmd, err := shellCommand(`find "` + cfg.Root + `" -mindepth 2 -maxdepth 2 -type d -name .git | sort | while read gitdir; do repo=$(dirname "$gitdir"); git -C "$repo" restore .aift/capabilities.json .aift/providers.json .aift/workflows.json .aift/repos.json var/events/events.jsonl 2>/dev/null || true; rm -f "$repo/.aift/module.json"; done`)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Full(cfg config.Config) error {
	if err := Repair(cfg); err != nil {
		return err
	}
	return Run(cfg)
}
