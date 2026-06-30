package doctor

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func Git(cfg config.Config) error {
	root := cfg.Root
	cmd := exec.Command("sh", "-c", `find "`+root+`" -mindepth 2 -maxdepth 2 -type d -name .git | sort | while read gitdir; do repo=$(dirname "$gitdir"); echo "== $repo =="; git -C "$repo" status --short; done`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Repair(cfg config.Config) error {
	fmt.Fprintln(os.Stdout, "Repairing generated runtime state")
	cmd := exec.Command("sh", "-c", `find "`+cfg.Root+`" -mindepth 2 -maxdepth 2 -type d -name .git | sort | while read gitdir; do repo=$(dirname "$gitdir"); git -C "$repo" restore .aift/capabilities.json .aift/providers.json .aift/workflows.json .aift/repos.json var/events/events.jsonl 2>/dev/null || true; rm -f "$repo/.aift/module.json"; done`)
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
