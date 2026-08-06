package uxi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildFederationWorldMapsVisibleLocation(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "AIFT-Runtime")
	writeLocationManifest(t, repoPath, `{"schema":"aift.location.v1","label":"Sacramento, California","latitude":38.58157,"longitude":-121.4944,"precision":"city","visibility":"federation","source":"operator-declared"}`)

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
	if node.Evidence != ".aift/location.json" || node.XP != 100 || world.Progress.XP != 100 {
		t.Fatalf("node=%#v progress=%#v", node, world.Progress)
	}
	if world.Progress.CompletedQuests != 1 || world.Progress.OpenQuests != 0 {
		t.Fatalf("progress=%#v quests=%#v", world.Progress, world.Quests)
	}
}

func TestBuildFederationWorldDoesNotExposePrivateCoordinates(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "Private")
	writeLocationManifest(t, repoPath, `{"schema":"aift.location.v1","label":"Private place","latitude":12.34567,"longitude":45.67891,"precision":"exact","visibility":"private"}`)

	world := BuildFederationWorld([]Repository{{ID: "private", Name: "Private", Path: repoPath}})
	if len(world.Nodes) != 0 || world.Progress.Hidden != 1 {
		t.Fatalf("world=%#v", world)
	}
	if len(world.Unmapped) != 1 || world.Unmapped[0].State != "hidden" {
		t.Fatalf("unmapped=%#v", world.Unmapped)
	}
	if world.Progress.XP != 100 || world.Progress.CompletedQuests != 1 || world.Progress.OpenQuests != 0 {
		t.Fatalf("progress=%#v quests=%#v", world.Progress, world.Quests)
	}
}

func TestBuildFederationWorldVisibilityDoesNotChangeXP(t *testing.T) {
	root := t.TempDir()
	privatePath := filepath.Join(root, "Private")
	publicPath := filepath.Join(root, "Public")
	writeLocationManifest(t, privatePath, `{"schema":"aift.location.v1","label":"Private","latitude":1,"longitude":1,"visibility":"private"}`)
	writeLocationManifest(t, publicPath, `{"schema":"aift.location.v1","label":"Public","latitude":2,"longitude":2,"visibility":"public"}`)

	privateWorld := BuildFederationWorld([]Repository{{ID: "private", Name: "Private", Path: privatePath}})
	publicWorld := BuildFederationWorld([]Repository{{ID: "public", Name: "Public", Path: publicPath}})
	if privateWorld.Progress.XP != publicWorld.Progress.XP || privateWorld.Progress.XP != 100 {
		t.Fatalf("private=%#v public=%#v", privateWorld.Progress, publicWorld.Progress)
	}
}

func TestBuildFederationWorldRejectsInvalidCoordinates(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "Invalid")
	writeLocationManifest(t, repoPath, `{"schema":"aift.location.v1","label":"Invalid","latitude":100,"longitude":0,"precision":"city","visibility":"federation"}`)

	world := BuildFederationWorld([]Repository{{ID: "invalid", Name: "Invalid", Path: repoPath}})
	if world.Progress.Invalid != 1 || len(world.Nodes) != 0 {
		t.Fatalf("world=%#v", world)
	}
	if world.Unmapped[0].State != "invalid" || world.Progress.XP != 0 {
		t.Fatalf("unmapped=%#v progress=%#v", world.Unmapped, world.Progress)
	}
}

func TestBuildFederationWorldRejectsMissingCoordinate(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "MissingCoordinate")
	writeLocationManifest(t, repoPath, `{"schema":"aift.location.v1","label":"Incomplete","longitude":0,"precision":"city","visibility":"federation"}`)
	world := BuildFederationWorld([]Repository{{ID: "missing-coordinate", Name: "MissingCoordinate", Path: repoPath}})
	if world.Progress.Invalid != 1 || !strings.Contains(world.Unmapped[0].Reason, "latitude is required") {
		t.Fatalf("world=%#v", world)
	}
}

func TestBuildFederationWorldRejectsUnknownFields(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "Unknown")
	writeLocationManifest(t, repoPath, `{"schema":"aift.location.v1","label":"Unknown","latitude":0,"longitude":0,"precision":"city","visibility":"federation","secret":"unexpected"}`)
	world := BuildFederationWorld([]Repository{{ID: "unknown", Name: "Unknown", Path: repoPath}})
	if world.Progress.Invalid != 1 || !strings.Contains(world.Unmapped[0].Reason, "unknown field") {
		t.Fatalf("world=%#v", world)
	}
}

func TestBuildFederationWorldRejectsOversizedDeclaration(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "Oversized")
	writeLocationManifest(t, repoPath, strings.Repeat("x", maxLocationManifestBytes+1))
	world := BuildFederationWorld([]Repository{{ID: "oversized", Name: "Oversized", Path: repoPath}})
	if world.Progress.Invalid != 1 || !strings.Contains(world.Unmapped[0].Reason, "64 KiB") {
		t.Fatalf("world=%#v", world)
	}
}

func TestBuildFederationWorldRejectsInvalidUpdatedAt(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "InvalidTimestamp")
	writeLocationManifest(t, repoPath, `{"schema":"aift.location.v1","label":"Invalid timestamp","latitude":0,"longitude":0,"visibility":"federation","updated_at":"yesterday"}`)
	world := BuildFederationWorld([]Repository{{ID: "invalid-timestamp", Name: "InvalidTimestamp", Path: repoPath}})
	if world.Progress.Invalid != 1 || !strings.Contains(world.Unmapped[0].Reason, "RFC3339") {
		t.Fatalf("world=%#v", world)
	}
}

func TestBuildFederationWorldCreatesMissingLocationQuest(t *testing.T) {
	world := BuildFederationWorld([]Repository{{ID: "missing", Name: "Missing", Path: t.TempDir()}})
	if world.Progress.Unmapped != 1 || world.Progress.OpenQuests != 1 || world.Progress.CompletedQuests != 0 {
		t.Fatalf("world=%#v", world)
	}
	if len(world.Quests) != 1 || world.Quests[0].Status != "open" {
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

func writeLocationManifest(t *testing.T, repoPath, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(repoPath, ".aift"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoPath, ".aift", "location.json"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
