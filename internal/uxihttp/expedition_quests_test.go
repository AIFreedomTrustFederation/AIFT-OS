package uxihttp

import (
 "strings"
 "testing"
)

func TestExpeditionLineageQuestWorkbenchContract(t *testing.T) {
 required := []string{
  "id=\"quest-grid\"",
  "id=\"quest-status\"",
  "function renderPortableQuests(packet)",
  "renderPortableQuests(packet);",
  "0 independently verified achievements from this import.",
  "No attested outcomes",
  "No synthetic rewards",
  "m.kind===\"all\"",
  "aria-valuenow",
 }
 for _, token := range required {
  if !strings.Contains(expeditionHTML, token) {
   t.Errorf("missing evidence-derived quest contract: %s", token)
  }
 }
}
