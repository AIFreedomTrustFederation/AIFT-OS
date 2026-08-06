package uxi

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
	"sort"
	"time"
)

const (
	geometrySchemaV1 = "aift.federation.geometry.v1"
	mandelbrotLimit  = 64
)

var goldenAngle = math.Pi * (3 - math.Sqrt(5))

// FederationGeometry is the deterministic, evidence-derived spatial contract
// shared by every Living Federation renderer.
type FederationGeometry struct {
	Schema      string         `json:"schema"`
	GeneratedAt time.Time      `json:"generated_at"`
	Law         GeometryLaw    `json:"law"`
	Nodes       []GeometryNode `json:"nodes"`
}

// GeometryLaw identifies the mathematical rules used by compatible renderers.
type GeometryLaw struct {
	Recurrence       string  `json:"recurrence"`
	Dimensions       int     `json:"dimensions"`
	IterationLimit   int     `json:"iteration_limit"`
	Distribution     string  `json:"distribution"`
	PhyllotaxisAngle float64 `json:"phyllotaxis_angle"`
	TruthBoundary    string  `json:"truth_boundary"`
}

// Vector3 is a normalized three-dimensional coordinate.
type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// ComplexSeed is the repository's stable coordinate in the Mandelbrot plane.
type ComplexSeed struct {
	Real       float64 `json:"real"`
	Imaginary  float64 `json:"imaginary"`
	Iterations int     `json:"iterations"`
	Bounded    bool    `json:"bounded"`
	Complexity float64 `json:"complexity"`
}

// GeometryNode is one repository organism in the shared spatial world.
type GeometryNode struct {
	ID           string               `json:"id"`
	RepositoryID string               `json:"repository_id"`
	Name         string               `json:"name"`
	Role         string               `json:"role"`
	Status       string               `json:"status"`
	Seed         string               `json:"seed"`
	Mandelbrot   ComplexSeed          `json:"mandelbrot"`
	Position     Vector3              `json:"position"`
	SacredForm   string               `json:"sacred_form"`
	Symmetry     int                  `json:"symmetry"`
	Growth       int                  `json:"growth"`
	Coherence    int                  `json:"coherence"`
	Evidence     int                  `json:"evidence_count"`
	QuestIDs     []string             `json:"quest_ids"`
	Capabilities []CapabilityGeometry `json:"capabilities,omitempty"`
}

// CapabilityGeometry places a capability around its repository using
// golden-angle phyllotaxis while retaining its evidence-backed status.
type CapabilityGeometry struct {
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	Angle    float64 `json:"angle"`
	Radius   float64 `json:"radius"`
	Elevation float64 `json:"elevation"`
}

// BuildFederationGeometry produces stable identities and coordinates. Observed
// repository state controls growth; it never changes the underlying seed.
func BuildFederationGeometry(repositories []Repository) FederationGeometry {
	repos := append([]Repository(nil), repositories...)
	sort.Slice(repos, func(i, j int) bool {
		if repos[i].ID != repos[j].ID {
			return repos[i].ID < repos[j].ID
		}
		return repos[i].Name < repos[j].Name
	})
	tree := BuildFederationTree(repos)
	questIDs := make(map[string][]string, len(repos))
	for _, quest := range tree.Quests {
		questIDs[quest.RepositoryID] = append(questIDs[quest.RepositoryID], quest.ID)
	}

	world := FederationGeometry{
		Schema: geometrySchemaV1,
		GeneratedAt: time.Now().UTC(),
		Law: GeometryLaw{
			Recurrence: "z(n+1)=z(n)^2+c",
			Dimensions: 3,
			IterationLimit: mandelbrotLimit,
			Distribution: "fibonacci-sphere",
			PhyllotaxisAngle: goldenAngle,
			TruthBoundary: "geometry expresses observed evidence and never proves unobserved health",
		},
		Nodes: make([]GeometryNode, 0, len(repos)),
	}
	for i, repo := range repos {
		ready := readyCapabilities(repo.Capabilities)
		growth := repositoryGrowth(repo, ready)
		digest := sha256.Sum256([]byte(repo.ID + "|" + repo.Role))
		c := complexCoordinate(digest)
		iterations := mandelbrotIterations(c.Real, c.Imaginary, mandelbrotLimit)
		node := GeometryNode{
			ID: "geometry-" + repo.ID,
			RepositoryID: repo.ID,
			Name: repo.Name,
			Role: repo.Role,
			Status: repo.Status,
			Seed: hex.EncodeToString(digest[:]),
			Mandelbrot: ComplexSeed{
				Real: c.Real,
				Imaginary: c.Imaginary,
				Iterations: iterations,
				Bounded: iterations == mandelbrotLimit,
				Complexity: round6(float64(iterations) / mandelbrotLimit),
			},
			Position: fibonacciSphere(i, len(repos)),
			SacredForm: sacredForm(repo.Role),
			Symmetry: sacredSymmetry(repo.Role),
			Growth: growth,
			Coherence: growth,
			Evidence: len(repo.Evidence),
			QuestIDs: append([]string(nil), questIDs[repo.ID]...),
			Capabilities: capabilityGeometry(repo.Capabilities),
		}
		world.Nodes = append(world.Nodes, node)
	}
	return world
}

