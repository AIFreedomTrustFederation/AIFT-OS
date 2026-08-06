package uxi

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ForgeMission is the persisted Forge mission contract plus derived observation fields.
type ForgeMission struct {
	Schema            string           `json:"schema,omitempty"`
	ID                string           `json:"id,omitempty"`
	Title             string           `json:"title,omitempty"`
	AuthorityLevel    int              `json:"authorityLevel,omitempty"`
	State             string           `json:"state,omitempty"`
	TargetRepository  string           `json:"targetRepository,omitempty"`
	Risk              string           `json:"risk,omitempty"`
	Progress          int              `json:"progress,omitempty"`
	ComputedProgress  int              `json:"computedProgress,omitempty"`
	EstimatedTasks    int              `json:"estimatedTasks,omitempty"`
	EstimatedFiles    int              `json:"estimatedFiles,omitempty"`
	Scope             []string         `json:"scope,omitempty"`
	Limits            ForgeLimits      `json:"limits,omitempty"`
	Tasks             []ForgeTask      `json:"tasks,omitempty"`
	Approvals         []map[string]any `json:"approvals,omitempty"`
	UpdatedAt         string           `json:"updatedAt,omitempty"`
	Source            string           `json:"source"`
	ObservationStatus string           `json:"observation_status"`
}

// ForgeLimits declares the mission's allowed mutation boundary.
type ForgeLimits struct {
	MayCreateFiles         bool `json:"mayCreateFiles"`
	MayModifyExistingUI    bool `json:"mayModifyExistingUi"`
	MayModifyExistingAPI   bool `json:"mayModifyExistingApi"`
	MayDeleteFiles         bool `json:"mayDeleteFiles"`
	MayTouchOtherRepos     bool `json:"mayTouchOtherRepositories"`
	MayInstallDependencies bool `json:"mayInstallDependencies"`
	MayCommitAutomatically bool `json:"mayCommitAutomatically"`
}

// ForgeTask is one task persisted in a Forge mission.
type ForgeTask struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Status    string   `json:"status"`
	DependsOn []string `json:"dependsOn,omitempty"`
	Files     []string `json:"files,omitempty"`
}

// InspectForgeMission reads only persisted mission evidence and never invents defaults.
func InspectForgeMission(aiftRoot string) (ForgeMission, []Evidence, error) {
	path := filepath.Join(aiftRoot, "AIFT-Forge", ".forge", "mission.json")
	now := time.Now().UTC()
	mission := ForgeMission{Source: path, ObservationStatus: "missing"}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return mission, []Evidence{newEvidence("forge_mission", path, "No persisted Forge mission was observed", "", "missing", now)}, nil
	}
	if err != nil {
		return mission, nil, fmt.Errorf("read Forge mission: %w", err)
	}
	if err := json.Unmarshal(data, &mission); err != nil {
		mission = ForgeMission{Source: path, ObservationStatus: "blocked"}
		return mission, []Evidence{newEvidence("forge_mission", path, "Forge mission JSON is malformed", err.Error(), "failed", now)}, nil
	}
	mission.Source = path
	mission.ObservationStatus = "observed"
	mission.ComputedProgress = forgeProgress(mission.Tasks)
	evidence := []Evidence{newEvidence("forge_mission", path, fmt.Sprintf("Persisted Forge mission observed: %s", mission.Title), "", "observed", now)}
	if mission.Schema != "aift.forge.mission.v1" {
		evidence = append(evidence, newEvidence("schema", path, "Forge mission schema is missing or unexpected", mission.Schema, "warning", now))
	}
	if mission.Progress != mission.ComputedProgress {
		evidence = append(evidence, newEvidence("consistency", path, "Declared Forge mission progress differs from task evidence", fmt.Sprintf("declared=%d computed=%d", mission.Progress, mission.ComputedProgress), "warning", now))
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
