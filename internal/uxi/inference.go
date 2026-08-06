package uxi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Completer interface {
	Complete(ctx context.Context, system string, messages []ChatMessage) (string, error)
}

type OpenAICompatibleClient struct {
	Endpoint string
	Model    string
	Client   *http.Client
}

func NewLocalInferenceClient(endpoint, model string) *OpenAICompatibleClient {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		endpoint = "http://127.0.0.1:8080/v1"
	}
	if model == "" {
		model = "local"
	}
	return &OpenAICompatibleClient{
		Endpoint: endpoint,
		Model:    model,
		Client:   &http.Client{Timeout: 90 * time.Second},
	}
}

func (c *OpenAICompatibleClient) Complete(ctx context.Context, system string, messages []ChatMessage) (string, error) {
	if c == nil {
		return "", errors.New("inference client is nil")
	}
	all := make([]ChatMessage, 0, len(messages)+1)
	all = append(all, ChatMessage{Role: "system", Content: system})
	all = append(all, messages...)
	requestBody := map[string]any{
		"model":       c.Model,
		"messages":    all,
		"stream":      false,
		"temperature": 0.2,
	}
	data, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("local inference unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("local inference returned %s", resp.Status)
	}
	var decoded struct {
		Choices []struct {
			Message ChatMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", err
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return "", errors.New("local inference returned no content")
	}
	return strings.TrimSpace(decoded.Choices[0].Message.Content), nil
}
