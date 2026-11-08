package uxihttp

import (
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestExpeditionBiologicalEvidenceContract(t *testing.T) {
 for _, required := range []string{"Repository roots nourish", "No independently verified product fruits", "/v1/federation/tree", "prefers-reduced-motion", "aria-live", "No cross-system relationship", "Evidence gap"} {
  if !strings.Contains(expeditionHTML, required) { t.Errorf("missing expedition contract %q", required) }
 }
 for _, forbidden := range []string{"localStorage", "POST /", "fetch(\"https://", "Math.random()", "XP awarded"} {
  if strings.Contains(expeditionHTML, forbidden) { t.Errorf("unexpected unsafe claim or side effect %q", forbidden) }
 }
}

func TestExpeditionRouteServesReadOnlyHTML(t *testing.T) {
 s := &Server{mux:http.NewServeMux()}
 s.mux.HandleFunc("GET /expedition",func(w http.ResponseWriter,r *http.Request){writeHTML(w,expeditionHTML)})
 rr:=httptest.NewRecorder()
 s.mux.ServeHTTP(rr,httptest.NewRequest("GET","/expedition",nil))
 if rr.Code!=http.StatusOK || !strings.Contains(rr.Body.String(),"Living Federation Expedition"){t.Fatalf("expedition route failed: %d",rr.Code)}
}
