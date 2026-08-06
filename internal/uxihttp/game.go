package uxihttp

import (
	_ "embed"
	"strings"
)

//go:embed tree-game.html
var treeGameBaseHTML string

//go:embed world-game.html
var worldGameBaseHTML string

//go:embed geometry-runtime.js
var geometryRuntimeJS string

var treeGameHTML = composeGameHTML(treeGameBaseHTML)
var worldGameHTML = composeGameHTML(worldGameBaseHTML)

func composeGameHTML(base string) string {
	const marker = "</body>"
	if !strings.Contains(base, marker) {
		panic("Federation game composition marker not found")
	}
	script := "  <script>\n" + geometryRuntimeJS + "\n  </script>\n"
	return strings.Replace(base, marker, script+marker, 1)
}
