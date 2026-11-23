package uxihttp

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/uxi"
)

func TestClassifyInferenceErrorSafeCategories(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"deadline", context.DeadlineExceeded, "inference_timeout"},
		{"cancelled", context.Canceled, "inference_canceled"},
		{"provider timeout", &uxi.InferenceFailure{Kind: "timeout"}, "inference_timeout"},
		{"connection", &uxi.InferenceFailure{Kind: "connection_error", Cause: errors.New("private path /secret")}, "inference_connection_error"},
		{"context rejection", &uxi.InferenceFailure{Kind: "request_rejected_possible_context_limit"}, "inference_possible_context_limit"},
		{"capacity", &uxi.InferenceFailure{Kind: "capacity_limited"}, "inference_capacity_limited"},
		{"unavailable", &uxi.InferenceFailure{Kind: "provider_unavailable"}, "inference_provider_unavailable"},
		{"wrapped", fmt.Errorf("wrapped: %w", &uxi.InferenceFailure{Kind: "connection_error"}), "inference_connection_error"},
		{"unknown", errors.New("private path /secret"), "inference_failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyInferenceError(tt.err)
			if got != tt.want {
				t.Fatalf("classifyInferenceError() = %q, want %q", got, tt.want)
			}
		})
	}
}
