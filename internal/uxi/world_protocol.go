package uxi

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"sort"
	"strconv"
)

const worldProtocolSchemaV1 = "aift.world.v1"

const logicalScale = 1_000_000_000

// WorldSnapshot is the engine-independent, evidence-derived spatial state.
type WorldSnapshot struct {
	Schema           string            `json:"schema"`
	WorldID          string            `json:"world_id"`
	Revision         uint64            `json:"revision"`
	CoordinateSystem CoordinateSystem  `json:"coordinate_system"`
	Entities         []WorldEntity     `json:"entities"`
	Relations        []WorldRelation   `json:"relations"`
	Chunks           []WorldChunk      `json:"chunks"`
	VisualGrammar    WorldVisualGrammar `json:"visual_grammar"`
	Governance       WorldGovernance   `json:"governance"`
}

// CoordinateSystem keeps canonical coordinates outside renderer precision limits.
type CoordinateSystem struct {
	Name       string          `json:"name"`
	Dimensions int             `json:"dimensions"`
	Unit       string          `json:"unit"`
	Origin     LogicalVector64 `json:"origin"`
}

// LogicalVector64 encodes signed 64-bit coordinates as decimal strings so JSON
// clients do not lose precision.
type LogicalVector64 struct {
	X string `json:"x"`
	Y string `json:"y"`
	Z string `json:"z"`
}

// WorldEntity is one streamable object in the Living Federation.
type WorldEntity struct {
	ID           string         `json:"id"`
	Kind         string         `json:"kind"`
	ParentID     *string        `json:"parent_id,omitempty"`
	Revision     uint64         `json:"revision"`
	Position     LogicalVector64 `json:"position"`
	Status       string         `json:"status"`
	Name         string         `json:"name"`
	Role         string         `json:"role"`
	Properties   map[string]any `json:"properties,omitempty"`
	EvidenceRefs []string       `json:"evidence_refs"`
	Visual       WorldVisual    `json:"visual"`
}

// WorldVisual separates symbolic presentation hints from factual state.
type WorldVisual struct {
	Form          string  `json:"form"`
	IdentityPhase float64 `json:"identity_phase"`
	Coherence     float64 `json:"coherence"`
	PaletteToken  string  `json:"palette_token,omitempty"`
	LODPriority   int     `json:"lod_priority,omitempty"`
}

// WorldRelation connects two stable entities.
type WorldRelation struct {
	ID       string  `json:"id"`
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Kind     string  `json:"kind"`
	Directed bool    `json:"directed"`
	Weight   float64 `json:"weight,omitempty"`
}

// WorldChunk is an independently streamable spatial partition.
type WorldChunk struct {
	ID          string          `json:"id"`
	Level       int             `json:"level"`
	Center      LogicalVector64 `json:"center"`
	Radius      string          `json:"radius"`
	EntityIDs   []string        `json:"entity_ids"`
	ContentHash string          `json:"content_hash"`
}

// WorldVisualGrammar is a renderer-neutral semantic palette and LOD contract.
type WorldVisualGrammar struct {
	Palette     map[string]string `json:"palette"`
	SemanticLOD []WorldLOD        `json:"semantic_lod"`
}

// WorldLOD declares which entity kinds are meaningful at an altitude.
type WorldLOD struct {
	ID           string   `json:"id"`
	MinAltitude  string   `json:"min_altitude"`
	VisibleKinds []string `json:"visible_kinds"`
}

// WorldGovernance makes the read/propose boundary explicit to every client.
type WorldGovernance struct {
	TruthSource          string `json:"truth_source"`
	MutationMode         string `json:"mutation_mode"`
	HumanConsentRequired bool   `json:"human_consent_required"`
}

