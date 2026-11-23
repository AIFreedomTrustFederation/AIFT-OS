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

func TestReconcileCommittedExchangeAfterTerminalRecordLoss(t *testing.T) {
	server, store := newTestServer(t)
	session, err := store.CreateSession("reconcile")
	if err != nil {
		t.Fatal(err)
	}
	id := "0123456789abcdef0123456789abcdef"
	accepted := messageJob{ID: id, SessionID: session.ID, Status: "running", CreatedAt: time.Now().UTC()}
	if err := server.messageJobs.store.save(accepted); err != nil {
		t.Fatal(err)
	}
	committed, err := server.Engine.HandleMessageForJob(context.Background(), session.ID, "hello", id)
	if err != nil {
		t.Fatal(err)
	}
	if len(committed.Turns) != 2 {
		t.Fatalf("turns=%d", len(committed.Turns))
	}
	recovered, err := New(server.Engine)
	if err != nil {
		t.Fatal(err)
	}
	job := recovered.messageJobs.jobs[id]
	if job == nil || job.Status != "completed" || job.Result == nil {
		t.Fatalf("recovered=%+v", job)
	}
	latest, err := store.GetSession(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(latest.Turns) != 2 {
		t.Fatalf("unexpected replay, turns=%d", len(latest.Turns))
	}
}

func TestUncommittedInterruptedJobBlocksReplay(t *testing.T) {
	server, store := newTestServer(t)
	session, err := store.CreateSession("interrupted")
	if err != nil {
		t.Fatal(err)
	}
	id := "fedcba9876543210fedcba9876543210"
	if err := server.messageJobs.store.save(messageJob{ID: id, SessionID: session.ID, Status: "running", CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	recovered, err := New(server.Engine)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.messageJobs.jobs[id].Status != "interrupted" {
		t.Fatal("uncommitted job not interrupted")
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+session.ID+"/messages/async", bytes.NewBufferString(`{"content":"retry"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	recovered.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("replay=%d", rr.Code)
	}
}
