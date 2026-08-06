package uxi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectForgeMissionDoesNotInventDefault(t *testing.T) {
	mission, evidence, err := InspectForgeMission(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if mission.ObservationStatus != "missing" {
		t.Fatalf("mission=%#v", mission)
	}
	if len(evidence) != 1 || evidence[0].Status != "missing" {
		t.Fatalf("evidence=%#v", evidence)
	}
}

func TestInspectForgeMissionComputesProgress(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "AIFT-Forge", ".forge")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	data := `{"schema":"aift.forge.mission.v1","id":"m1","title":"Build UXI","state":"awaiting-approval","targetRepository":"AIFT-OS","risk":"low","progress":99,"authorityLevel":2,"estimatedTasks":2,"estimatedFiles":4,"updatedAt":"2026-08-05T00:00:00Z","tasks":[{"id":"one","title":"One","status":"complete"},{"id":"two","title":"Two","status":"ready"}]}`
	if err := os.WriteFile(filepath.Join(path, "mission.json"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	mission, evidence, err := InspectForgeMission(root)
	if err != nil {
		t.Fatal(err)
	}
	if mission.ComputedProgress != 50 {
		t.Fatalf("progress=%d", mission.ComputedProgress)
	}
	if mission.TargetRepository != "AIFT-OS" || mission.AuthorityLevel != 2 || mission.EstimatedTasks != 2 || mission.EstimatedFiles != 4 || mission.UpdatedAt == "" {
		t.Fatalf("mission=%#v", mission)
	}
	if len(evidence) != 2 {
		t.Fatalf("evidence=%#v", evidence)
	}
}
