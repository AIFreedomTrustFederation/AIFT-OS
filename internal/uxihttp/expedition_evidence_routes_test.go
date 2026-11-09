package uxihttp

import (
	"os"
	"strings"
	"testing"
)

// Route review must be bounded, source-linked, and isolated from local readiness.
func TestExpeditionEvidenceRouteObservatoryContract(t *testing.T) {
	b, err := os.ReadFile("expedition.html")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	required := []string{
		"function lineageRoutes(packet,from,to,limit=3)",
		"inspected++<20000", "queue.length>5000",
		"current.seen.has(edge.to)",
		"journeyAlternatives=lineageRoutes(portableGraph,start.id,goal.id,3)",
		"function showJourneyRoute(index)",
		"if(oldPacket!==packet)resetJourney(",
		"journeyDossier.textContent=dossier.join",
		"navigator.clipboard.writeText(journeyDossier.textContent)",
		"No achievements, product certification or readiness changes",
		"Evidence gap: no directed contribution route",
		"evidence.rel=\"noopener noreferrer\"",
	}
	for _, token := range required {
		if !strings.Contains(s, token) {
			t.Errorf("evidence route contract missing %q", token)
		}
	}
	forbidden := []string{
		"fetch(edge.evidenceUrl)", "localStorage.setItem(\"journey",
		"verifiedProducts.push(", "awardAchievement(",
	}
	for _, token := range forbidden {
		if strings.Contains(s, token) {
			t.Errorf("unsafe evidence route behavior %q", token)
		}
	}
}
