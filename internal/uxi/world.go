package uxi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FederationWorld is a privacy-safe, evidence-derived geographic view of the federation.
type FederationWorld struct {
	Schema      string          `json:"schema"`
	GeneratedAt time.Time       `json:"generated_at"`
	Projection  string          `json:"projection"`
	Privacy     WorldPrivacy    `json:"privacy"`
	Nodes       []WorldNode     `json:"nodes"`
	Unmapped    []WorldUnmapped `json:"unmapped"`
	Quests      []WorldQuest    `json:"quests"`
	Progress    WorldProgress   `json:"progress"`
}

// WorldPrivacy documents the map's location-handling contract.
type WorldPrivacy struct {
	DeviceLocationStorage string `json:"device_location_storage"`
	RepositoryLocations    string `json:"repository_locations"`
	ExternalRequests       bool   `json:"external_requests"`
	DefaultPrecision       string `json:"default_precision"`
}

// WorldNode is one repository with a valid, shareable location declaration.
type WorldNode struct {
	ID           string   `json:"id"`
	RepositoryID string   `json:"repository_id"`
	Repository   string   `json:"repository"`
	Role         string   `json:"role"`
	Status       string   `json:"status"`
	Label        string   `json:"label"`
	Latitude     float64  `json:"latitude"`
	Longitude    float64  `json:"longitude"`
	Precision    string   `json:"precision"`
	Visibility   string   `json:"visibility"`
	Source       string   `json:"source"`
	Evidence     string   `json:"evidence"`
	XP           int      `json:"xp"`
	QuestIDs     []string `json:"quest_ids"`
}

// WorldUnmapped identifies repositories that cannot be shown geographically.
type WorldUnmapped struct {
	RepositoryID string `json:"repository_id"`
	Repository   string `json:"repository"`
	Role         string `json:"role"`
	Status       string `json:"status"`
	Reason       string `json:"reason"`
	State        string `json:"state"`
}

// WorldQuest is a location-readiness objective derived from repository evidence.
type WorldQuest struct {
	ID           string `json:"id"`
	RepositoryID string `json:"repository_id"`
	Repository   string `json:"repository"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	RewardXP     int    `json:"reward_xp"`
	Evidence     string `json:"evidence,omitempty"`
}

// WorldProgress summarizes declared geographic coverage without rewarding oversharing.
type WorldProgress struct {
	Repositories    int `json:"repositories"`
	Mapped          int `json:"mapped"`
	Unmapped        int `json:"unmapped"`
	Hidden          int `json:"hidden"`
	Invalid         int `json:"invalid"`
	XP              int `json:"xp"`
	Level           int `json:"level"`
	OpenQuests      int `json:"open_quests"`
	CompletedQuests int `json:"completed_quests"`
}

type locationManifest struct {
	Schema     string   `json:"schema"`
	Label      string   `json:"label"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
	Precision  string   `json:"precision"`
	Visibility string   `json:"visibility"`
	Source     string   `json:"source"`
	UpdatedAt  string   `json:"updated_at"`
}

// BuildFederationWorld reads optional .aift/location.json declarations from observed repositories.
func BuildFederationWorld(repositories []Repository) FederationWorld {
	world := FederationWorld{
		Schema: "aift.federation.world.v1", GeneratedAt: time.Now().UTC(), Projection: "equirectangular",
		Privacy: WorldPrivacy{
			DeviceLocationStorage: "browser-memory-only",
			RepositoryLocations:    ".aift/location.json with explicit federation or public visibility",
			ExternalRequests:       false,
			DefaultPrecision:       "city",
		},
		Nodes: []WorldNode{}, Unmapped: []WorldUnmapped{}, Quests: []WorldQuest{},
	}

	repos := append([]Repository(nil), repositories...)
	sort.Slice(repos, func(i, j int) bool {
		if repos[i].Name != repos[j].Name {
			return repos[i].Name < repos[j].Name
		}
		return repos[i].ID < repos[j].ID
	})

	for _, repo := range repos {
		world.Progress.Repositories++
		manifestPath := filepath.Join(repo.Path, ".aift", "location.json")
		manifest, err := readLocationManifest(manifestPath)
		quests := locationQuests(repo, manifest, err)
		questIDs := make([]string, 0, len(quests))
		for _, quest := range quests {
			questIDs = append(questIDs, quest.ID)
			if quest.Status == "complete" {
				world.Progress.CompletedQuests++
			} else {
				world.Progress.OpenQuests++
			}
		}
		world.Quests = append(world.Quests, quests...)

		switch {
		case errors.Is(err, os.ErrNotExist):
			world.Progress.Unmapped++
			world.Unmapped = append(world.Unmapped, worldUnmapped(repo, "missing .aift/location.json", "missing"))
		case err != nil:
			world.Progress.Invalid++
			world.Unmapped = append(world.Unmapped, worldUnmapped(repo, err.Error(), "invalid"))
		case manifest.Visibility == "private":
			world.Progress.Hidden++
			world.Unmapped = append(world.Unmapped, worldUnmapped(repo, "location declaration is private", "hidden"))
		default:
			world.Progress.Mapped++
			world.Progress.XP += 150
			world.Nodes = append(world.Nodes, WorldNode{
				ID: "world-" + repo.ID, RepositoryID: repo.ID, Repository: repo.Name, Role: repo.Role, Status: repo.Status,
				Label: manifest.Label, Latitude: *manifest.Latitude, Longitude: *manifest.Longitude,
				Precision: manifest.Precision, Visibility: manifest.Visibility, Source: manifest.Source,
				Evidence: ".aift/location.json", XP: 150, QuestIDs: questIDs,
			})
		}
	}

	world.Progress.Level = 1 + world.Progress.XP/1000
	return world
}