// BuildWorldSnapshot projects discovered repositories into a deterministic,
// large-coordinate world suitable for WebGL, native engines, and accessible clients.
func BuildWorldSnapshot(repositories []Repository) WorldSnapshot {
	geometry := BuildFederationGeometry(repositories)
	entities := make([]WorldEntity, 0, len(geometry.Nodes)*2)
	relations := make([]WorldRelation, 0)
	for _, node := range geometry.Nodes {
		evidenceRefs := repositoryEvidenceRefs(repositories, node.RepositoryID)
		repositoryID := "repository:" + node.RepositoryID
		entities = append(entities, WorldEntity{
			ID: repositoryID, Kind: "repository", Revision: 1,
			Position: logicalVector(node.Position, logicalScale),
			Status: node.Status, Name: node.Name, Role: node.Role,
			Properties: map[string]any{
				"mandelbrot_seed": node.Seed,
				"mandelbrot_bounded": node.Mandelbrot.Bounded,
				"growth": node.Growth,
				"quest_ids": node.QuestIDs,
			},
			EvidenceRefs: evidenceRefs,
			Visual: WorldVisual{
				Form: node.SacredForm, IdentityPhase: node.Torus.IdentityPhase,
				Coherence: round6(float64(node.Coherence) / 100),
				PaletteToken: repositoryPalette(node.Status), LODPriority: 100,
			},
		})
		for index, capability := range node.Capabilities {
			capabilityID := repositoryID + "/capability:" + stableWorldID(capability.Name)
			parent := repositoryID
			position := capabilityWorldPosition(node.Position, capability)
			entities = append(entities, WorldEntity{
				ID: capabilityID, Kind: "capability", ParentID: &parent, Revision: 1,
				Position: position, Status: capability.Status, Name: capability.Name,
				Role: "repository-capability", EvidenceRefs: []string{},
				Visual: WorldVisual{
					Form: "sphere", IdentityPhase: round6(capability.Angle),
					Coherence: capabilityCoherence(capability.Status),
					PaletteToken: repositoryPalette(capability.Status),
					LODPriority: 50 - index,
				},
			})
			relations = append(relations, WorldRelation{
				ID: repositoryID + "->" + capabilityID,
				Source: repositoryID, Target: capabilityID,
				Kind: "contains", Directed: true, Weight: 1,
			})
		}
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].ID < entities[j].ID })
	sort.Slice(relations, func(i, j int) bool { return relations[i].ID < relations[j].ID })
	entityIDs := make([]string, len(entities))
	for index := range entities {
		entityIDs[index] = entities[index].ID
	}
	chunkHash := worldContentHash(struct {
		Entities  []WorldEntity
		Relations []WorldRelation
	}{entities, relations})
	snapshot := WorldSnapshot{
		Schema: worldProtocolSchemaV1,
		WorldID: "aift-federation",
		CoordinateSystem: CoordinateSystem{
			Name: "aift.logical.cartesian.v1", Dimensions: 3,
			Unit: "logical-nanounit", Origin: zeroLogicalVector(),
		},
		Entities: entities,
		Relations: relations,
		Chunks: []WorldChunk{{
			ID: "federation-root", Level: 0, Center: zeroLogicalVector(),
			Radius: strconv.FormatInt(logicalScale, 10),
			EntityIDs: entityIDs, ContentHash: chunkHash,
		}},
		VisualGrammar: defaultWorldVisualGrammar(),
		Governance: WorldGovernance{
			TruthSource: "recorded-evidence", MutationMode: "proposal-only",
			HumanConsentRequired: true,
		},
	}
	snapshot.Revision = worldRevision(snapshot)
	return snapshot
}

func logicalVector(value Vector3, scale int64) LogicalVector64 {
	return LogicalVector64{
		X: strconv.FormatInt(int64(math.Round(value.X*float64(scale))), 10),
		Y: strconv.FormatInt(int64(math.Round(value.Y*float64(scale))), 10),
		Z: strconv.FormatInt(int64(math.Round(value.Z*float64(scale))), 10),
	}
}

func zeroLogicalVector() LogicalVector64 {
	return LogicalVector64{X: "0", Y: "0", Z: "0"}
}

func capabilityWorldPosition(parent Vector3, capability CapabilityGeometry) LogicalVector64 {
	offset := .08 * capability.Radius
	return logicalVector(Vector3{
		X: parent.X + math.Cos(capability.Angle)*offset,
		Y: parent.Y + capability.Elevation*.04,
		Z: parent.Z + math.Sin(capability.Angle)*offset,
	}, logicalScale)
}

func capabilityCoherence(status string) float64 {
	if status == "ready" || status == "active" {
		return 1
	}
	if status == "blocked" {
		return .15
	}
	return .55
}

func repositoryPalette(status string) string {
	switch status {
	case "ready", "active":
		return "living-green"
	case "blocked":
		return "boundary-pink"
	case "detected", "planned":
		return "emergence-orange"
	default:
		return "connection-cyan"
	}
}

func repositoryEvidenceRefs(repositories []Repository, repositoryID string) []string {
	for _, repository := range repositories {
		if repository.ID != repositoryID {
			continue
		}
		refs := make([]string, 0, len(repository.Evidence))
		for _, evidence := range repository.Evidence {
			if evidence.ID != "" {
				refs = append(refs, evidence.ID)
			}
		}
		sort.Strings(refs)
		return refs
	}
	return []string{}
}

func stableWorldID(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:8])
}

func worldContentHash(value any) string {
	payload, _ := json.Marshal(value)
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func worldRevision(snapshot WorldSnapshot) uint64 {
	snapshot.Revision = 0
	payload, _ := json.Marshal(snapshot)
	digest := sha256.Sum256(payload)
	return binary.BigEndian.Uint64(digest[:8]) & uint64(0x7fffffffffffffff)
}

func defaultWorldVisualGrammar() WorldVisualGrammar {
	return WorldVisualGrammar{
		Palette: map[string]string{
			"ink": "#02030A", "connection-cyan": "#00E5E5",
			"living-green": "#00E676", "knowledge-blue": "#147DF5",
			"transformation-coral": "#FF4F46", "boundary-pink": "#FF256E",
			"emergence-orange": "#FF7A35", "solar-yellow": "#FFD21F",
			"crown-violet": "#9B5CFF", "coherence-white": "#F4FBFF",
		},
		SemanticLOD: []WorldLOD{
			{ID: "federation", MinAltitude: "1000000000", VisibleKinds: []string{"federation", "layer", "repository"}},
			{ID: "repository", MinAltitude: "10000000", VisibleKinds: []string{"repository", "mission", "quest"}},
			{ID: "capability", MinAltitude: "100000", VisibleKinds: []string{"repository", "capability", "evidence"}},
			{ID: "evidence", MinAltitude: "0", VisibleKinds: []string{"capability", "evidence", "event"}},
		},
	}
}
