package uxihttp

import (
	"os"
	"strings"
	"testing"
)

func TestExpeditionSourceDossierContract(t *testing.T) {
	page, err := os.ReadFile("expedition.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(page)
	required := []string{
		"id=\"evidence-dossier\"",
		"id=\"dossier-rail\"",
		"id=\"dossier-source\"",
		"id=\"dossier-gaps\"",
		"id=\"dossier-copy\"",
		"function dossierReport()",
		"function renderDossier()",
		"renderDossier();renderMatrix();",
		"function inspect(item){state.selected=item;renderDossier();",
		"state.items.some(n=>n.id===root.id)",
		"navigator.clipboard.writeText(dossierReport())",
		"No repository is classified as a fruit.",
		"no verified finished product and complete lineage",
		"@media(max-width:420px)",
	}
	for _, marker := range required {
		if !strings.Contains(html, marker) {
			t.Errorf("missing evidence dossier contract: %q", marker)
		}
	}
	for _, forbidden := range []string{
		"fetch(\"https://",
		"localStorage.setItem",
		"Math.random()",
	} {
		if strings.Contains(html, forbidden) {
			t.Errorf("unexpected remote source, persistence or fabricated progress: %q", forbidden)
		}
	}
}
