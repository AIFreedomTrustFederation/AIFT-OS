package uxi

import (
	"reflect"
	"testing"
)

func TestFederationGeometryIsDeterministic(t *testing.T) {
	repositories := []Repository{
		{ID: "aift-os", Name: "AIFT-OS", Role: "federation-kernel", Status: "ready", Git: true, Languages: []string{"Go"}, Capabilities: []Capability{{Name: "uxi", Status: "ready"}}, Evidence: []Evidence{{Kind: "test"}}},
		{ID: "mobox", Name: "mobox", Role: "compatibility-runtime", Status: "detected", Git: true},
	}
	first := BuildFederationGeometry(repositories)
	second := BuildFederationGeometry(repositories)
	first.GeneratedAt = second.GeneratedAt
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("geometry is not deterministic:\nfirst=%#v\nsecond=%#v", first, second)
	}
	if first.Schema != geometrySchemaV1 || first.Law.Dimensions != 3 || len(first.Nodes) != 2 {
		t.Fatalf("geometry=%#v", first)
	}
	for _, node := range first.Nodes {
		if node.Seed == "" || node.Mandelbrot.Iterations < 0 || node.Mandelbrot.Iterations > mandelbrotLimit {
			t.Fatalf("node=%#v", node)
		}
		if node.Position.X < -1 || node.Position.X > 1 || node.Position.Y < -1 || node.Position.Y > 1 || node.Position.Z < -1 || node.Position.Z > 1 {
			t.Fatalf("position=%#v", node.Position)
		}
	}
}

func TestGeometryBindsQuestsAndCapabilities(t *testing.T) {
	repository := Repository{
		ID: "aift-os", Name: "AIFT-OS", Role: "federation-kernel", Status: "ready", Git: true,
		Capabilities: []Capability{{Name: "inspect", Status: "ready"}, {Name: "repair", Status: "planned"}},
		Evidence:     []Evidence{{Kind: "filesystem"}},
	}
	world := BuildFederationGeometry([]Repository{repository})
	node := world.Nodes[0]
	if node.SacredForm != "star-tetrahedron" || node.Symmetry != 8 {
		t.Fatalf("form=%s symmetry=%d", node.SacredForm, node.Symmetry)
	}
	if len(node.QuestIDs) != 3 || len(node.Capabilities) != 2 {
		t.Fatalf("node=%#v", node)
	}
	if node.Capabilities[0].Angle == node.Capabilities[1].Angle {
		t.Fatalf("capabilities do not use phyllotaxis: %#v", node.Capabilities)
	}
}

func TestMandelbrotIterationBounds(t *testing.T) {
	if got := mandelbrotIterations(0, 0, mandelbrotLimit); got != mandelbrotLimit {
		t.Fatalf("origin iterations=%d", got)
	}
	if got := mandelbrotIterations(2, 2, mandelbrotLimit); got >= mandelbrotLimit {
		t.Fatalf("escaping point iterations=%d", got)
	}
}
