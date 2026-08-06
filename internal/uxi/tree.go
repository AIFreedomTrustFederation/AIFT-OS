package uxi

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// FederationTree is a truthful, evidence-derived game view of the local federation.
type FederationTree struct {
	Schema      string        `json:"schema"`
	GeneratedAt time.Time     `json:"generated_at"`
	RootID      string        `json:"root_id"`
	Nodes       []TreeNode    `json:"nodes"`
	Edges       []TreeEdge    `json:"edges"`
	Layers      []LivingLayer `json:"layers"`
	Quests      []TreeQuest   `json:"quests"`
	Progress    TreeProgress  `json:"progress"`
}

// TreeNode represents the federation root, one living layer, or one repository leaf.
type TreeNode struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Kind            string       `json:"kind"`
	Branch          string       `json:"branch"`
	Layer           int          `json:"layer"`
	LayerName       string       `json:"layer_name"`
	Role            string       `json:"role,omitempty"`
	Status          string       `json:"status"`
	XP              int          `json:"xp"`
	Level           int          `json:"level"`
	Growth          int          `json:"growth"`
	CapabilityCount int          `json:"capability_count"`
	ReadyCount      int          `json:"ready_count"`
	EvidenceCount   int          `json:"evidence_count"`
	Languages       []string     `json:"languages,omitempty"`
	Capabilities    []Capability `json:"capabilities,omitempty"`
	Evidence        []Evidence   `json:"evidence,omitempty"`
	QuestIDs        []string     `json:"quest_ids,omitempty"`
}

// TreeEdge connects the federation root to layers and layers to repositories.
type TreeEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Kind   string `json:"kind"`
}

// LivingLayer is one permanent architectural domain in the federation spectrum.
type LivingLayer struct {
	Number int    `json:"number"`
	Roman  string `json:"roman"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Phase  string `json:"phase"`
	Quote  string `json:"quote"`
}

// TreeQuest is a truthful readiness objective derived from observed repository state.
type TreeQuest struct {
	ID           string `json:"id"`
	RepositoryID string `json:"repository_id"`
	Repository   string `json:"repository"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	RewardXP     int    `json:"reward_xp"`
	Evidence     string `json:"evidence,omitempty"`
}

// TreeProgress summarizes federation growth without claiming execution or health not observed.
type TreeProgress struct {
	XP             int `json:"xp"`
	Level          int `json:"level"`
	CurrentLevelXP int `json:"current_level_xp"`
	NextLevelXP    int `json:"next_level_xp"`
	Coherence      int `json:"coherence"`
	Repositories   int `json:"repositories"`
	Ready          int `json:"ready"`
	Detected       int `json:"detected"`
	Blocked        int `json:"blocked"`
	OpenQuests     int `json:"open_quests"`
	Completed      int `json:"completed_quests"`
}

var federationLayers = []LivingLayer{
	{Number: 1, Roman: "I", Name: "Sovereign Foundation", Color: "#ff2200", Phase: "Red · Foundation · Root", Quote: "Nothing above can stand without what is verified below."},
	{Number: 2, Roman: "II", Name: "Knowledge", Color: "#ff8800", Phase: "Orange · Creation · Memory", Quote: "Documentation is how the system remembers."},
	{Number: 3, Roman: "III", Name: "Intelligence", Color: "#ffdd00", Phase: "Yellow · Intellect · Discernment", Quote: "AI assists discovery and does not invent reality."},
	{Number: 4, Roman: "IV", Name: "Federation", Color: "#00dd55", Phase: "Green · Equilibrium · Present", Quote: "Every participant remains sovereign within shared coherence."},
	{Number: 5, Roman: "V", Name: "Economy", Color: "#00ccff", Phase: "Cyan · Harmony · Exchange", Quote: "Value follows truthful contribution and verified capability."},
	{Number: 6, Roman: "VI", Name: "Applications", Color: "#3355ff", Phase: "Blue · Truth · Service", Quote: "Software becomes civilization when it serves sovereign beings."},
	{Number: 7, Roman: "VII", Name: "Exploration", Color: "#9922ff", Phase: "Violet · Convergence · Frontier", Quote: "The frontier is reached through a verified foundation."},
}

