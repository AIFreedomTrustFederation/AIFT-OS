package uxihttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDurableJobRecoveryAndIdempotency(t *testing.T) {
	server, store := newTestServer(t)
	session, err := store.CreateSession("durable")
	if err != nil {
		t.Fatal(err)
	}
	url := "/v1/sessions/" + session.ID + "/messages/async"
	post := func(s *Server, text, key string) (int, messageJob) {
		req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{"content":"`+text+`"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		var job messageJob
		_ = json.Unmarshal(rr.Body.Bytes(), &job)
		return rr.Code, job
	}
	code, first := post(server, "hello", "retry-1")
	if code != http.StatusAccepted {
		t.Fatalf("POST=%d", code)
	}
	code, duplicate := post(server, "hello", "retry-1")
	if code != http.StatusAccepted || duplicate.ID != first.ID {
		t.Fatalf("duplicate=%d %+v", code, duplicate)
	}
	code, _ = post(server, "different", "retry-1")
	if code != http.StatusConflict {
		t.Fatalf("conflict=%d", code)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rr := httptest.NewRecorder()
		server.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/sessions/"+session.ID+"/message-jobs/"+first.ID, nil))
		var job messageJob
		if err := json.Unmarshal(rr.Body.Bytes(), &job); err != nil {
			t.Fatal(err)
		}
		if job.Status == "completed" {
			break
		}
		if job.Status == "failed" {
			t.Fatalf("job failed: %s", job.Error)
		}
		time.Sleep(10 * time.Millisecond)
	}
	recovered, err := New(server.Engine)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	recovered.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/sessions/"+session.ID+"/message-jobs/"+first.ID, nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("recovered GET=%d", rr.Code)
	}
	var job messageJob
	if err := json.Unmarshal(rr.Body.Bytes(), &job); err != nil {
		t.Fatal(err)
	}
	if job.Status != "completed" || job.Result == nil {
		t.Fatalf("recovered job=%+v", job)
	}
	code, again := post(recovered, "hello", "retry-1")
	if code != http.StatusAccepted || again.ID != first.ID {
		t.Fatalf("restart duplicate=%d %+v", code, again)
	}
}
