package uxi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInferenceUsesDefaultClientWhenNil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer server.Close()
	client := &OpenAICompatibleClient{Endpoint: server.URL, Model: "local"}
	answer, err := client.Complete(context.Background(), "system", nil)
	if err != nil {
		t.Fatal(err)
	}
	if answer != "ok" {
		t.Fatalf("answer=%q", answer)
	}
}

func TestInferenceRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", maxInferenceResponseBytes+1)))
	}))
	defer server.Close()
	client := &OpenAICompatibleClient{Endpoint: server.URL, Model: "local"}
	if _, err := client.Complete(context.Background(), "system", nil); err == nil {
		t.Fatal("expected size error")
	}
}