// BuildFederationTree converts observed repositories into the dual-tree game model.
func BuildFederationTree(repositories []Repository) FederationTree {
	now := time.Now().UTC()
	tree := FederationTree{
		Schema: "aift.federation.tree.v1", GeneratedAt: now, RootID: "federation-root",
		Layers: append([]LivingLayer(nil), federationLayers...), Nodes: []TreeNode{}, Edges: []TreeEdge{}, Quests: []TreeQuest{},
	}

	tree.Nodes = append(tree.Nodes, TreeNode{
		ID: tree.RootID, Name: "AI Freedom Trust Federation", Kind: "federation", Branch: "root",
		Layer: 0, LayerName: "One Root", Status: "observed", Level: 1, Growth: 0,
	})
	for _, layer := range federationLayers {
		id := fmt.Sprintf("layer-%d", layer.Number)
		tree.Nodes = append(tree.Nodes, TreeNode{
			ID: id, Name: layer.Name, Kind: "layer", Branch: "trunk", Layer: layer.Number,
			LayerName: layer.Name, Status: "architectural", Level: layer.Number, Growth: 100,
		})
		tree.Edges = append(tree.Edges, TreeEdge{Source: tree.RootID, Target: id, Kind: "living-layer"})
	}

	repos := append([]Repository(nil), repositories...)
	sort.Slice(repos, func(i, j int) bool { return repos[i].Name < repos[j].Name })
	coherenceTotal := 0
	for _, repo := range repos {
		layer := repositoryLayer(repo.Role)
		branch := repositoryBranch(repo.Role)
		readyCount := readyCapabilities(repo.Capabilities)
		growth := repositoryGrowth(repo, readyCount)
		xp := repositoryXP(repo, readyCount, growth)
		quests := repositoryQuests(repo, readyCount)
		questIDs := make([]string, 0, len(quests))
		for _, quest := range quests {
			questIDs = append(questIDs, quest.ID)
			if quest.Status == "complete" {
				tree.Progress.Completed++
			} else {
				tree.Progress.OpenQuests++
			}
		}
		tree.Quests = append(tree.Quests, quests...)

		nodeID := "repo-" + repo.ID
		tree.Nodes = append(tree.Nodes, TreeNode{
			ID: nodeID, Name: repo.Name, Kind: "repository", Branch: branch, Layer: layer,
			LayerName: layerName(layer), Role: repo.Role, Status: repo.Status, XP: xp,
			Level: 1 + xp/250, Growth: growth, CapabilityCount: len(repo.Capabilities), ReadyCount: readyCount,
			EvidenceCount: len(repo.Evidence), Languages: append([]string(nil), repo.Languages...),
			Capabilities: append([]Capability(nil), repo.Capabilities...), Evidence: append([]Evidence(nil), repo.Evidence...), QuestIDs: questIDs,
		})
		tree.Edges = append(tree.Edges, TreeEdge{Source: fmt.Sprintf("layer-%d", layer), Target: nodeID, Kind: branch})

		tree.Progress.Repositories++
		tree.Progress.XP += xp
		coherenceTotal += growth
		switch strings.ToLower(repo.Status) {
		case "ready":
			tree.Progress.Ready++
		case "blocked":
			tree.Progress.Blocked++
		default:
			tree.Progress.Detected++
		}
	}

	tree.Progress.Level = 1 + tree.Progress.XP/1000
	tree.Progress.CurrentLevelXP = tree.Progress.XP % 1000
	tree.Progress.NextLevelXP = 1000
	if tree.Progress.Repositories > 0 {
		tree.Progress.Coherence = coherenceTotal / tree.Progress.Repositories
	}
	tree.Nodes[0].XP = tree.Progress.XP
	tree.Nodes[0].Level = tree.Progress.Level
	tree.Nodes[0].Growth = tree.Progress.Coherence
	return tree
}

