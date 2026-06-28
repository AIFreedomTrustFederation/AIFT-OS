package discoveryengine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/fsutil"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/jsonfile"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/sliceutil"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Evidence struct {
	Kind        string `json:"kind"`
	Path        string `json:"path"`
	Description string `json:"description"`
	ObservedAt  string `json:"observedAt"`
}

type Runtime struct {
	Name     string     `json:"name"`
	Kind     string     `json:"kind"`
	Version  string     `json:"version,omitempty"`
	Evidence []Evidence `json:"evidence"`
}

type DiscoveryObject struct {
	ID           string            `json:"id"`
	Kind         string            `json:"kind"`
	Name         string            `json:"name"`
	Status       string            `json:"status"`
	Location     string            `json:"location"`
	Description  string            `json:"description"`
	Evidence     []Evidence        `json:"evidence"`
	Runtimes     []Runtime         `json:"runtimes"`
	Manifests    []string          `json:"manifests"`
	Docs         []string          `json:"docs"`
	Schemas      []string          `json:"schemas"`
	Workflows    []string          `json:"workflows"`
	Commands     map[string]string `json:"commands"`
	Capabilities []string          `json:"capabilities"`
	Services     []string          `json:"services"`
	HealthChecks []string          `json:"healthChecks"`
	GeneratedAt  string            `json:"generatedAt"`
}

type Snapshot struct {
	SchemaVersion string            `json:"schemaVersion"`
	GeneratedAt   string            `json:"generatedAt"`
	Source        string            `json:"source"`
	Objects       []DiscoveryObject `json:"objects"`
}

func Scan(cfg config.Config) error {
	snap, err := Build(cfg)
	if err != nil {
		return err
	}
	if err := Write(cfg, snap); err != nil {
		return err
	}
	if err := WriteReport(cfg, snap); err != nil {
		return err
	}
	return events.Emit(cfg, "discovery.scan.completed", "discoveryengine", "discovery scan completed", map[string]string{
		"objects": fmt.Sprint(len(snap.Objects)),
	})
}

func Build(cfg config.Config) (Snapshot, error) {
	now := time.Now().Format(time.RFC3339)

	snap := Snapshot{
		SchemaVersion: "aift.discovery.v1",
		GeneratedAt:   now,
		Source:        "filesystem.git.manifests",
		Objects:       []DiscoveryObject{},
	}

	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return snap, err
	}

	for _, repo := range repos {
		snap.Objects = append(snap.Objects, DiscoverRepository(now, repo.Name, repo.Path))
	}

	sort.Slice(snap.Objects, func(i, j int) bool {
		return snap.Objects[i].ID < snap.Objects[j].ID
	})

	return snap, nil
}

func DiscoverRepository(now string, name string, path string) DiscoveryObject {
	obj := DiscoveryObject{
		ID:           "repository." + safeID(name),
		Kind:         "repository",
		Name:         name,
		Status:       "detected",
		Location:     path,
		Description:  "Repository discovered from filesystem and Git evidence.",
		Evidence:     []Evidence{},
		Runtimes:     []Runtime{},
		Manifests:    []string{},
		Docs:         []string{},
		Schemas:      []string{},
		Workflows:    []string{},
		Commands:     map[string]string{},
		Capabilities: []string{},
		Services:     []string{},
		HealthChecks: []string{},
		GeneratedAt:  now,
	}

	addEvidence(&obj, now, "git", filepath.Join(path, ".git"), "Git repository exists.")

	discoverDocs(&obj, now, path)
	discoverSchemas(&obj, now, path)
	discoverWorkflows(&obj, now, path)
	discoverManifests(&obj, now, path)
	discoverRuntimes(&obj, now, path)
	discoverAIFTContracts(&obj, now, path)

	if len(obj.Commands) > 0 || len(obj.Capabilities) > 0 || len(obj.Services) > 0 {
		obj.Status = "ready"
	}

	obj.Docs = sliceutil.Unique(obj.Docs)
	obj.Schemas = sliceutil.Unique(obj.Schemas)
	obj.Workflows = sliceutil.Unique(obj.Workflows)
	obj.Manifests = sliceutil.Unique(obj.Manifests)
	obj.Capabilities = sliceutil.Unique(obj.Capabilities)
	obj.Services = sliceutil.Unique(obj.Services)
	obj.HealthChecks = sliceutil.Unique(obj.HealthChecks)

	return obj
}

