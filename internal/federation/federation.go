package federation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/providers"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/repo"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workflow"
)

type Snapshot struct {
	Repos          int `json:"repos"`
	Dirty          int `json:"dirty"`
	ValidManifests int `json:"validManifests"`
}

func Scan(cfg config.Config) error {
	if err := manifests.EnsureAll(cfg); err != nil {
		return err
	}
	if err := repo.EnsureExampleCommand(cfg); err != nil {
		return err
	}
	if err := providers.WriteRegistry(cfg); err != nil {
		return err
	}
	if err := workflow.WriteRegistry(cfg); err != nil {
		return err
	}
	if err := registry.Generate(cfg); err != nil {
		return err
	}
	if err := reports.Dashboard(cfg); err != nil {
		return err
	}
	if err := reports.Deps(cfg); err != nil {
		return err
	}

	snap, err := SnapshotState(cfg)
	if err != nil {
		return err
	}

	out := filepath.Join(cfg.OSHome, "registry", "federation-snapshot.json")
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(out, append(data, '\n'), 0644); err != nil {
		return err
	}

	if err := events.Emit(cfg, "federation.scan", "federation", "federation scan complete", map[string]string{
		"repos": fmt.Sprint(snap.Repos),
		"dirty": fmt.Sprint(snap.Dirty),
	}); err != nil {
		return err
	}

	fmt.Println("Wrote", out)
	return nil
}

func SnapshotState(cfg config.Config) (Snapshot, error) {
	repos, err := repo.List(cfg)
	if err != nil {
		return Snapshot{}, err
	}

	var snap Snapshot
	snap.Repos = len(repos)
	for _, r := range repos {
		if r.Dirty {
			snap.Dirty++
		}
		if r.ManifestValid {
			snap.ValidManifests++
		}
	}

	return snap, nil
}

func Verify(cfg config.Config) error {
	if err := Scan(cfg); err != nil {
		return err
	}
	snap, err := SnapshotState(cfg)
	if err != nil {
		return err
	}
	fmt.Printf("Federation verified: repos=%d dirty=%d validManifests=%d\n", snap.Repos, snap.Dirty, snap.ValidManifests)
	return nil
}

func Graph(cfg config.Config) error {
	if err := reports.Deps(cfg); err != nil {
		return err
	}
	fmt.Println("Wrote", filepath.Join(cfg.OSHome, "reports", "dependency-graph.md"))
	return nil
}
