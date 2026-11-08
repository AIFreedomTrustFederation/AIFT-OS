package uxihttp

import (
	"os"
	"strings"
	"testing"
)

func TestExpeditionImmersiveTheatreContract(t *testing.T) {
	page, err := os.ReadFile("expedition.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(page)
	required := []string{
		"Tree of Emergence · Immersive Lineage Theatre",
		"id=\"theatre-scene\"",
		"id=\"theatre-root-list\"",
		"id=\"theatre-detail\"",
		"function renderTheatre()",
		"function theatreSelect(root)",
		"state.items.slice().sort",
		"root.name",
		"aria-pressed",
		"prefers-reduced-motion",
		"No root-to-trunk lines are asserted without evidence",
		"No verified source-to-product lineage is established.",
		"renderMatrix();renderTheatre();",
	}
	for _, marker := range required {
		if !strings.Contains(html, marker) {
			t.Errorf("missing immersive theatre contract: %q", marker)
		}
	}
	forbidden := []string{
		"fetch(\"https://",
		"localStorage.setItem",
		"Math.random()",
	}
	for _, marker := range forbidden {
		if strings.Contains(html, marker) {
			t.Errorf("unexpected network, persistence, or fabricated randomness: %q", marker)
		}
	}
}
