package uxihttp

import (
 "os"
 "strings"
 "testing"
)

func TestUniverseEvidenceExpeditionContract(t *testing.T) {
 data,err:=os.ReadFile("index.html")
 if err!=nil {t.Fatal(err)}
 html:=string(data)
 for _,marker:=range []string{
  `id="universeExpeditionTitle"`,
  `id="universeQuests"`,
  `id="universeQuestStatus"`,
  `const universeVisited=new Set()`,
  `function renderUniverseQuests()`,
  `Rootkeeper`,
  `Knowledge Gardener`,
  `universeVisited.add(target.id)`,
  `no product fruits or live health inferred.`,
 } {if !strings.Contains(html,marker) {t.Errorf("missing expedition marker %q",marker)}}
}
