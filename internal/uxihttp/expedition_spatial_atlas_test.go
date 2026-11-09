package uxihttp

import (
	"os"
	"strings"
	"testing"
)

// Spatial import is a read-only view over curated, untrusted source claims.
func TestExpeditionPortableSpatialAtlasContract(t *testing.T) {
	b, err := os.ReadFile("expedition.html")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	required := []string{
		"id=\"portable-map\"", "id=\"portable-layer\"", "id=\"portable-map-minimap\"",
		"function renderPortableAtlas(packet)", "renderPortableAtlas(packet);",
		"packet.nodes.filter", "packet.edges.filter", "slice(0,96)", "slice(0,240)",
		"portableNavigate(n.id)", "e.key===\"Enter\"||e.key===\" \"",
		"role:\"button\"", "aria-pressed", "atlasElement(\"title\")",
		"All fruits are unverified candidates", "The complete graph remains in the text directory",
		"prefers-reduced-motion:reduce", "noopener noreferrer",
	}
	for _, token := range required {
		if !strings.Contains(s, token) {
			t.Errorf("spatial atlas missing %q", token)
		}
	}
	forbidden := []string{"fetch(n.url)", "localStorage.setItem(\"portable", "verifiedProducts.push(", "innerHTML=n.label"}
	for _, token := range forbidden {
		if strings.Contains(s, token) {
			t.Errorf("unsafe imported atlas behavior %q", token)
		}
	}
}
