package uxihttp

import (
	"os"
	"strings"
	"testing"
)

func TestEvidenceLineageJourneyContract(t *testing.T) {
	b, err := os.ReadFile("index.html")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	required := []string{
		"id=\"lineageJourneyStatus\" role=\"status\"",
		"id=\"lineagePrev\"", "id=\"lineageNext\"", "id=\"lineageHome\"",
		"function lineagePath(node)", "function renderLineageJourney()",
		"function showLineageJourney(node)", "if(!lineageJourney.some(item=>item.id===node.id))showLineageJourney(node)",
		"aria-current", "data-node-id", "prefers-reduced-motion:reduce",
		"No finished fruit is inferred",
	}
	for _, marker := range required {
		if !strings.Contains(s, marker) {
			t.Errorf("missing lineage journey contract %q", marker)
		}
	}
}

func TestUniverseSpatialCameraContract(t *testing.T) {
	b, err := os.ReadFile("index.html")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, marker := range []string{
		"id=\"universeZoomIn\"", "id=\"universeZoomOut\"",
		"id=\"universeFocus\"", "id=\"universeCameraReset\"",
		"id=\"universeCameraStatus\" role=\"status\"",
		"function applyUniverseCamera()", "universePositions=positions",
		"setAttribute(\"viewBox\"", "universeCamera.zoom>=3",
		"universeCamera.zoom<=1", "prefers-reduced-motion:reduce",
	} {
		if !strings.Contains(s, marker) {
			t.Errorf("missing spatial camera contract %q", marker)
		}
	}
}

func TestUniverseProvenanceObservatoryContract(t *testing.T) {
	b, err := os.ReadFile("index.html")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, marker := range []string{
		"id=\"universeLedgerTitle\"", "id=\"universeLedger\"",
		"id=\"universeLedgerCount\" role=\"status\"",
		"function renderUniverseLedger()", "renderUniverseLedger();applyUniverseCamera()",
		"node.kind===\"root\"?\"SOURCE ROOT · OBSERVED\":\"CAPABILITY LEAF · REPORTED\"",
		"No product outcome can be inferred", "0 certified fruits asserted",
		"selectUniverse(node)", "prefers-reduced-motion: reduce",
	} {
		if !strings.Contains(s, marker) {
			t.Errorf("missing provenance observatory contract %q", marker)
		}
	}
}

func TestEvidenceExpeditionProgressContract(t *testing.T) {
	b, err := os.ReadFile("index.html")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, marker := range []string{
		"Source Cartography", "Capability Atlas", "Lineage Expedition",
		"rootById.has(leaf.parent)", "universeVisited.has(item.id)",
		"document.createElement(\"progress\")",
		"meter.setAttribute(\"aria-label\"",
		"Evidence gap: no qualifying local records.",
		"No XP, CI verification, product fruit or achievement certification inferred.",
	} {
		if !strings.Contains(s, marker) {
			t.Errorf("missing evidence expedition contract %q", marker)
		}
	}
}

func TestFederationNavigatorEvidenceContract(t *testing.T) {
	b, err := os.ReadFile("index.html")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, marker := range []string{
		"id=\"universeNavigationTitle\"",
		"id=\"universeMiniMap\"",
		"id=\"universeNavigationDestinations\"",
		"id=\"universeNavigationReset\"",
		"function renderUniverseNavigator(nodes,positions)",
		"renderUniverseNavigator(nodes,positions)",
		"const byId=new Map(nodes.map(n=>[n.id,n]))",
		"n.kind===\"root\"",
		"aria-live=\"polite\"",
		"No repositories are presented as fruits.",
		"prefers-reduced-motion: reduce",
	} {
		if !strings.Contains(s, marker) {
			t.Errorf("missing federation navigator contract %q", marker)
		}
	}
}

func TestUniverseSemanticEvidenceDeck(t *testing.T) {
	b, err := os.ReadFile("index.html")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, marker := range []string{
		"id=\"universeSemanticTitle\"",
		"id=\"universeSemanticRail\"",
		"id=\"universeSemanticCards\"",
		"id=\"universeSemanticStatus\" role=\"status\"",
		"function renderUniverseSemantic()",
		"renderUniverseSemantic();",
		"fruit:[]",
		"Repository roots are never fruits.",
		"selectUniverse(target)",
		"prefers-reduced-motion:reduce",
	} {
		if !strings.Contains(s, marker) {
			t.Errorf("missing semantic evidence deck contract %q", marker)
		}
	}
}

func TestUniverseEvidenceMatrixContract(t *testing.T) {
 b, err := os.ReadFile("index.html")
 if err != nil { t.Fatal(err) }
 s := string(b)
 for _, marker := range []string{
  "id=\"universeEvidenceTitle\"", "id=\"universeEvidenceSummary\" role=\"status\"",
  "id=\"universeEvidenceMatrix\"", "function renderUniverseEvidenceMatrix()",
  "renderUniverseSemantic();renderUniverseEvidenceMatrix();renderUniverseNavigator(nodes,positions)",
  "rootIds.has(leaf.parent)", "0 certified outcomes", "Evidence gap: no linked capability records.",
  "No relationship is inferred.", "aria-label", "prefers-reduced-motion:reduce",
 } {
  if !strings.Contains(s,marker) { t.Errorf("missing evidence matrix contract %q",marker) }
 }
}
