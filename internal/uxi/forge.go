package uxi

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ForgeMission struct {
	Schema            string           `json:"schema,omitempty"`
	ID                string           `json:"id,omitempty"`
	Title             string           `json:"title,omitempty"`
	AuthorityLevel    int              `json:"authority_level,omitempty"`
	State             string           `json:"state,omitempty"`
	TargetRepository  string           `json:"target_repository,omitempty"`
	Risk              string           `json:"risk,omitempty"`
	Progress          int              `json:"progress,omitempty"`
	ComputedProgress  int              `json:"computed_progress,omitempty"`
	EstimatedTasks    int              `json:"estimated_tasks,omitempty"`
	EstimatedFiles    int              `json:"estimated_files,omitempty"`
	Scope             []string         `json:"scope,omitempty"`
	Limits            ForgeLimits      `json:"limits,omitempty"`
	Tasks             []ForgeTask      `json:"tasks,omitempty"`
	Approvals         []map[string]any `json:"approvals,omitempty"`
	UpdatedAt         string           `json:"updated_at,omitempty"`
	Source            string           `json:"source"`
	ObservationStatus string           `json:"observation_status"`
}

type ForgeLimits struct {
	MayCreateFiles         bool `json:"mayCreateFiles"`
	MayModifyExistingUI    bool `json:"mayModifyExistingUi"`
	MayModifyExistingAPI   bool `json:"mayModifyExistingApi"`
	MayDeleteFiles         bool `json:"mayDeleteFiles"`
	MayTouchOtherRepos     bool `json:"mayTouchOtherRepositories"`
	MayInstallDependencies bool `json:"mayInstallDependencies"`
	MayCommitAutomatically bool `json:"mayCommitAutomatically"`
}

type ForgeTask struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Status    string   `json:"status"`
	DependsOn []string `json:"dependsOn,omitempty"`
	Files     []string `json:"files,omitempty"`
}

func InspectForgeMission(aiftRoot string) (ForgeMission, []Evidence, error) {
	path := filepath.Join(aiftRoot, "AIFT-Forge", ".forge", "mission.json")
	now := time.Now().UTC()
	mission := ForgeMission{Source: path, ObservationStatus: "missing"}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return mission, []Evidence{{
			ID: newID("evd"), Kind: "forge_mission", Source: path,
			Summary: "No persisted Forge mission was observed", Status: "missing", ObservedAt: now,
		}}, nil
	}
	if err != nil {
		return mission, nil, fmt.Errorf("read Forge mission: %w", err)
	}
	if err := json.Unmarshal(data, &mission); err != nil {
		mission = ForgeMission{Source: path, ObservationStatus: "blocked"}
		return mission, []Evidence{{
			ID: newID("evd"), Kind: "forge_mission", Source: path,
			Summary: "Forge mission JSON is malformed", Detail: err.Error(), Status: "failed", ObservedAt: now,
		}}, nil
	}
	mission.Source = path
	mission.ObservationStatus = "observed"
	mission.ComputedProgress = forgeProgress(mission.Tasks)
	evidence := []Evidence{{
		ID: newID("evd"), Kind: "forge_mission", Source: path,
		Summary: fmt.Sprintf("Persisted Forge mission observed: %s", mission.Title), Status: "observed", ObservedAt: now,
	}}
	if mission.Schema != "aift.forge.mission.v1" {
		evidence = append(evidence, Evidence{
			ID: newID("evd"), Kind: "schema", Source: path,
			Summary: "Forge mission schema is missing or unexpected", Detail: mission.Schema, Status: "warning", ObservedAt: now,
		})
	}
	if mission.Progress != mission.ComputedProgress {
		evidence = append(evidence, Evidence{
			ID: newID("evd"), Kind: "consistency", Source: path,
			Summary: "Declared Forge mission progress differs from task evidence",
			Detail:  fmt.Sprintf("declared=%d computed=%d", mission.Progress, mission.ComputedProgress), Status: "warning", ObservedAt: now,
		})
	}
	return mission, evidence, nil
}

func forgeProgress(tasks []ForgeTask) int {
	if len(tasks) == 0 {
		return 0
	}
	complete := 0
	for _, task := range tasks {
		if task.Status == "complete" || task.Status == "succeeded" {
			complete++
		}
	}
	return (complete*100 + len(tasks)/2) / len(tasks)
}
