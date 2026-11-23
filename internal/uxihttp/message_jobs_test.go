package uxihttp

import (
 "bytes"
 "context"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "testing"
 "time"

)

func TestAsyncMessageSurvivesBrowserDisconnect(t *testing.T) {
 server, store := newTestServer(t)
 session, err := store.CreateSession("disconnected-browser")
 if err != nil { t.Fatal(err) }
 req := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+session.ID+"/messages/async", bytes.NewBufferString(`{"content":"Reply with exactly AIFT_READY"}`))
 req.Header.Set("Content-Type", "application/json")
 ctx, cancel := context.WithCancel(req.Context())
 req = req.WithContext(ctx)
 recorder := httptest.NewRecorder()
 server.Handler().ServeHTTP(recorder, req)
 if recorder.Code != http.StatusAccepted { t.Fatalf("accepted code=%d body=%s", recorder.Code, recorder.Body.String()) }
 var job messageJob
 if err := json.Unmarshal(recorder.Body.Bytes(), &job); err != nil { t.Fatal(err) }
 cancel() // simulate browser closing its request
 deadline := time.Now().Add(5*time.Second)
 for time.Now().Before(deadline) {
  check := httptest.NewRecorder()
  server.Handler().ServeHTTP(check, httptest.NewRequest(http.MethodGet, "/v1/sessions/"+session.ID+"/message-jobs/"+job.ID, nil))
  if check.Code != http.StatusOK { t.Fatalf("poll code=%d", check.Code) }
  var current messageJob
  if err := json.Unmarshal(check.Body.Bytes(), &current); err != nil { t.Fatal(err) }
  if current.Status == "completed" {
   if current.Result == nil || len(current.Result.Turns) != 2 || current.Result.Turns[1].Content != "local answer" { t.Fatalf("result=%+v", current.Result) }
   return
  }
  if current.Status == "failed" { t.Fatalf("failed: %s", current.Error) }
  time.Sleep(10*time.Millisecond)
 }
 t.Fatal("detached inference did not complete")
}

func TestAsyncMessageRejectsInvalidSession(t *testing.T) {
 server, _ := newTestServer(t)
 req := httptest.NewRequest(http.MethodPost, "/v1/sessions/missing/messages/async", bytes.NewBufferString(`{"content":"hello"}`))
 req.Header.Set("Content-Type", "application/json")
 rr := httptest.NewRecorder()
 server.Handler().ServeHTTP(rr, req)
 if rr.Code != http.StatusNotFound { t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String()) }
}

