package uxi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// InferenceFailure classifies a local model failure without exposing provider payloads or credentials.
// Status is stable machine-readable evidence, not proof that a model is online.
type InferenceFailure struct {
	Kind       string
	StatusCode int
	Cause      error
}

func (e *InferenceFailure) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("local inference %s (HTTP %d)", e.Kind, e.StatusCode)
	}
	if e.Cause != nil {
		return fmt.Sprintf("local inference %s: %v", e.Kind, e.Cause)
	}
	return "local inference " + e.Kind
}
func (e *InferenceFailure) Unwrap() error { return e.Cause }

func classifyInferenceHTTP(code int) error {
	kind := "provider_http_error"
	switch {
	case code == http.StatusRequestTimeout || code == http.StatusGatewayTimeout:
		kind = "timeout"
	case code == http.StatusRequestEntityTooLarge || code == http.StatusUnprocessableEntity:
		kind = "request_rejected_possible_context_limit"
	case code == http.StatusTooManyRequests:
		kind = "capacity_limited"
	case code == http.StatusServiceUnavailable || code == http.StatusBadGateway:
		kind = "provider_unavailable"
	}
	return &InferenceFailure{Kind: kind, StatusCode: code}
}
func classifyInferenceTransport(err error) error {
	kind := "connection_error"
	var nerr net.Error
	switch {
	case errors.Is(err, context.Canceled):
		kind = "cancelled"
	case errors.Is(err, context.DeadlineExceeded):
		kind = "timeout"
	case errors.As(err, &nerr) && nerr.Timeout():
		kind = "timeout"
	}
	return &InferenceFailure{Kind: kind, Cause: err}
}
func inferenceFailureKind(err error) string {
	var failure *InferenceFailure
	if errors.As(err, &failure) {
		return failure.Kind
	}
	if err == nil {
		return ""
	}
	value := strings.ToLower(err.Error())
	if strings.Contains(value, "context") && (strings.Contains(value, "exceed") || strings.Contains(value, "length")) {
		return "possible_context_limit"
	}
	return "inference_error"
}
