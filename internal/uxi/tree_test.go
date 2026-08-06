package uxi

import "testing"

func TestBuildFederationTreeDerivesProgressFromEvidence(t *testing.T) {
	repositories := []Repository{
		{
			ID: "aift-runtime", Name: "AIFT-Runtime", Role: "runtime-prototype", Status: "ready", Git: true,
			Languages:    []string{"Shell"},
			Capabilities: []Capability{{Name: "runtime.health", Status: "ready"}, {Name: "governed.execution", Status: "planned"}},
			Evidence:     []Evidence{{Kind: "filesystem", Source: "/tmp/AIFT-Runtime/.git", Status: "observed"}, {Kind: "capability_manifest", Source: "/tmp/AIFT-Runtime/.aift/capabilities.json", Status: "observed"}},
		},
		{
			ID: "booksmith-ai", Name: "booksmith-ai", Role: "knowledge-application", Status: "detected", Git: true,
			Evidence: []Evidence{{Kind: "filesystem", Source: "/tmp/booksmith-ai/.git", Status: "observed"}},
		},
	}

	tree := BuildFederationTree(repositories)
	if tree.Schema != "aift.federation.tree.v1" {
		t.Fatalf("schema=%q", tree.Schema)
	}
	if len(tree.Nodes) != 10 { // root + seven layers + two repositories
		t.Fatalf("nodes=%d", len(tree.Nodes))
	}
	if len(tree.Edges) != 9 {
		t.Fatalf("edges=%d", len(tree.Edges))
	}
	if tree.Progress.Repositories != 2 || tree.Progress.Ready != 1 || tree.Progress.Detected != 1 {
		t.Fatalf("progress=%#v", tree.Progress)
	}
	if tree.Progress.Coherence <= 0 || tree.Progress.Coherence > 100 {
		t.Fatalf("coherence=%d", tree.Progress.Coherence)
	}

	var runtime, booksmith TreeNode
	for _, node := range tree.Nodes {
		switch node.Name {
		case "AIFT-Runtime":
			runtime = node
		case "booksmith-ai":
			booksmith = node
		}
	}
	if runtime.Branch != "life" || runtime.Layer != 1 || runtime.ReadyCount != 1 {
		t.Fatalf("runtime=%#v", runtime)
	}
	if booksmith.Branch != "knowledge" || booksmith.Layer != 2 {
		t.Fatalf("booksmith=%#v", booksmith)
	}
	if runtime.XP <= booksmith.XP || runtime.Growth <= booksmith.Growth {
		t.Fatalf("runtime=%#v booksmith=%#v", runtime, booksmith)
	}
}

func TestBuildFederationTreeCreatesOpenReadinessQuests(t *testing.T) {
	tree := BuildFederationTree([]Repository{{
		ID: "aift-forge", Name: "AIFT-Forge", Role: "software-mission-engine", Status: "detected", Git: true,
		Evidence: []Evidence{{Kind: "filesystem", Source: "/tmp/AIFT-Forge/.git", Status: "observed"}},
	}})
	if tree.Progress.OpenQuests != 2 || tree.Progress.Completed != 1 {
		t.Fatalf("progress=%#v quests=%#v", tree.Progress, tree.Quests)
	}
	if tree.Quests[1].Title != "Name its living capabilities" || tree.Quests[1].Status != "open" {
		t.Fatalf("quest=%#v", tree.Quests[1])
	}
}

func TestBuildFederationTreeIsDeterministicByRepositoryName(t *testing.T) {
	repositories := []Repository{
		{ID: "z", Name: "Zeta", Role: "federated-application", Status: "detected"},
		{ID: "a", Name: "Alpha", Role: "federated-application", Status: "detected"},
	}
	tree := BuildFederationTree(repositories)
	if tree.Nodes[8].Name != "Alpha" || tree.Nodes[9].Name != "Zeta" {
		t.Fatalf("order=%q,%q", tree.Nodes[8].Name, tree.Nodes[9].Name)
	}
}