func complexCoordinate(digest [32]byte) ComplexSeed {
	realUnit := float64(binary.BigEndian.Uint64(digest[0:8])) / float64(^uint64(0))
	imaginaryUnit := float64(binary.BigEndian.Uint64(digest[8:16])) / float64(^uint64(0))
	return ComplexSeed{
		Real: round6(-2 + 3*realUnit),
		Imaginary: round6(-1.5 + 3*imaginaryUnit),
	}
}

func mandelbrotIterations(realPart, imaginaryPart float64, limit int) int {
	zr, zi := 0.0, 0.0
	for iteration := 0; iteration < limit; iteration++ {
		if zr*zr+zi*zi > 4 {
			return iteration
		}
		zr, zi = zr*zr-zi*zi+realPart, 2*zr*zi+imaginaryPart
	}
	return limit
}

func fibonacciSphere(index, total int) Vector3 {
	if total <= 1 {
		return Vector3{Z: 1}
	}
	y := 1 - 2*float64(index)/float64(total-1)
	radius := math.Sqrt(math.Max(0, 1-y*y))
	angle := goldenAngle * float64(index)
	return Vector3{X: round6(math.Cos(angle) * radius), Y: round6(y), Z: round6(math.Sin(angle) * radius)}
}

func capabilityGeometry(capabilities []Capability) []CapabilityGeometry {
	result := make([]CapabilityGeometry, 0, len(capabilities))
	total := math.Max(1, float64(len(capabilities)))
	for index, capability := range capabilities {
		n := float64(index + 1)
		result = append(result, CapabilityGeometry{
			Name: capability.Name,
			Status: capability.Status,
			Angle: round6(float64(index) * goldenAngle),
			Radius: round6(math.Sqrt(n / total)),
			Elevation: round6((n-0.5)/total*2 - 1),
		})
	}
	return result
}

func sacredForm(role string) string {
	switch repositoryLayer(role) {
	case 1:
		return "cube"
	case 2:
		return "dodecahedron"
	case 3:
		return "icosahedron"
	case 4:
		return "star-tetrahedron"
	case 5:
		return "torus"
	case 6:
		return "octahedron"
	case 7:
		return "metatrons-cube"
	default:
		return "sphere"
	}
}

func sacredSymmetry(role string) int {
	switch sacredForm(role) {
	case "cube":
		return 4
	case "dodecahedron":
		return 5
	case "icosahedron":
		return 20
	case "star-tetrahedron":
		return 8
	case "torus":
		return 12
	case "octahedron":
		return 8
	case "metatrons-cube":
		return 13
	default:
		return 1
	}
}

func round6(value float64) float64 {
	return math.Round(value*1e6) / 1e6
}
