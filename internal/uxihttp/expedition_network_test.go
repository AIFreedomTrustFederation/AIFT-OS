package uxihttp

import (
	"strings"
	"testing"
)

func TestExpeditionRootNetworkObservatory(t *testing.T) {
	checks := []string{
		"Root Network Observatory",
		"Living Root Atlas",
		"function renderAtlas()",
		"function setAtlasZoom(next)",
		"aria-label=\\\"Spatial repository root atlas\\\"",
		"No roots plotted; no provenance inferred.",
		"Evidence Readiness Observatory",
		"function renderLedger()",
		"What would verify a fruit?",
		"id=\"root-network\"",
		"id=\"root-sort\"",
		"id=\"clear-selection\"",
		"function renderNetwork()",
		"function inspect(item)",
		"No source-to-product edges inferred.",
		"prefers-reduced-motion",
		"/v1/federation/tree",
	}
	for _, check := range checks {
		if !strings.Contains(expeditionHTML, check) {
			t.Errorf("expedition must contain %q", check)
		}
	}
	if strings.Contains(expeditionHTML, "role=\"listitem\"") {
		t.Fatal("interactive buttons must not be assigned listitem role")
	}
}
