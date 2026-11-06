package uxihttp

import (
	"os"
	"strings"
	"testing"
)

func TestUniverseEvidenceExpeditionContract(t *testing.T) {
	data, err := os.ReadFile("index.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	for _, marker := range []string{
		`id="universeExpeditionTitle"`,
		`id="universeQuests"`,
		`id="universeQuestStatus"`,
		`const universeVisited=new Set()`,
		`function renderUniverseQuests()`,
		`Rootkeeper`,
		`Knowledge Gardener`,
		`universeVisited.add(target.id)`,
		`no product fruits or live health inferred.`,
	} {
		if !strings.Contains(html, marker) {
			t.Errorf("missing expedition marker %q", marker)
		}
	}
}

// TestGameTabCompositionPreservesUniverseNavigation guards against coupling the
// Tree of Life button to the closing nav tag when additional tabs are present.
func TestGameTabCompositionPreservesUniverseNavigation(t *testing.T) {
	html := composeIndexHTML()
	for _, marker := range []string{
		`id="treeTab" role="tab" aria-selected="false">Tree Game</button>`,
		`id="worldTab" role="tab" aria-selected="false">World Game</button>`,
		`id="universeTab"`,
		`id="universeQuests"`,
		`</nav>`,
	} {
		if !strings.Contains(html, marker) {
			t.Errorf("missing composed navigation marker %q", marker)
		}
	}
	if count := strings.Count(html, `id="worldTab"`); count != 1 {
		t.Errorf("expected exactly one world tab, got %d", count)
	}
}
