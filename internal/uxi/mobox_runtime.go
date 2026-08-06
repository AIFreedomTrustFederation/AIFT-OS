package uxi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MoboxRuntimeInspectAdapter observes the upstream-derived Windows
// compatibility runtime without invoking Wine, Box64, package managers, or games.
type MoboxRuntimeInspectAdapter struct {
	AIFTRoot string
}

func (a MoboxRuntimeInspectAdapter) Descriptor() AdapterDescriptor {
	return AdapterDescriptor{
		Kind: "mobox.runtime.inspect", Version: "v1",
		Description: "Inspect the Android Windows compatibility runtime without launching it.",
		Mutating: false, Risk: "low",
	}
}

func (a MoboxRuntimeInspectAdapter) Invoke(_ context.Context, request AdapterRequest) (AdapterResult, error) {
	repositoryPath := filepath.Join(a.AIFTRoot, "mobox")
	info, err := os.Stat(repositoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return AdapterResult{}, fmt.Errorf("mobox compatibility runtime repository not found under AIFT root")
		}
		return AdapterResult{}, err
	}
	if !info.IsDir() {
		return AdapterResult{}, fmt.Errorf("mobox compatibility runtime path is not a directory")
	}
	checks := []struct {
		name string
		path string
	}{
		{name: "git repository", path: filepath.Join(repositoryPath, ".git")},
		{name: "installer", path: filepath.Join(repositoryPath, "install")},
		{name: "runtime menu", path: filepath.Join(repositoryPath, "menu")},
		{name: "components", path: filepath.Join(repositoryPath, "components")},
	}
	observed := make(map[string]bool, len(checks))
	evidence := make([]Evidence, 0, len(checks)+1)
	for _, check := range checks {
		_, statErr := os.Stat(check.path)
		present := statErr == nil
		observed[check.name] = present
		status := "observed"
		detail := "present"
		if !present {
			status = "missing"
			detail = "not observed"
		}
		evidence = append(evidence, Evidence{
			ID: newID("evd"), Kind: "mobox_runtime", Source: check.path,
			Summary: check.name, Detail: detail, Status: status,
			ObservedAt: time.Now().UTC(),
		})
	}
	return AdapterResult{
		Summary: "MoBox Windows compatibility runtime inspected without execution",
		Evidence: evidence,
		Data: map[string]any{
			"runtime": "windows-compatibility",
			"repository": "AIFreedomTrustFederation/mobox",
			"path": repositoryPath,
			"target": request.Target,
			"observed": observed,
			"execution_performed": false,
			"boundary": "Wine and Box64 runtime; not the Federation Console or world truth authority",
		},
	}, nil
}
