package uxi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxInferenceResponseBytes = 4 << 20

// ChatMessage is one OpenAI-compatible chat message.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Completer produces one local assistant response.
type Completer interface {
	Complete(ctx context.Context, system string, messages []ChatMessage) (string, error)
}

// OpenAICompatibleClient calls a local OpenAI-compatible model endpoint.
type OpenAICompatibleClient struct {
	Endpoint string
	Model    string
	Client   *http.Client
}

// NewLocalInferenceClient constructs a bounded local inference client.
func NewLocalInferenceClient(endpoint, model string) *OpenAICompatibleClient {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		endpoint = "http://127.0.0.1:8080/v1"
	}
	if model == "" {
		model = "local"
	}
	return &OpenAICompatibleClient{Endpoint: endpoint, Model: model, Client: &http.Client{Timeout: 90 * time.Second}}
}

// Complete requests a non-streaming local completion with a bounded response body.
func (c *OpenAICompatibleClient) Complete(ctx context.Context, system string, messages []ChatMessage) (string, error) {
	if c == nil {
		return "", errors.New("inference client is nil")
	}
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 90 * time.Second}
	}
	endpoint := strings.TrimRight(strings.TrimSpace(c.Endpoint), "/")
	if endpoint == "" {
		endpoint = "http://127.0.0.1:8080/v1"
	}
	model := strings.TrimSpace(c.Model)
	if model == "" {
		model = "local"
	}
	all := make([]ChatMessage, 0, len(messages)+1)
	all = append(all, ChatMessage{Role: "system", Content: system})
	all = append(all, messages...)
	data, err := json.Marshal(map[string]any{"model": model, "messages": all, "stream": false, "temperature": 0.2})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("local inference unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("local inference returned %s", resp.Status)
	}
	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxInferenceResponseBytes+1))
	if err != nil {
		return "", err
	}
	if len(payload) > maxInferenceResponseBytes {
		return "", errors.New("local inference response exceeded 4 MiB limit")
	}
	var decoded struct {
		Choices []struct {
			Message ChatMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return "", err
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return "", errors.New("local inference returned no content")
	}
	return strings.TrimSpace(decoded.Choices[0].Message.Content), nil
}
