package uxi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildFederationWorldMapsVisibleLocation(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "AIFT-Runtime")
	if err := os.MkdirAll(filepath.Join(repoPath, ".aift"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":"aift.location.v1","label":"Sacramento, California","latitude":38.58157,"longitude":-121.4944,"precision":"city","visibility":"federation","source":"operator-declared"}`
	if err := os.WriteFile(filepath.Join(repoPath, ".aift", "location.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}

	world := BuildFederationWorld([]Repository{{
		ID: "aift-runtime", Name: "AIFT-Runtime", Path: repoPath, Role: "runtime-prototype", Status: "ready", Git: true,
	}})
	if world.Schema != "aift.federation.world.v1" || world.Progress.Mapped != 1 {
		t.Fatalf("world=%#v", world)
	}
	if len(world.Nodes) != 1 {
		t.Fatalf("nodes=%#v", world.Nodes)
	}
	node := world.Nodes[0]
	if node.Latitude != 38.58 || node.Longitude != -121.49 {
		t.Fatalf("coordinates=%v,%v", node.Latitude, node.Longitude)
	}
	if node.Evidence != ".aift/location.json" || node.XP != 150 {
		t.Fatalf("node=%#v", node)
	}
	if world.Progress.CompletedQuests != 2 || world.Progress.OpenQuests != 0 {
		t.Fatalf("progress=%#v quests=%#v", world.Progress, world.Quests)
	}
}

func TestBuildFederationWorldDoesNotExposePrivateCoordinates(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "Private")
	if err := os.MkdirAll(filepath.Join(repoPath, ".aift"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":"aift.location.v1","label":"Private place","latitude":12.34567,"longitude":45.67891,"precision":"exact","visibility":"private"}`
	if err := os.WriteFile(filepath.Join(repoPath, ".aift", "location.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}

	world := BuildFederationWorld([]Repository{{ID: "private", Name: "Private", Path: repoPath}})
	if len(world.Nodes) != 0 || world.Progress.Hidden != 1 {
		t.Fatalf("world=%#v", world)
	}
	if len(world.Unmapped) != 1 || world.Unmapped[0].State != "hidden" {
		t.Fatalf("unmapped=%#v", world.Unmapped)
	}
	if world.Progress.CompletedQuests != 1 || world.Progress.OpenQuests != 1 {
		t.Fatalf("progress=%#v quests=%#v", world.Progress, world.Quests)
	}
}

func TestBuildFederationWorldRejectsInvalidCoordinates(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "Invalid")
	if err := os.MkdirAll(filepath.Join(repoPath, ".aift"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":"aift.location.v1","label":"Invalid","latitude":100,"longitude":0,"precision":"city","visibility":"federation"}`
	if err := os.WriteFile(filepath.Join(repoPath, ".aift", "location.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}

	world := BuildFederationWorld([]Repository{{ID: "invalid", Name: "Invalid", Path: repoPath}})
	if world.Progress.Invalid != 1 || len(world.Nodes) != 0 {
		t.Fatalf("world=%#v", world)
	}
	if world.Unmapped[0].State != "invalid" {
		t.Fatalf("unmapped=%#v", world.Unmapped)
	}
}

func TestBuildFederationWorldRejectsMissingCoordinate(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "MissingCoordinate")
	if err := os.MkdirAll(filepath.Join(repoPath, ".aift"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":"aift.location.v1","label":"Incomplete","longitude":0,"precision":"city","visibility":"federation"}`
	if err := os.WriteFile(filepath.Join(repoPath, ".aift", "location.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	world := BuildFederationWorld([]Repository{{ID: "missing-coordinate", Name: "MissingCoordinate", Path: repoPath}})
	if world.Progress.Invalid != 1 || !strings.Contains(world.Unmapped[0].Reason, "latitude is required") {
		t.Fatalf("world=%#v", world)
	}
}

func TestBuildFederationWorldRejectsUnknownFields(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "Unknown")
	if err := os.MkdirAll(filepath.Join(repoPath, ".aift"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":"aift.location.v1","label":"Unknown","latitude":0,"longitude":0,"precision":"city","visibility":"federation","secret":"unexpected"}`
	if err := os.WriteFile(filepath.Join(repoPath, ".aift", "location.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	world := BuildFederationWorld([]Repository{{ID: "unknown", Name: "Unknown", Path: repoPath}})
	if world.Progress.Invalid != 1 || !strings.Contains(world.Unmapped[0].Reason, "unknown field") {
		t.Fatalf("world=%#v", world)
	}
}

func TestBuildFederationWorldRejectsOversizedDeclaration(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "Oversized")
	if err := os.MkdirAll(filepath.Join(repoPath, ".aift"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoPath, ".aift", "location.json"), []byte(strings.Repeat("x", maxLocationManifestBytes+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	world := BuildFederationWorld([]Repository{{ID: "oversized", Name: "Oversized", Path: repoPath}})
	if world.Progress.Invalid != 1 || !strings.Contains(world.Unmapped[0].Reason, "64 KiB") {
		t.Fatalf("world=%#v", world)
	}
}

func TestBuildFederationWorldRejectsInvalidUpdatedAt(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "InvalidTimestamp")
	if err := os.MkdirAll(filepath.Join(repoPath, ".aift"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":"aift.location.v1","label":"Invalid timestamp","latitude":0,"longitude":0,"visibility":"federation","updated_at":"yesterday"}`
	if err := os.WriteFile(filepath.Join(repoPath, ".aift", "location.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	world := BuildFederationWorld([]Repository{{ID: "invalid-timestamp", Name: "InvalidTimestamp", Path: repoPath}})
	if world.Progress.Invalid != 1 || !strings.Contains(world.Unmapped[0].Reason, "RFC3339") {
		t.Fatalf("world=%#v", world)
	}
}

func TestBuildFederationWorldCreatesMissingLocationQuests(t *testing.T) {
	world := BuildFederationWorld([]Repository{{ID: "missing", Name: "Missing", Path: t.TempDir()}})
	if world.Progress.Unmapped != 1 || world.Progress.OpenQuests != 2 {
		t.Fatalf("world=%#v", world)
	}
	if world.Quests[0].Status != "open" || world.Quests[1].Status != "open" {
		t.Fatalf("quests=%#v", world.Quests)
	}
}

func TestBuildFederationWorldUsesStableNameAndIDOrdering(t *testing.T) {
	root := t.TempDir()
	world := BuildFederationWorld([]Repository{
		{ID: "z", Name: "Shared", Path: filepath.Join(root, "z")},
		{ID: "a", Name: "Shared", Path: filepath.Join(root, "a")},
	})
	if world.Unmapped[0].RepositoryID != "a" || world.Unmapped[1].RepositoryID != "z" {
		t.Fatalf("order=%#v", world.Unmapped)
	}
}
