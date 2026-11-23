package uxihttp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/uxi"
)

func TestInferenceFailureClassification(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"timeout", context.DeadlineExceeded, "inference_timeout"},
		{"wrapped timeout", errors.Join(errors.New("provider detail"), context.DeadlineExceeded), "inference_timeout"},
		{"canceled", context.Canceled, "inference_canceled"},
		{"provider failure", errors.New("private model endpoint and credentials"), "inference_failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyInferenceError(tc.err); got != tc.want {
				t.Fatalf("code=%q, want %q", got, tc.want)
			}
			payload, err := json.Marshal(messageJob{Error: "inference failed", ErrorCode: tc.want})
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(payload), "private model endpoint") || !strings.Contains(string(payload), tc.want) {
				t.Fatalf("unsafe or missing typed failure: %s", payload)
			}
		})
	}
}

func TestPublicJobHidesRetryMetadata(t *testing.T) {
	job := messageJob{IdempotencyKey: "private-retry-key", RequestHash: "private-hash"}
	public := publicMessageJob(job)
	encoded, err := json.Marshal(public)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "private-") {
		t.Fatal("retry metadata leaked to API response")
	}
	if job.IdempotencyKey == "" || job.RequestHash == "" {
		t.Fatal("public projection mutated durable record")
	}
}

func TestPublicJobRedactsNestedInferenceErrorWithoutMutatingDurableSession(t *testing.T) {
	private := "private provider URL /home/user/model.gguf token=secret"
	session := &uxi.Session{ID: "session", Turns: []uxi.Turn{
		{Role: "assistant", Content: "Local inference failed", Metadata: map[string]any{
			"inference_error":        private,
			"inference_failure_kind": "connection_error",
			"mode":                   "degraded",
		}},
	}}
	job := messageJob{Status: "completed", Result: session, IdempotencyKey: "secret-key"}
	projected := publicMessageJob(job)
	encoded, err := json.Marshal(projected)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), private) || strings.Contains(string(encoded), "inference_error") || strings.Contains(string(encoded), "secret-key") {
		t.Fatalf("private diagnostics leaked: %s", encoded)
	}
	if !strings.Contains(string(encoded), "connection_error") || !strings.Contains(string(encoded), "Local inference failed") {
		t.Fatalf("safe failure evidence missing: %s", encoded)
	}
	if session.Turns[0].Metadata["inference_error"] != private || job.IdempotencyKey != "secret-key" {
		t.Fatal("public projection mutated durable record")
	}
	projected.Result.Turns[0].Metadata["mode"] = "tampered"
	if session.Turns[0].Metadata["mode"] != "degraded" {
		t.Fatal("public metadata aliases durable session")
	}
}
