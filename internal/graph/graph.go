package graph

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
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Node struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	Repo       string            `json:"repo,omitempty"`
	Status     string            `json:"status"`
	Evidence   string            `json:"evidence"`
	Properties map[string]string `json:"properties,omitempty"`
}

type Edge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Type     string `json:"type"`
	Evidence string `json:"evidence"`
}

type Graph struct {
	GeneratedAt string `json:"generatedAt"`
	Nodes       []Node `json:"nodes"`
	Edges       []Edge `json:"edges"`
}

type builder struct {
	nodes map[string]Node
	edges map[string]Edge
}

func Build(cfg config.Config) error {
	g, err := Discover(cfg)
	if err != nil {
		return err
	}

	if err := writeAll(cfg, g); err != nil {
		return err
	}

	return events.Emit(cfg, "graph.build", "graph", "federation graph built", map[string]string{
		"nodes": fmt.Sprint(len(g.Nodes)),
		"edges": fmt.Sprint(len(g.Edges)),
	})
}

func Discover(cfg config.Config) (Graph, error) {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return Graph{}, err
	}

	b := builder{
		nodes: map[string]Node{},
		edges: map[string]Edge{},
	}

	b.addNode(Node{
		ID:       "federation:root",
		Type:     "Federation",
		Name:     "AI Freedom Trust Federation",
		Status:   "detected",
		Evidence: "AIFT workspace root discovered",
	})

	for _, r := range repos {
		b.discoverRepo(cfg, r)
	}

	return b.graph(), nil
}

func (b *builder) discoverRepo(cfg config.Config, r workspace.Repo) {
	repoID := "repo:" + r.Name

	b.addNode(Node{
		ID:       repoID,
		Type:     "Repository",
		Name:     r.Name,
		Repo:     r.Name,
		Status:   "detected",
		Evidence: ".git directory discovered by workspace scanner",
		Properties: map[string]string{
			"path": r.Path,
		},
	})
	b.addEdge("federation:root", repoID, "contains", "repository discovered under AIFT workspace")

	b.fileNode(r, ".aift/repo.json", "Manifest", "repo-manifest", repoID, "declares")
	b.fileNode(r, ".aift/capabilities.json", "CapabilitySet", "capabilities", repoID, "declares")
	b.fileNode(r, ".aift/manual.json", "ManualContract", "manual-contract", repoID, "declares")
	b.fileNode(r, "package.json", "Package", "npm-package", repoID, "contains")
	b.fileNode(r, "go.mod", "Package", "go-module", repoID, "contains")
	b.fileNode(r, "README.md", "ManualPage", "readme", repoID, "documents")
	b.fileNode(r, "docs/manual/source/index.md", "Manual", "unix-manual-source", repoID, "documents")

	b.discoverCapabilities(r, repoID)
	b.discoverManual(r, repoID)
	b.discoverPackageScripts(r, repoID)
	b.discoverGoModule(r, repoID)
}

func (b *builder) fileNode(r workspace.Repo, rel string, typ string, name string, repoID string, edgeType string) {
	path := filepath.Join(r.Path, rel)
	if !exists(path) {
		return
	}
	id := typ + ":" + r.Name + ":" + name
	b.addNode(Node{
		ID:       id,
		Type:     typ,
		Name:     name,
		Repo:     r.Name,
		Status:   "detected",
		Evidence: rel + " exists",
		Properties: map[string]string{
			"path": rel,
		},
	})
	b.addEdge(repoID, id, edgeType, rel+" exists")
}

