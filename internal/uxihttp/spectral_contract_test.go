package uxihttp

import (
 "os"
 "strings"
 "testing"
)

// Spectral presentation must not change provenance or local command behavior.
func TestLivingUniverseSpectralContract(t *testing.T) {
 data,err:=os.ReadFile("index.html")
 if err!=nil { t.Fatal(err) }
 html:=string(data)
 required:=[]string{
 "--aift-ruby:#ff5877","--aift-emerald:#43e5a7","--aift-azure:#64b4ff","--aift-aurum:#f3cd82",
 "--aift-ribbon:linear-gradient("," .universe-expedition", ".universe-quest[data-visited=\"true\"]",
 "@media(max-width:700px)","@media(prefers-reduced-motion:reduce)","@media(forced-colors:active)",
 }
 for _,marker:=range required { if !strings.Contains(html,marker) { t.Errorf("missing spectral contract %q",marker) } }
}
