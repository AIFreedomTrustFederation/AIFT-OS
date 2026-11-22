package uxi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInferenceEndpointLoopbackPolicy(t *testing.T) {
	for _, endpoint := range []string{
		"http://127.0.0.1:8080/v1",
		"http://[::1]:8080/v1",
	} {
		if err := validateLocalInferenceEndpoint(endpoint); err != nil {
			t.Errorf("valid endpoint %q rejected: %v", endpoint, err)
		}
	}
	for _, endpoint := range []string{
		"https://127.0.0.1:8080/v1",
		"http://example.com:8080/v1",
		"http://192.168.1.1:8080/v1",
		"http://localhost:8080/v1",
		"http://127.0.0.1/v1",
		"http://127.0.0.1:0/v1",
		"http://127.0.0.1:65536/v1",
		"http://127.0.0.1:notaport/v1",
		"http://user:secret@127.0.0.1:8080/v1",
		"http://127.0.0.1:8080/v1?key=secret",
		"http://127.0.0.1:8080/admin",
	} {
		if err := validateLocalInferenceEndpoint(endpoint); err == nil {
			t.Errorf("unsafe endpoint %q accepted", endpoint)
		}
	}
}

func TestInferenceDoesNotFollowRedirects(t *testing.T) {
	destinationCalled := false
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		destinationCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer destination.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
	}))
	defer source.Close()
	client := NewLocalInferenceClient(source.URL+"/v1", "local")
	_, err := client.Complete(context.Background(), "system", []ChatMessage{{Role: "user", Content: "hello"}})
	if err == nil || destinationCalled || !strings.Contains(err.Error(), "302") && !strings.Contains(err.Error(), "307") {
		t.Fatalf("redirect should be rejected, err=%v, destinationCalled=%v", err, destinationCalled)
	}
}

func TestInferenceRejectsRemoteEndpointBeforeRequest(t *testing.T) {
	client := NewLocalInferenceClient("http://example.com:8080/v1", "local")
	_, err := client.Complete(context.Background(), "system", nil)
	if err == nil || !strings.Contains(err.Error(), "loopback") {
		t.Fatalf("remote endpoint was not rejected: %v", err)
	}
}
