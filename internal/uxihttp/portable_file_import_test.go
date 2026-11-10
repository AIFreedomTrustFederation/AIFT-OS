package uxihttp

import (
	"os"
	"strings"
	"testing"
)

func TestPortableLineageFileImportContract(t *testing.T) {
	data, err := os.ReadFile("expedition.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	for _, marker := range []string{
		`id="portable-file" type="file" accept=".json,application/json"`,
		`$("portable-file").addEventListener("change",async event=>{`,
		`if(file.size>500000)`,
		`if(raw.length>500000)`,
		`$("portable-input").value=raw;`,
		`$("portable-inspect").click();`,
		`renderPortable(null);`,
		`event.target.value="";`,
		`$("portable-file").value="";`,
		`function inspectPortablePacket(packet){`,
		`e.relation!=="contributes-to"`,
		`renderPortableAtlas(packet);`,
		`renderPortableQuests(packet);`,
	} {
		if !strings.Contains(html, marker) {
			t.Errorf("missing portable import contract %q", marker)
		}
	}
}
