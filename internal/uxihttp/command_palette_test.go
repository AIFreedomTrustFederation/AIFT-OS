package uxihttp

import (
 "os"
 "strings"
 "testing"
)

// The command palette must remain local, read-only and keyboard accessible.
func TestLocalCommandPaletteContract(t *testing.T) {
 html, err := os.ReadFile("index.html")
 if err != nil { t.Fatal(err) }
 required := []string{
  `id="paletteTrigger"`,
  `role="dialog" aria-modal="true"`,
  `id="paletteQuery" type="search"`,
  `role="listbox"`,
  `function paletteDestinations()`,
  `...state.sessions.map(`,
  `...state.repos.map(`,
  `function openPalette()`,
  `function closePalette()`,
  `function activatePalette(index)`,
  `event.key==="ArrowDown"`,
  `event.key==="ArrowUp"`,
  `event.key==="Escape"`,
  `event.key==="Tab"`,
  `event.key.toLowerCase()==="k"`,
  `textContent=text`,
  `No matching local destinations.`,
 }
 for _, marker := range required {
  if !strings.Contains(string(html), marker) { t.Errorf("missing palette contract %q", marker) }
 }
}