func discoverDocs(obj *DiscoveryObject, now string, root string) {
	candidates := []string{"README.md", "AGENTS.md", "docs", "manual", "book", "site"}
	for _, rel := range candidates {
		path := filepath.Join(root, rel)
		if fsutil.Exists(path) {
			obj.Docs = append(obj.Docs, rel)
			addEvidence(obj, now, "documentation", path, "Documentation path exists.")
		}
	}
}

func discoverSchemas(obj *DiscoveryObject, now string, root string) {
	candidates := []string{"schemas", "schema", ".aift/schemas"}
	for _, rel := range candidates {
		path := filepath.Join(root, rel)
		if fsutil.Exists(path) {
			obj.Schemas = append(obj.Schemas, rel)
			addEvidence(obj, now, "schema", path, "Schema path exists.")
		}
	}
}

func discoverWorkflows(obj *DiscoveryObject, now string, root string) {
	candidates := []string{".github/workflows", "workflows", ".aift/workflows"}
	for _, rel := range candidates {
		path := filepath.Join(root, rel)
		if fsutil.Exists(path) {
			obj.Workflows = append(obj.Workflows, rel)
			addEvidence(obj, now, "workflow", path, "Workflow path exists.")
		}
	}
}

func discoverManifests(obj *DiscoveryObject, now string, root string) {
	manifestFiles := []string{
		"package.json",
		"go.mod",
		"Cargo.toml",
		"pyproject.toml",
		"requirements.txt",
		"deno.json",
		"bun.lockb",
		"pnpm-lock.yaml",
		"yarn.lock",
		"package-lock.json",
		"Dockerfile",
		"docker-compose.yml",
		"compose.yml",
		".aift/module.json",
		".aift/capabilities.json",
		".aift/services.json",
		".aift/events.json",
		".aift/manual.json",
	}
	for _, rel := range manifestFiles {
		path := filepath.Join(root, rel)
		if fsutil.Exists(path) {
			obj.Manifests = append(obj.Manifests, rel)
			addEvidence(obj, now, "manifest", path, "Manifest file exists.")
		}
	}
}

func discoverRuntimes(obj *DiscoveryObject, now string, root string) {
	if fsutil.Exists(filepath.Join(root, "package.json")) {
		obj.Runtimes = append(obj.Runtimes, Runtime{Name: "node", Kind: "javascript", Evidence: []Evidence{ev(now, "manifest", filepath.Join(root, "package.json"), "package.json exists.")}})
		obj.Capabilities = append(obj.Capabilities, "node.package")
		jsonfile.ReadPackageCommands(root, obj.Commands)
	}
	if fsutil.Exists(filepath.Join(root, "go.mod")) {
		obj.Runtimes = append(obj.Runtimes, Runtime{Name: "go", Kind: "go", Evidence: []Evidence{ev(now, "manifest", filepath.Join(root, "go.mod"), "go.mod exists.")}})
		obj.Capabilities = append(obj.Capabilities, "go.module")
		obj.Commands["go:test"] = "go test ./..."
		obj.Commands["go:build"] = "go build ./..."
	}
	if fsutil.Exists(filepath.Join(root, "Cargo.toml")) {
		obj.Runtimes = append(obj.Runtimes, Runtime{Name: "cargo", Kind: "rust", Evidence: []Evidence{ev(now, "manifest", filepath.Join(root, "Cargo.toml"), "Cargo.toml exists.")}})
		obj.Capabilities = append(obj.Capabilities, "rust.crate")
		obj.Commands["cargo:test"] = "cargo test"
		obj.Commands["cargo:build"] = "cargo build"
	}
	if fsutil.Exists(filepath.Join(root, "pyproject.toml")) || fsutil.Exists(filepath.Join(root, "requirements.txt")) {
		obj.Runtimes = append(obj.Runtimes, Runtime{Name: "python", Kind: "python", Evidence: []Evidence{ev(now, "manifest", root, "Python manifest evidence exists.")}})
		obj.Capabilities = append(obj.Capabilities, "python.project")
	}
	if fsutil.Exists(filepath.Join(root, "Dockerfile")) {
		obj.Runtimes = append(obj.Runtimes, Runtime{Name: "docker", Kind: "container", Evidence: []Evidence{ev(now, "manifest", filepath.Join(root, "Dockerfile"), "Dockerfile exists.")}})
		obj.Capabilities = append(obj.Capabilities, "container.image")
	}
}

