package uxihttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExpeditionBiologicalEvidenceContract(t *testing.T) {
	for _, required := range []string{"Repository roots nourish", "No independently verified product fruits", "/v1/federation/tree", "prefers-reduced-motion", "aria-live", "No cross-system relationship", "Evidence gap"} {
		if !strings.Contains(expeditionHTML, required) {
			t.Errorf("missing expedition contract %q", required)
		}
	}
	for _, forbidden := range []string{"localStorage", "POST /", "fetch(\"https://", "Math.random()", "XP awarded"} {
		if strings.Contains(expeditionHTML, forbidden) {
			t.Errorf("unexpected unsafe claim or side effect %q", forbidden)
		}
	}
}

func TestExpeditionRouteServesReadOnlyHTML(t *testing.T) {
	s := &Server{mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /expedition", func(w http.ResponseWriter, r *http.Request) { writeHTML(w, expeditionHTML) })
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, httptest.NewRequest("GET", "/expedition", nil))
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "Living Federation Expedition") {
		t.Fatalf("expedition route failed: %d", rr.Code)
	}
}

func TestExpeditionSourceComparisonContract(t *testing.T) {
	for _, required := range []string{"Source Comparison Theatre", "compare-a", "compare-b", "compare-clear", "renderComparison", "No relationship, dependency, product or achievement is inferred", "Contribution edges", "Not attested by this projection", "comparison-grid"} {
		if !strings.Contains(expeditionHTML, required) {
			t.Errorf("comparison theatre missing %q", required)
		}
	}
	for _, forbidden := range []string{"fetch(\"https://", "localStorage", "XP awarded"} {
		if strings.Contains(expeditionHTML, forbidden) {
			t.Errorf("comparison theatre has forbidden side effect %q", forbidden)
		}
	}
}

func TestExpeditionEvidenceJourneyContract(t *testing.T) {
	for _, required := range []string{"Evidence Journey", "journey-steps", "journey-focus", "journey-back", "journey-next", "journey-reset", "renderJourney", "setJourneyStep", "Exploration position, not verified completion", "No finished product fruit can be certified", "No edge from "} {
		if !strings.Contains(expeditionHTML, required) {
			t.Errorf("evidence journey missing %q", required)
		}
	}
	for _, forbidden := range []string{"XP awarded", "localStorage", "fetch(\"https://"} {
		if strings.Contains(expeditionHTML, forbidden) {
			t.Errorf("evidence journey includes forbidden behavior %q", forbidden)
		}
	}
}

func TestExpeditionEvidencePassportContract(t *testing.T) {
	for _, required := range []string{
		"Evidence Passport", "passport-track", "passport-focus", "passport-reset",
		"passportStages", "renderPassport", "setPassportStage", "Inspect root in observatory",
		"No verified outcome, contribution edge, quest completion or XP is inferred",
		"No repository root selected", "Evidence gap", "aria-pressed",
		"prefers-reduced-motion", "passport-grid",
	} {
		if !strings.Contains(expeditionHTML, required) {
			t.Errorf("evidence passport missing %q", required)
		}
	}
	for _, forbidden := range []string{"localStorage", "Math.random()", "XP awarded", "fetch(\\\"https://"} {
		if strings.Contains(expeditionHTML, forbidden) {
			t.Errorf("passport contains unsafe behavior %q", forbidden)
		}
	}
}

func TestExpeditionSharedContextContract(t *testing.T) {
	for _, required := range []string{
		"Expedition workspaces", "atlas-title", "lab-title", "compare-title",
		"passport-title", "ledger-title", "quests-title",
		"copy-context", "clear-context", "context-feedback",
		"requestedRoot", "rootLink", "syncContext", "navigator.clipboard.writeText",
		"state.items.find(n=>n.id===requestedRoot)", "history.replaceState",
		"Shared root was not found in current local observations.",
	} {
		if !strings.Contains(expeditionHTML, required) {
			t.Errorf("shared inspection context missing %q", required)
		}
	}
	for _, forbidden := range []string{"localStorage", "sessionStorage", "fetch(\"https://", "XP awarded"} {
		if strings.Contains(expeditionHTML, forbidden) {
			t.Errorf("shared context includes forbidden side effect %q", forbidden)
		}
	}
}

func TestExpeditionEvidenceQuestBoardContract(t *testing.T) {
	for _, required := range []string{
		"questDefinitions", "questFilter", "renderQuests", "quest-board",
		"Evidence expeditions", "Source inspectable", "Evidence gaps",
		"No completion or achievement issued", "No product fruit can be certified",
		"requires", "aria-pressed", "prefers-reduced-motion",
	} {
		if !strings.Contains(expeditionHTML, required) {
			t.Errorf("evidence quest board missing %q", required)
		}
	}
	for _, forbidden := range []string{"XP awarded", "Math.random()", "localStorage", "fetch(\"https://"} {
		if strings.Contains(expeditionHTML, forbidden) {
			t.Errorf("evidence quest board includes forbidden side effect %q", forbidden)
		}
	}
}

func TestExpeditionConstellationWorkbenchContract(t *testing.T) {
for _, required := range []string{
	"Federation Root Constellation Workbench", "constellation-workbench",
	"constellation-map", "constellation-results", "constellation-search",
	"renderConstellation", "chooseConstellationRoot",
	"No inferred products, dependencies, XP or achievements.",
	"No integration edges attested.", "aria-pressed", "aria-live",
	"state.items.slice()", "state.selected=root", "prefers-reduced-motion",
} {
	if !strings.Contains(expeditionHTML, required) {
	t.Errorf("constellation workbench missing %q", required)
	}
}
for _, forbidden := range []string{"Math.random()", "localStorage", "sessionStorage", "fetch(\"https://", "XP awarded"} {
	if strings.Contains(expeditionHTML, forbidden) {
	t.Errorf("constellation workbench unsafe content %q", forbidden)
	}
}
}