func (b *builder) discoverCapabilities(r workspace.Repo, repoID string) {
	path := filepath.Join(r.Path, ".aift", "capabilities.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var rc struct {
		Capabilities []struct {
			Name     string `json:"name"`
			Status   string `json:"status"`
			Version  int    `json:"version"`
			Command  string `json:"command"`
			Evidence string `json:"evidence"`
		} `json:"capabilities"`
	}
	if json.Unmarshal(data, &rc) != nil {
		return
	}

	for _, cap := range rc.Capabilities {
		id := "capability:" + r.Name + ":" + cap.Name
		status := cap.Status
		if status == "" {
			status = "unknown"
		}
		evidence := cap.Evidence
		if evidence == "" {
			evidence = ".aift/capabilities.json"
		}

		b.addNode(Node{
			ID:       id,
			Type:     "Capability",
			Name:     cap.Name,
			Repo:     r.Name,
			Status:   status,
			Evidence: evidence,
			Properties: map[string]string{
				"version": fmt.Sprint(cap.Version),
				"command": cap.Command,
			},
		})
		b.addEdge(repoID, id, "provides", ".aift/capabilities.json declares capability")

		if cap.Command != "" {
			cmdID := "command:" + r.Name + ":" + cap.Name
			b.addNode(Node{
				ID:       cmdID,
				Type:     "Command",
				Name:     cap.Command,
				Repo:     r.Name,
				Status:   status,
				Evidence: "capability command declared",
			})
			b.addEdge(id, cmdID, "implemented_by", "capability command field")
		}
	}
}

func (b *builder) discoverManual(r workspace.Repo, repoID string) {
	path := filepath.Join(r.Path, ".aift", "manual.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var m struct {
		Title      string `json:"title"`
		SourcePath string `json:"sourcePath"`
		Builder    struct {
			Repo       string `json:"repo"`
			Capability string `json:"capability"`
			Status     string `json:"status"`
		} `json:"builder"`
		Status struct {
			ManualSource string `json:"manualSource"`
			PDFBuild     string `json:"pdfBuild"`
			WebPublish   string `json:"webPublish"`
		} `json:"status"`
	}
	if json.Unmarshal(data, &m) != nil {
		return
	}

	manualID := "manual:" + r.Name
	b.addNode(Node{
		ID:       manualID,
		Type:     "Manual",
		Name:     m.Title,
		Repo:     r.Name,
		Status:   m.Status.ManualSource,
		Evidence: ".aift/manual.json",
		Properties: map[string]string{
			"sourcePath": m.SourcePath,
			"pdfBuild":   m.Status.PDFBuild,
			"webPublish": m.Status.WebPublish,
		},
	})
	b.addEdge(repoID, manualID, "documents", ".aift/manual.json declares manual")

	builderID := "repo:" + m.Builder.Repo
	capID := "capability:" + m.Builder.Repo + ":" + m.Builder.Capability

	b.addNode(Node{
		ID:       capID,
		Type:     "Capability",
		Name:     m.Builder.Capability,
		Repo:     m.Builder.Repo,
		Status:   m.Builder.Status,
		Evidence: ".aift/manual.json builder declaration",
	})
	b.addEdge(manualID, capID, "built_by", "manual builder capability declaration")
	b.addEdge(capID, builderID, "owned_by", "manual builder repository declaration")
}

