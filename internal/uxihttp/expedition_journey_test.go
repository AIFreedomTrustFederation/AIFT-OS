package uxihttp

import (
 "strings"
 "testing"
)

func TestExpeditionEvidenceJourneyContract(t *testing.T) {
 checks := []string{
  "Lineage Journey &amp; Review Packet",
  "id=\"ej-rail\"",
  "id=\"ej-packet\"",
  "function evidencePacket()",
  "function renderEvidenceJourney()",
  "integrationEdges:[]",
  "verifiedProducts:[]",
  "not independently attested",
  "navigator.clipboard.writeText",
  "renderEvidenceJourney();renderConstellation()",
  "@media(max-width:430px)",
 }
 for _, needle := range checks {
  if !strings.Contains(expeditionHTML, needle) {
   t.Errorf("missing evidence journey contract %q", needle)
  }
 }
}
