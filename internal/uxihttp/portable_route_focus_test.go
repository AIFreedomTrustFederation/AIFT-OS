package uxihttp

import (
 "os"
 "strings"
 "testing"
)

// The directed evidence journey must remain connected to the same imported
// constellation, not become a separate, unverifiable visual representation.
func TestPortableRouteFocusContract(t *testing.T) {
 source, err := os.ReadFile("expedition.html")
 if err != nil { t.Fatal(err) }
 page := string(source)
 required := []string{
  "const routeEdges=new Set(",
  "routeEdges.has(e.from+",
  "renderPortableAtlas(portableGraph);",
  "Focus contributor ",
  "semanticButtons.forEach((button,i)=>button.setAttribute",
  "selected directed route highlighted, not independently verified",
  "journeyAlternatives=[];journeySelected=-1;journeyPacket=null;",
 }
 for _, marker := range required {
  if !strings.Contains(page, marker) { t.Errorf("route/constellation integration missing %q", marker) }
 }
 if strings.Contains(page, "verified achievement awarded from route") {
  t.Fatal("imported route cannot grant achievements")
 }
}
