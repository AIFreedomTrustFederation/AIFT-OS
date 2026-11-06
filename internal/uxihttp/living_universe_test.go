package uxihttp

import (
	"os"
	"strings"
	"testing"
)

// A Living Universe node must preserve the source-to-outcome biological taxonomy.
func TestLivingUniverseLocalEvidenceContract(t *testing.T) {
	raw, err := os.ReadFile("index.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(raw)
	required := []string{
		`id="universeTab"`,
		`id="universeView"`,
		`id="universeSvg"`,
		`id="universeIndex"`,
		`id="universeSearch"`,
		`id="universeKind"`,
		`function universeModel()`,
		`kind:"root"`,
		`kind:"trunk"`,
		`kind:"leaf"`,
		`state.tree?.nodes`,
		`root.repo.capabilities`,
		`function selectUniverse(node)`,
		`function renderUniverse()`,
		`Local projection from existing MoBox evidence`,
		`No fruit is invented from repository status`,
		`living-universe.html`,
		`@media(prefers-reduced-motion:reduce)`,
	}
	for _, marker := range required {
		if !strings.Contains(html, marker) {
			t.Errorf("missing Living Universe contract %q", marker)
		}
	}
	if strings.Contains(html, `kind:"fruit",name:repo.name`) {
		t.Error("repositories must never be represented as fruits")
	}
}
