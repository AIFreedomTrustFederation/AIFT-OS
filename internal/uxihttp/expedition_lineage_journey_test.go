package uxihttp

import (
	"os"
	"strings"
	"testing"
)

func TestExpeditionDirectedLineageJourneyContract(t *testing.T) {
	body, err := os.ReadFile("expedition.html")
	if err != nil {
		t.Fatal(err)
	}
	page := string(body)
	required := []string{
		"function lineageShortestPath(packet,from,to)",
		"function lineageRoutes(packet,from,to,limit=3)",
		"inspected++<20000",
		"queue.length>5000",
		"renderJourneyComparison(portableGraph,journeyAlternatives)",
		"for(const edge of adj.get(current.id)||[])",
		"if(current.seen.has(edge.to))continue",
		"journeySource=journeyRoot.value",
		"renderLineageJourney(packet)",
		"start.kind!==\"root\"",
		"Evidence gap: no directed contribution route",
		"evidence.href=edge.evidenceUrl",
		"evidence.rel=\"noopener noreferrer\"",
		"No independent product certification or achievement awarded",
	}
	for _, token := range required {
		if !strings.Contains(page, token) {
			t.Errorf("missing directed lineage journey contract %q", token)
		}
	}
}
