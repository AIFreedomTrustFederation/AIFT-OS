package uxi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"strconv"
	"time"
)

const (
	maxInferenceResponseBytes = 4 << 20
	defaultInferenceEndpoint  = "http://127.0.0.1:8080/v1"
	defaultInferenceModel     = "local"
	defaultInferenceTimeout   = 90 * time.Second
)

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
	endpoint, model = normalizeInferenceConfig(endpoint, model)
	return &OpenAICompatibleClient{Endpoint: endpoint, Model: model, Client: &http.Client{Timeout: defaultInferenceTimeout}}
}

func normalizeInferenceConfig(endpoint, model string) (string, string) {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		endpoint = defaultInferenceEndpoint
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultInferenceModel
	}
	return endpoint, model
}

// Complete requests a non-streaming local completion with a bounded response body.
func (c *OpenAICompatibleClient) Complete(ctx context.Context, system string, messages []ChatMessage) (string, error) {
	if c == nil {
		return "", errors.New("inference client is nil")
	}
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: defaultInferenceTimeout}
	}
	endpoint, model := normalizeInferenceConfig(c.Endpoint, c.Model)
	if err := validateLocalInferenceEndpoint(endpoint); err != nil {
		return "", err
	}
	// Never follow redirects, even when a custom HTTP client is injected.
	boundedClient := *client
	boundedClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
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
	resp, err := boundedClient.Do(req)
	if err != nil {
		return "", classifyInferenceTransport(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", classifyInferenceHTTP(resp.StatusCode)
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

// validateLocalInferenceEndpoint prevents a configured provider from becoming
// an outbound network or privileged local-socket request. Only numeric loopback
// HTTP hosts are admitted; no credentials, fragments, or query strings.
func validateLocalInferenceEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil || u == nil || u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("local inference endpoint must be a plain loopback HTTP URL")
	}
	ip := net.ParseIP(u.Hostname())
	if ip == nil || !ip.IsLoopback() || u.Port() == "" {
		return errors.New("local inference endpoint must use a numeric loopback address and explicit port")
	}
	port, portErr := strconv.Atoi(u.Port())
	if portErr != nil || port < 1 || port > 65535 {
		return errors.New("local inference endpoint port must be between 1 and 65535")
	}
	if u.Path != "" && u.Path != "/v1" {
		return errors.New("local inference endpoint path must be /v1")
	}
	return nil
}
