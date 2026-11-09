package uxihttp

import (
 "os"
 "strings"
 "testing"
)

func TestExpeditionDirectedLineageJourneyContract(t *testing.T) {
 body, err := os.ReadFile("expedition.html")
 if err != nil { t.Fatal(err) }
 page := string(body)
 required := []string{
  "function lineageShortestPath(packet,from,to)",
  "for(const edge of adjacency.get(id)||[])",
  "if(visited.has(edge.to))continue",
  "journeySource=journeyRoot.value",
  "renderLineageJourney(packet)",
  "start.kind!==\"root\"",
  "no directed contribution path",
  "evidence.href=edge.evidenceUrl",
  "evidence.rel=\"noopener noreferrer\"",
  "No independent product certification or achievement awarded",
 }
 for _, token := range required {
  if !strings.Contains(page, token) { t.Errorf("missing directed lineage journey contract %q", token) }
 }
}
