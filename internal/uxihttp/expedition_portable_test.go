package uxihttp

import (
	"strings"
	"testing"
)

func TestPortableLineageImportBoundaries(t *testing.T) {
	checks := []string{
		"id=\"portable-lineage\"",
		"id=\"portable-input\"",
		"id=\"portable-status\"",
		"function inspectPortablePacket(packet)",
		"packet.schema!==portableSchema",
		"packet.nodes.length>2000",
		"packet.edges.length>5000",
		"ids.has(n.id)",
		"e.relation!==\"contributes-to\"",
		"portableUrl(e.evidenceUrl)",
		"renderPortable(null)",
		"independent verification not established",
		"rel=\"noopener noreferrer\"",
		"Local evidence unchanged.",
	}
	for _, needle := range checks {
		if !strings.Contains(expeditionHTML, needle) {
			t.Errorf("missing portable lineage safety contract %q", needle)
		}
	}
}