func discoverAIFTContracts(obj *DiscoveryObject, now string, root string) {
	if fsutil.Exists(filepath.Join(root, ".aift", "module.json")) {
		obj.Capabilities = append(obj.Capabilities, "aift.module")
	}
	if fsutil.Exists(filepath.Join(root, ".aift", "capabilities.json")) {
		obj.Capabilities = append(obj.Capabilities, jsonfile.ReadNamedList(root, "capabilities.json", "capabilities")...)
	}
	if fsutil.Exists(filepath.Join(root, ".aift", "services.json")) {
		obj.Services = append(obj.Services, jsonfile.ReadNamedList(root, "services.json", "services")...)
	}
	if fsutil.Exists(filepath.Join(root, ".aift", "commands", "verify.sh")) {
		obj.Commands["aift:verify"] = "sh .aift/commands/verify.sh"
		obj.HealthChecks = append(obj.HealthChecks, ".aift/commands/verify.sh")
	}
}

func List(cfg config.Config) error {
	snap, err := LoadOrBuild(cfg)
	if err != nil {
		return err
	}
	fmt.Printf("%-40s %-12s %-12s %-24s %s\n", "OBJECT", "KIND", "STATUS", "RUNTIMES", "NAME")
	for _, obj := range snap.Objects {
		fmt.Printf("%-40s %-12s %-12s %-24s %s\n", obj.ID, obj.Kind, obj.Status, runtimeNames(obj.Runtimes), obj.Name)
	}
	return nil
}

func ObjectInfo(cfg config.Config, id string) error {
	snap, err := LoadOrBuild(cfg)
	if err != nil {
		return err
	}
	for _, obj := range snap.Objects {
		if obj.ID == id || obj.Name == id {
			data, err := json.MarshalIndent(obj, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}
	}
	return fmt.Errorf("discovery object not found: %s", id)
}

func Report(cfg config.Config) error {
	snap, err := LoadOrBuild(cfg)
	if err != nil {
		return err
	}
	if err := WriteReport(cfg, snap); err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(cfg.OSHome, "reports", "discovery.md"))
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

func Write(cfg config.Config, snap Snapshot) error {
	return jsonfile.Write(filepath.Join(cfg.OSHome, "registry", "discovery.json"), snap, false)
}

func WriteReport(cfg config.Config, snap Snapshot) error {
	out := filepath.Join(cfg.OSHome, "reports", "discovery.md")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# AIFT Discovery Report\n\n")
	b.WriteString("Generated from filesystem, Git, manifest, runtime, documentation, workflow, schema, and AIFT contract evidence.\n\n")
	b.WriteString("| Object | Status | Runtimes | Manifests | Commands | Capabilities | Services |\n")
	b.WriteString("|---|---|---|---:|---:|---:|---:|\n")

	for _, obj := range snap.Objects {
		b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%d` | `%d` | `%d` | `%d` |\n",
			obj.ID,
			obj.Status,
			runtimeNames(obj.Runtimes),
			len(obj.Manifests),
			len(obj.Commands),
			len(obj.Capabilities),
			len(obj.Services),
		))
	}

	return os.WriteFile(out, []byte(b.String()), 0644)
}

func LoadOrBuild(cfg config.Config) (Snapshot, error) {
	path := filepath.Join(cfg.OSHome, "registry", "discovery.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return Build(cfg)
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return Snapshot{}, err
	}
	return snap, nil
}

func addEvidence(obj *DiscoveryObject, now string, kind string, path string, description string) {
	obj.Evidence = append(obj.Evidence, ev(now, kind, path, description))
}

func ev(now string, kind string, path string, description string) Evidence {
	return Evidence{
		Kind:        kind,
		Path:        path,
		Description: description,
		ObservedAt:  now,
	}
}

func runtimeNames(runtimes []Runtime) string {
	names := []string{}
	for _, rt := range runtimes {
		names = append(names, rt.Name)
	}
	return strings.Join(sliceutil.Unique(names), ",")
}

func safeID(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "/", ".")
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, "_", "-")
	return value
}


