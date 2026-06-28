package graph

import (
	"strings"
	"testing"
)

func TestSafeID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"federation:root", "federation_root"},
		{"repo:my-app", "repo_my_app"},
		{"node.js/dep", "node_js_dep"},
		{"hello world", "hello_world"},
		{"simple", "simple"},
		{"", ""},
	}
	for _, tt := range tests {
		got := safeID(tt.input)
		if got != tt.want {
			t.Errorf("safeID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestXmlEsc(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"a & b", "a &amp; b"},
		{"<tag>", "&lt;tag&gt;"},
		{`say "hi"`, "say &quot;hi&quot;"},
		{"it's", "it&apos;s"},
		{"a & <b> \"c\" 'd'", "a &amp; &lt;b&gt; &quot;c&quot; &apos;d&apos;"},
		{"", ""},
	}
	for _, tt := range tests {
		got := xmlEsc(tt.input)
		if got != tt.want {
			t.Errorf("xmlEsc(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCypherType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"depends_on", "DEPENDS_ON"},
		{"contains", "CONTAINS"},
		{"", "RELATED_TO"},
		{"built-by", "BUILT_BY"},
	}
	for _, tt := range tests {
		got := cypherType(tt.input)
		if got != tt.want {
			t.Errorf("cypherType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]int{"banana": 2, "apple": 1, "cherry": 3}
	got := sortedKeys(m)
	want := []string{"apple", "banana", "cherry"}
	if len(got) != len(want) {
		t.Fatalf("sortedKeys returned %d keys, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("sortedKeys[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSortedKeysEmpty(t *testing.T) {
	m := map[string]int{}
	got := sortedKeys(m)
	if len(got) != 0 {
		t.Errorf("sortedKeys(empty) returned %d keys, want 0", len(got))
	}
}

func TestAddNode(t *testing.T) {
	b := builder{nodes: map[string]Node{}, edges: map[string]Edge{}}

	b.addNode(Node{ID: "n1", Type: "Test", Name: "node1", Status: "active", Evidence: "test"})
	if len(b.nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(b.nodes))
	}
	if b.nodes["n1"].Name != "node1" {
		t.Errorf("node name = %q, want %q", b.nodes["n1"].Name, "node1")
	}
}

func TestAddNodeDefaultStatus(t *testing.T) {
	b := builder{nodes: map[string]Node{}, edges: map[string]Edge{}}
	b.addNode(Node{ID: "n1", Type: "Test", Name: "node1"})

	if b.nodes["n1"].Status != "unknown" {
		t.Errorf("default status = %q, want %q", b.nodes["n1"].Status, "unknown")
	}
}

func TestAddNodeDefaultEvidence(t *testing.T) {
	b := builder{nodes: map[string]Node{}, edges: map[string]Edge{}}
	b.addNode(Node{ID: "n1", Type: "Test", Name: "node1", Status: "ok"})

	if b.nodes["n1"].Evidence != "not specified" {
		t.Errorf("default evidence = %q, want %q", b.nodes["n1"].Evidence, "not specified")
	}
}

func TestAddNodeOverwrite(t *testing.T) {
	b := builder{nodes: map[string]Node{}, edges: map[string]Edge{}}
	b.addNode(Node{ID: "n1", Type: "A", Name: "first", Status: "a", Evidence: "e"})
	b.addNode(Node{ID: "n1", Type: "B", Name: "second", Status: "b", Evidence: "e"})

	if b.nodes["n1"].Name != "second" {
		t.Errorf("overwritten node name = %q, want %q", b.nodes["n1"].Name, "second")
	}
}

func TestAddEdge(t *testing.T) {
	b := builder{nodes: map[string]Node{}, edges: map[string]Edge{}}
	b.addEdge("a", "b", "contains", "test evidence")

	if len(b.edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(b.edges))
	}

	key := "a|contains|b"
	e, ok := b.edges[key]
	if !ok {
		t.Fatalf("edge key %q not found", key)
	}
	if e.From != "a" || e.To != "b" || e.Type != "contains" {
		t.Errorf("edge = %+v, want from=a to=b type=contains", e)
	}
}

func TestAddEdgeSkipsEmpty(t *testing.T) {
	b := builder{nodes: map[string]Node{}, edges: map[string]Edge{}}
	b.addEdge("", "b", "contains", "test")
	b.addEdge("a", "", "contains", "test")
	b.addEdge("", "", "contains", "test")

	if len(b.edges) != 0 {
		t.Errorf("expected 0 edges for empty from/to, got %d", len(b.edges))
	}
}

func TestBuilderGraph(t *testing.T) {
	b := builder{nodes: map[string]Node{}, edges: map[string]Edge{}}
	b.addNode(Node{ID: "b-node", Type: "T", Name: "B", Status: "ok", Evidence: "e"})
	b.addNode(Node{ID: "a-node", Type: "T", Name: "A", Status: "ok", Evidence: "e"})
	b.addEdge("a-node", "b-node", "links", "test")

	g := b.graph()

	if len(g.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(g.Nodes))
	}
	if g.Nodes[0].ID != "a-node" {
		t.Errorf("nodes should be sorted by ID; first = %q", g.Nodes[0].ID)
	}
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g.Edges))
	}
	if g.GeneratedAt == "" {
		t.Error("GeneratedAt should not be empty")
	}
}

func TestBuilderGraphEdgeSorting(t *testing.T) {
	b := builder{nodes: map[string]Node{}, edges: map[string]Edge{}}
	b.addEdge("z", "a", "alpha", "e")
	b.addEdge("a", "b", "beta", "e")
	b.addEdge("a", "c", "alpha", "e")
	b.addEdge("a", "b", "alpha", "e")

	g := b.graph()

	if len(g.Edges) != 4 {
		t.Fatalf("expected 4 edges, got %d", len(g.Edges))
	}
	if g.Edges[0].From != "a" || g.Edges[0].Type != "alpha" || g.Edges[0].To != "b" {
		t.Errorf("first edge should be a|alpha|b, got %s|%s|%s", g.Edges[0].From, g.Edges[0].Type, g.Edges[0].To)
	}
}

func TestWriteMermaidFormat(t *testing.T) {
	g := Graph{
		Edges: []Edge{
			{From: "repo:a", To: "dep:b", Type: "depends_on"},
		},
	}

	var b strings.Builder
	b.WriteString("graph TD\n")
	for _, e := range g.Edges {
		b.WriteString("  " + safeID(e.From) + " -->|" + e.Type + "| " + safeID(e.To) + "\n")
	}

	out := b.String()
	if !strings.Contains(out, "graph TD") {
		t.Error("mermaid output should start with graph TD")
	}
	if !strings.Contains(out, "repo_a -->|depends_on| dep_b") {
		t.Errorf("mermaid output missing expected edge line, got: %s", out)
	}
}

func TestPrintMatchingRepo(t *testing.T) {
	g := Graph{
		Nodes: []Node{
			{ID: "n1", Type: "Repo", Name: "myapp", Repo: "myapp", Status: "ok"},
			{ID: "n2", Type: "Dep", Name: "other", Repo: "other", Status: "ok"},
		},
	}

	matched := 0
	for _, n := range g.Nodes {
		if n.Repo == "myapp" || n.Name == "myapp" {
			matched++
		}
	}
	if matched != 1 {
		t.Errorf("expected 1 match for repo=myapp, got %d", matched)
	}
}

func TestPrintMatchingType(t *testing.T) {
	g := Graph{
		Nodes: []Node{
			{ID: "n1", Type: "Repository", Name: "a"},
			{ID: "n2", Type: "Dependency", Name: "b"},
			{ID: "n3", Type: "repository", Name: "c"},
		},
	}

	matched := 0
	for _, n := range g.Nodes {
		if strings.EqualFold(n.Type, "repository") {
			matched++
		}
	}
	if matched != 2 {
		t.Errorf("expected 2 case-insensitive matches for type=repository, got %d", matched)
	}
}

func TestPrintMatchingStatus(t *testing.T) {
	g := Graph{
		Nodes: []Node{
			{ID: "n1", Status: "detected"},
			{ID: "n2", Status: "planned"},
			{ID: "n3", Status: "detected"},
		},
	}

	matched := 0
	for _, n := range g.Nodes {
		if n.Status == "detected" {
			matched++
		}
	}
	if matched != 2 {
		t.Errorf("expected 2 matches for status=detected, got %d", matched)
	}
}
