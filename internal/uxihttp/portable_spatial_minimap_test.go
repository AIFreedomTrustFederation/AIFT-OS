package uxihttp

import (
	"os"
	"strings"
	"testing"
)

func TestPortableSpatialMinimapContract(t *testing.T) {
	b, err := os.ReadFile("expedition.html")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, marker := range []string{
		"mini.dataset.spatialOverview",
		"const overview=atlasElement(\"svg\"",
		"const activeRoute=new Set(",
		"const windowRect=atlasElement(\"rect\"",
		"function updateAtlasMinimapViewport()",
		"rect.setAttribute(\"width\",String(1000/atlasViewport.scale))",
		"updateAtlasMinimapViewport();",
		"mini.replaceChildren();",
		"pointer-events",
	} {
		if !strings.Contains(s, marker) {
			t.Errorf("spatial minimap contract missing %q", marker)
		}
	}
	if !strings.Contains(s, "aria-hidden=\"true\"") {
		t.Error("spatial overview must retain an accessible text alternative")
	}
}