func (b *builder) discoverPackageScripts(r workspace.Repo, repoID string) {
	path := filepath.Join(r.Path, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var pkg struct {
		Name            string            `json:"name"`
		Version         string            `json:"version"`
		Scripts         map[string]string `json:"scripts"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return
	}

	pkgID := "package:" + r.Name + ":npm"
	for script, command := range pkg.Scripts {
		id := "script:" + r.Name + ":" + script
		b.addNode(Node{
			ID:       id,
			Type:     "Script",
			Name:     script,
			Repo:     r.Name,
			Status:   "detected",
			Evidence: "package.json scripts." + script,
			Properties: map[string]string{
				"command": command,
			},
		})
		b.addEdge(pkgID, id, "provides_script", "package.json scripts")
	}

	for dep, version := range pkg.Dependencies {
		b.addDependency(r, repoID, dep, version, "runtime", "package.json dependencies")
	}
	for dep, version := range pkg.DevDependencies {
		b.addDependency(r, repoID, dep, version, "development", "package.json devDependencies")
	}
}

func (b *builder) discoverGoModule(r workspace.Repo, repoID string) {
	path := filepath.Join(r.Path, "go.mod")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	moduleName := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}

	id := "module:" + r.Name + ":go"
	b.addNode(Node{
		ID:       id,
		Type:     "Module",
		Name:     moduleName,
		Repo:     r.Name,
		Status:   "detected",
		Evidence: "go.mod module declaration",
	})
	b.addEdge(repoID, id, "contains", "go.mod exists")
}

func (b *builder) addDependency(r workspace.Repo, repoID, dep, version, scope, evidence string) {
	id := "dependency:" + dep
	b.addNode(Node{
		ID:       id,
		Type:     "Dependency",
		Name:     dep,
		Status:   "detected",
		Evidence: evidence,
		Properties: map[string]string{
			"version": version,
			"scope":   scope,
		},
	})
	b.addEdge(repoID, id, "depends_on", evidence)
}

func (b *builder) addNode(n Node) {
	if n.Status == "" {
		n.Status = "unknown"
	}
	if n.Evidence == "" {
		n.Evidence = "not specified"
	}
	b.nodes[n.ID] = n
}

func (b *builder) addEdge(from, to, typ, evidence string) {
	if from == "" || to == "" {
		return
	}
	key := from + "|" + typ + "|" + to
	b.edges[key] = Edge{
		From:     from,
		To:       to,
		Type:     typ,
		Evidence: evidence,
	}
}

func (b *builder) graph() Graph {
	nodes := make([]Node, 0, len(b.nodes))
	for _, n := range b.nodes {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })

	edges := make([]Edge, 0, len(b.edges))
	for _, e := range b.edges {
		edges = append(edges, e)
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From == edges[j].From {
			if edges[i].Type == edges[j].Type {
				return edges[i].To < edges[j].To
			}
			return edges[i].Type < edges[j].Type
		}
		return edges[i].From < edges[j].From
	})

	return Graph{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Nodes:       nodes,
		Edges:       edges,
	}
}

func writeAll(cfg config.Config, g Graph) error {
	if err := writeJSON(cfg, g); err != nil {
		return err
	}
	if err := writeMermaid(cfg, g); err != nil {
		return err
	}
	if err := writeDOT(cfg, g); err != nil {
		return err
	}
	if err := writeGraphML(cfg, g); err != nil {
		return err
	}
	if err := writeCypher(cfg, g); err != nil {
		return err
	}
	if err := writeRDF(cfg, g); err != nil {
		return err
	}
	if err := writeReports(cfg, g); err != nil {
		return err
	}
	return nil
}

func writeJSON(cfg config.Config, g Graph) error {
	out := filepath.Join(cfg.OSHome, "registry", "graph.json")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(out, append(data, '\n'), 0644); err != nil {
		return err
	}
	fmt.Println("Wrote", out)
	return nil
}

func writeMermaid(cfg config.Config, g Graph) error {
	out := filepath.Join(cfg.OSHome, "registry", "graph.mermaid")
	var b strings.Builder
	b.WriteString("graph TD\n")
	for _, e := range g.Edges {
		b.WriteString(fmt.Sprintf("  %s -->|%s| %s\n", safeID(e.From), e.Type, safeID(e.To)))
	}
	return os.WriteFile(out, []byte(b.String()), 0644)
}

func writeDOT(cfg config.Config, g Graph) error {
	out := filepath.Join(cfg.OSHome, "registry", "graph.dot")
	var b strings.Builder
	b.WriteString("digraph AIFT {\n")
	for _, n := range g.Nodes {
		b.WriteString(fmt.Sprintf("  %q [label=%q];\n", n.ID, n.Name+"\\n"+n.Type+"\\n"+n.Status))
	}
	for _, e := range g.Edges {
		b.WriteString(fmt.Sprintf("  %q -> %q [label=%q];\n", e.From, e.To, e.Type))
	}
	b.WriteString("}\n")
	return os.WriteFile(out, []byte(b.String()), 0644)
}

func writeGraphML(cfg config.Config, g Graph) error {
	out := filepath.Join(cfg.OSHome, "registry", "graph.graphml")
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<graphml xmlns="http://graphml.graphdrawing.org/xmlns"><graph edgedefault="directed">` + "\n")
	for _, n := range g.Nodes {
		b.WriteString(fmt.Sprintf(`<node id="%s"><data key="type">%s</data><data key="name">%s</data><data key="status">%s</data></node>`+"\n", xmlEsc(n.ID), xmlEsc(n.Type), xmlEsc(n.Name), xmlEsc(n.Status)))
	}
	for i, e := range g.Edges {
		b.WriteString(fmt.Sprintf(`<edge id="e%d" source="%s" target="%s"><data key="type">%s</data></edge>`+"\n", i, xmlEsc(e.From), xmlEsc(e.To), xmlEsc(e.Type)))
	}
	b.WriteString("</graph></graphml>\n")
	return os.WriteFile(out, []byte(b.String()), 0644)
}

func writeCypher(cfg config.Config, g Graph) error {
	out := filepath.Join(cfg.OSHome, "registry", "graph.cypher")
	var b strings.Builder
	for _, n := range g.Nodes {
		b.WriteString(fmt.Sprintf("MERGE (n:Node {id:%q}) SET n.type=%q, n.name=%q, n.status=%q;\n", n.ID, n.Type, n.Name, n.Status))
	}
	for _, e := range g.Edges {
		b.WriteString(fmt.Sprintf("MATCH (a:Node {id:%q}), (b:Node {id:%q}) MERGE (a)-[:%s {evidence:%q}]->(b);\n", e.From, e.To, cypherType(e.Type), e.Evidence))
	}
	return os.WriteFile(out, []byte(b.String()), 0644)
}

func writeRDF(cfg config.Config, g Graph) error {
	out := filepath.Join(cfg.OSHome, "registry", "graph.rdf")
	var b strings.Builder
	b.WriteString("@prefix aift: <https://aifreedomtrust.org/aift#> .\n\n")
	for _, n := range g.Nodes {
		b.WriteString(fmt.Sprintf("aift:%s a aift:%s ; aift:name %q ; aift:status %q .\n", safeID(n.ID), safeID(n.Type), n.Name, n.Status))
	}
	for _, e := range g.Edges {
		b.WriteString(fmt.Sprintf("aift:%s aift:%s aift:%s .\n", safeID(e.From), safeID(e.Type), safeID(e.To)))
	}
	return os.WriteFile(out, []byte(b.String()), 0644)
}

func writeReports(cfg config.Config, g Graph) error {
	if err := os.MkdirAll(filepath.Join(cfg.OSHome, "reports"), 0755); err != nil {
		return err
	}

	typeCounts := map[string]int{}
	statusCounts := map[string]int{}
	for _, n := range g.Nodes {
		typeCounts[n.Type]++
		statusCounts[n.Status]++
	}

	var b strings.Builder
	b.WriteString("# Federation Knowledge Graph\n\n")
	b.WriteString(fmt.Sprintf("- Nodes: %d\n", len(g.Nodes)))
	b.WriteString(fmt.Sprintf("- Edges: %d\n\n", len(g.Edges)))
	b.WriteString("## Node Types\n\n")
	for _, k := range sortedKeys(typeCounts) {
		b.WriteString(fmt.Sprintf("- `%s`: %d\n", k, typeCounts[k]))
	}
	b.WriteString("\n## Statuses\n\n")
	for _, k := range sortedKeys(statusCounts) {
		b.WriteString(fmt.Sprintf("- `%s`: %d\n", k, statusCounts[k]))
	}
	b.WriteString("\n## Edges\n\n")
	for _, e := range g.Edges {
		b.WriteString(fmt.Sprintf("- `%s` --%s--> `%s` (%s)\n", e.From, e.Type, e.To, e.Evidence))
	}
	if err := os.WriteFile(filepath.Join(cfg.OSHome, "reports", "graph.md"), []byte(b.String()), 0644); err != nil {
		return err
	}

	summary := fmt.Sprintf("# Graph Summary\n\nNodes: %d\n\nEdges: %d\n\nGenerated: %s\n", len(g.Nodes), len(g.Edges), g.GeneratedAt)
	if err := os.WriteFile(filepath.Join(cfg.OSHome, "reports", "graph-summary.md"), []byte(summary), 0644); err != nil {
		return err
	}

	planned := "# Planned vs Running\n\n"
	for _, n := range g.Nodes {
		if n.Status == "planned" || n.Status == "ready" || n.Status == "v1" || n.Status == "running" || n.Status == "detected" || n.Status == "broken" {
			planned += fmt.Sprintf("- `%s` `%s` `%s`\n", n.Status, n.Type, n.Name)
		}
	}
	if err := os.WriteFile(filepath.Join(cfg.OSHome, "reports", "planned-vs-running.md"), []byte(planned), 0644); err != nil {
		return err
	}

	serviceMap := "# Service Map\n\n"
	for _, n := range g.Nodes {
		if n.Type == "Service" || n.Type == "Provider" || n.Type == "Capability" {
			serviceMap += fmt.Sprintf("- `%s` `%s` `%s`\n", n.Type, n.Status, n.Name)
		}
	}
	if err := os.WriteFile(filepath.Join(cfg.OSHome, "reports", "service-map.md"), []byte(serviceMap), 0644); err != nil {
		return err
	}

	orphaned := "# Orphaned Capabilities\n\n"
	hasIncoming := map[string]bool{}
	for _, e := range g.Edges {
		hasIncoming[e.To] = true
	}
	for _, n := range g.Nodes {
		if n.Type == "Capability" && !hasIncoming[n.ID] {
			orphaned += fmt.Sprintf("- `%s` `%s`\n", n.Status, n.ID)
		}
	}
	if err := os.WriteFile(filepath.Join(cfg.OSHome, "reports", "orphaned-capabilities.md"), []byte(orphaned), 0644); err != nil {
		return err
	}

	deps := "# Dependency Tree\n\n"
	for _, e := range g.Edges {
		if e.Type == "depends_on" {
			deps += fmt.Sprintf("- `%s` depends on `%s`\n", e.From, e.To)
		}
	}
	if err := os.WriteFile(filepath.Join(cfg.OSHome, "reports", "dependency-tree.md"), []byte(deps), 0644); err != nil {
		return err
	}

	return nil
}

func Query(cfg config.Config, args []string) error {
	if len(args) == 0 {
		return Build(cfg)
	}

	data, err := os.ReadFile(filepath.Join(cfg.OSHome, "registry", "graph.json"))
	if err != nil {
		if err := Build(cfg); err != nil {
			return err
		}
		data, err = os.ReadFile(filepath.Join(cfg.OSHome, "registry", "graph.json"))
		if err != nil {
			return err
		}
	}

	var g Graph
	if err := json.Unmarshal(data, &g); err != nil {
		return err
	}

	switch args[0] {
	case "summary":
		fmt.Printf("Nodes: %d\nEdges: %d\n", len(g.Nodes), len(g.Edges))
	case "repo":
		if len(args) < 2 {
			return fmt.Errorf("usage: aift graph repo <name>")
		}
		return printMatching(g, "repo", args[1])
	case "type":
		if len(args) < 2 {
			return fmt.Errorf("usage: aift graph type <type>")
		}
		return printMatching(g, "type", args[1])
	case "status":
		if len(args) < 2 {
			return fmt.Errorf("usage: aift graph status <status>")
		}
		return printMatching(g, "status", args[1])
	default:
		return fmt.Errorf("usage: aift graph [summary|repo|type|status]")
	}
	return nil
}

func printMatching(g Graph, field, value string) error {
	for _, n := range g.Nodes {
		ok := false
		switch field {
		case "repo":
			ok = n.Repo == value || n.Name == value
		case "type":
			ok = strings.EqualFold(n.Type, value)
		case "status":
			ok = n.Status == value
		}
		if ok {
			fmt.Printf("%-28s %-18s %-10s %s\n", n.Type, n.Status, n.Repo, n.Name)
		}
	}
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func safeID(s string) string {
	repl := strings.NewReplacer(":", "_", "-", "_", ".", "_", "/", "_", " ", "_")
	return repl.Replace(s)
}

func xmlEsc(s string) string {
	repl := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")
	return repl.Replace(s)
}

func cypherType(s string) string {
	s = strings.ToUpper(safeID(s))
	if s == "" {
		return "RELATED_TO"
	}
	return s
}

func sortedKeys(m map[string]int) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
