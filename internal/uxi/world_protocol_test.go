package uxi

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
)

func TestWorldSnapshotIsDeterministicAndPrecise(t *testing.T) {
	repositories := []Repository{{
		ID: "aift-os", Name: "AIFT-OS", Role: "federation-kernel",
		Status: "ready", Git: true,
		Capabilities: []Capability{{Name: "inspect", Status: "ready"}},
		Evidence: []Evidence{{ID: "evidence-1", Kind: "filesystem"}},
	}}
	first := BuildWorldSnapshot(repositories)
	second := BuildWorldSnapshot(repositories)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("world snapshot changed without evidence changes")
	}
	if first.Schema != "aift.world.v1" || first.WorldID != "aift-federation" {
		t.Fatalf("unexpected identity: %#v", first)
	}
	if len(first.Entities) != 2 || len(first.Relations) != 1 || len(first.Chunks) != 1 {
		t.Fatalf("unexpected topology: entities=%d relations=%d chunks=%d", len(first.Entities), len(first.Relations), len(first.Chunks))
	}
	if first.Revision == 0 || first.Chunks[0].ContentHash == "" {
		t.Fatal("snapshot is missing revision or content hash")
	}
	for _, coordinate := range []string{first.Entities[0].Position.X, first.Entities[0].Position.Y, first.Entities[0].Position.Z} {
		if _, err := strconv.ParseInt(coordinate, 10, 64); err != nil {
			t.Fatalf("coordinate %q is not an int64 decimal: %v", coordinate, err)
		}
	}
	payload, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) == "" {
		t.Fatal("snapshot did not encode")
	}
}

func TestWorldSnapshotRevisionChangesWithEvidence(t *testing.T) {
	repository := Repository{ID: "repo", Name: "Repo", Status: "detected"}
	before := BuildWorldSnapshot([]Repository{repository})
	repository.Evidence = []Evidence{{ID: "observed", Kind: "test"}}
	after := BuildWorldSnapshot([]Repository{repository})
	if before.Revision == after.Revision {
		t.Fatal("revision did not change when evidence changed")
	}
}
