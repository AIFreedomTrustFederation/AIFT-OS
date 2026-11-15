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
