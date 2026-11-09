package uxihttp

import (
	"os"
	"strings"
	"testing"
)

// The imported portal graph is untrusted and must remain an isolated,
// evidence-linked exploration surface rather than a local trust assertion.
func TestExpeditionPortableLineageTraversalContract(t *testing.T) {
	b, err := os.ReadFile("expedition.html")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	required := []string{
		"portable-trail", "portable-find", "portable-back", "portable-all",
		"function portableNavigate(id)", "portableHistory.push(portableFocus)",
		"portableHistory.pop()", "const related=new Set()",
		"if(e.from===focus.id)related.add(e.to)",
		"if(e.to===focus.id)related.add(e.from)",
		"packet.edges.filter(e=>e.to===node.id)",
		"packet.edges.filter(e=>e.from===node.id)",
		"No independent attestation",
		"portableGraph=packet",
		"a.rel=\"noopener noreferrer\"",
	}
	for _, token := range required {
		if !strings.Contains(s, token) {
			t.Errorf("portable lineage traversal missing %q", token)
		}
	}
	forbidden := []string{"localStorage.setItem(\"portable", "fetch(node.url)", "verifiedProducts.push("}
	for _, token := range forbidden {
		if strings.Contains(s, token) {
			t.Errorf("unsafe portable lineage behavior %q", token)
		}
	}
}