func readLocationManifest(path string) (locationManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return locationManifest{}, err
	}
	var manifest locationManifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return locationManifest{}, fmt.Errorf("invalid location JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return locationManifest{}, errors.New("invalid location JSON: expected exactly one object")
	}
	if manifest.Schema != "aift.location.v1" {
		return locationManifest{}, fmt.Errorf("unsupported location schema %q", manifest.Schema)
	}
	manifest.Label = strings.TrimSpace(manifest.Label)
	if manifest.Label == "" {
		return locationManifest{}, errors.New("location label is required")
	}
	if len(manifest.Label) > 120 {
		return locationManifest{}, errors.New("location label exceeds 120 characters")
	}
	if manifest.Latitude == nil {
		return locationManifest{}, errors.New("latitude is required")
	}
	if manifest.Longitude == nil {
		return locationManifest{}, errors.New("longitude is required")
	}
	if *manifest.Latitude < -90 || *manifest.Latitude > 90 || math.IsNaN(*manifest.Latitude) || math.IsInf(*manifest.Latitude, 0) {
		return locationManifest{}, errors.New("latitude must be between -90 and 90")
	}
	if *manifest.Longitude < -180 || *manifest.Longitude > 180 || math.IsNaN(*manifest.Longitude) || math.IsInf(*manifest.Longitude, 0) {
		return locationManifest{}, errors.New("longitude must be between -180 and 180")
	}
	manifest.Precision = strings.ToLower(strings.TrimSpace(manifest.Precision))
	if manifest.Precision == "" {
		manifest.Precision = "city"
	}
	decimals, ok := precisionDecimals(manifest.Precision)
	if !ok {
		return locationManifest{}, fmt.Errorf("unsupported location precision %q", manifest.Precision)
	}
	manifest.Visibility = strings.ToLower(strings.TrimSpace(manifest.Visibility))
	if manifest.Visibility == "" {
		manifest.Visibility = "private"
	}
	if manifest.Visibility != "private" && manifest.Visibility != "federation" && manifest.Visibility != "public" {
		return locationManifest{}, fmt.Errorf("unsupported location visibility %q", manifest.Visibility)
	}
	manifest.Source = strings.TrimSpace(manifest.Source)
	if manifest.Source == "" {
		manifest.Source = "operator-declared"
	}
	latitude := roundCoordinate(*manifest.Latitude, decimals)
	longitude := roundCoordinate(*manifest.Longitude, decimals)
	manifest.Latitude = &latitude
	manifest.Longitude = &longitude
	return manifest, nil
}

func precisionDecimals(precision string) (int, bool) {
	switch precision {
	case "country":
		return 0, true
	case "region":
		return 1, true
	case "city":
		return 2, true
	case "exact":
		return 5, true
	default:
		return 0, false
	}
}

func roundCoordinate(value float64, decimals int) float64 {
	factor := math.Pow10(decimals)
	return math.Round(value*factor) / factor
}

func locationQuests(repo Repository, manifest locationManifest, err error) []WorldQuest {
	base := "world-quest-" + repo.ID + "-"
	declare := WorldQuest{
		ID: base + "declare", RepositoryID: repo.ID, Repository: repo.Name,
		Title: "Anchor the repository on Earth", Description: "Add a valid .aift/location.json declaration.", Status: "open", RewardXP: 100,
	}
	share := WorldQuest{
		ID: base + "share", RepositoryID: repo.ID, Repository: repo.Name,
		Title: "Choose federation visibility", Description: "Explicitly choose federation or public visibility to place the repository on the shared map.", Status: "open", RewardXP: 50,
	}
	if err == nil {
		declare.Status = "complete"
		declare.Evidence = ".aift/location.json"
		if manifest.Visibility == "federation" || manifest.Visibility == "public" {
			share.Status = "complete"
			share.Evidence = manifest.Visibility + " visibility at " + manifest.Precision + " precision"
		}
	}
	return []WorldQuest{declare, share}
}

func worldUnmapped(repo Repository, reason, state string) WorldUnmapped {
	return WorldUnmapped{
		RepositoryID: repo.ID, Repository: repo.Name, Role: repo.Role, Status: repo.Status,
		Reason: reason, State: state,
	}
}