func repositoryLayer(role string) int {
	switch role {
	case "runtime-prototype", "compatibility-runtime", "infrastructure-and-nodes":
		return 1
	case "doctrine-and-research", "knowledge-application", "knowledge-product-specification":
		return 2
	case "model-registry", "software-mission-engine", "federation-genome":
		return 3
	case "federation-kernel":
		return 4
	case "stewardship-application":
		return 5
	case "video-application", "federated-application":
		return 6
	case "exploration-platform", "simulation-environment":
		return 7
	default:
		return 6
	}
}

func repositoryBranch(role string) string {
	switch role {
	case "federation-kernel":
		return "root"
	case "doctrine-and-research", "knowledge-application", "knowledge-product-specification", "model-registry", "federation-genome":
		return "knowledge"
	default:
		return "life"
	}
}

func layerName(number int) string {
	for _, layer := range federationLayers {
		if layer.Number == number {
			return layer.Name
		}
	}
	return "Applications"
}

func readyCapabilities(capabilities []Capability) int {
	ready := 0
	for _, capability := range capabilities {
		switch strings.ToLower(strings.TrimSpace(capability.Status)) {
		case "ready", "active", "v1":
			ready++
		}
	}
	return ready
}

func repositoryGrowth(repo Repository, readyCount int) int {
	growth := 20 // Git repository evidence.
	if len(repo.Languages) > 0 {
		growth += 15
	}
	if len(repo.Capabilities) > 0 {
		growth += 20
	}
	growth += minInt(readyCount*10, 30)
	switch strings.ToLower(repo.Status) {
	case "ready":
		growth += 15
	case "blocked":
		growth = minInt(growth, 15)
	}
	return minInt(growth, 100)
}

func repositoryXP(repo Repository, readyCount, growth int) int {
	xp := 25 + len(repo.Evidence)*10 + len(repo.Languages)*15 + readyCount*75 + growth
	switch strings.ToLower(repo.Status) {
	case "ready":
		xp += 100
	case "detected":
		xp += 25
	case "blocked":
		xp += 5
	}
	return xp
}

func repositoryQuests(repo Repository, readyCount int) []TreeQuest {
	base := "quest-" + repo.ID + "-"
	quests := []TreeQuest{
		{ID: base + "awaken", RepositoryID: repo.ID, Repository: repo.Name, Title: "Awaken the repository", Description: "Be discovered as a local Git repository.", Status: "complete", RewardXP: 25, Evidence: filepathEvidence(repo)},
	}
	manifest := TreeQuest{ID: base + "capabilities", RepositoryID: repo.ID, Repository: repo.Name, Title: "Name its living capabilities", Description: "Publish .aift/capabilities.json with evidence-backed capability records.", Status: "open", RewardXP: 100}
	if len(repo.Capabilities) > 0 {
		manifest.Status = "complete"
		manifest.Evidence = fmt.Sprintf("%d declared capabilities observed", len(repo.Capabilities))
	}
	quests = append(quests, manifest)

	proof := TreeQuest{ID: base + "prove-ready", RepositoryID: repo.ID, Repository: repo.Name, Title: "Prove one ready capability", Description: "Verify at least one capability and mark it ready only when its evidence succeeds.", Status: "open", RewardXP: 150}
	if readyCount > 0 && strings.EqualFold(repo.Status, "ready") {
		proof.Status = "complete"
		proof.Evidence = fmt.Sprintf("%d ready capabilities observed", readyCount)
	} else if strings.EqualFold(repo.Status, "blocked") {
		proof.Title = "Heal the blocked branch"
		proof.Description = "Repair failed capability evidence before claiming readiness."
	}
	return append(quests, proof)
}

func filepathEvidence(repo Repository) string {
	for _, evidence := range repo.Evidence {
		if evidence.Kind == "filesystem" {
			return evidence.Source
		}
	}
	return repo.Path
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
