package uxihttp

import (
 "os"
 "strings"
 "testing"
)

func TestExpeditionEvidenceMatrixContract(t *testing.T) {
 page, err := os.ReadFile("expedition.html")
 if err != nil { t.Fatal(err) }
 html := string(page)
 required := []string{
  "Federation Evidence Matrix", "matrix-body", "matrix-search",
  "function renderMatrix()", "state.items.filter(n=>n.stage===0",
  "No downstream contribution edges or product outcomes attested",
  "aria-label\",observed?\"Locally observed repository root",
  "Independent evidence unavailable", "function render(){renderMatrix();",
 }
 for _, marker := range required {
  if !strings.Contains(html, marker) { t.Errorf("missing evidence matrix contract: %q", marker) }
 }
 if strings.Contains(html, "Verified product fruits: "+ "\"+roots.length") {
  t.Fatal("repository count must never be presented as product fruit count")
 }
}
