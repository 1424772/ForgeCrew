// Package provider implements a minimal LLM provider client for
// OpenAI-compatible Chat Completions APIs.
//
// It reads model configuration from the models registry, resolves
// API keys from environment variables, and supports configurable
// base URLs with per-provider defaults.
package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/1424772/ForgeCrew/internal/models"
)

// Request is the input to a chat completion call.
type Request struct {
	System string // system message content
	Prompt string // user message content
	Model  string // model name override (empty = use definition default)
}

// Usage contains token usage statistics from the provider response.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Response is the result of a chat completion call.
type Response struct {
	Text       string
	Model      string
	Usage      *Usage // nil if provider didn't return usage data
	StatusCode int    // HTTP status code from the provider
}

// Client sends chat completion requests to an LLM provider.
type Client interface {
	Complete(ctx context.Context, req Request) (Response, error)
}

// DefaultBaseURLs maps provider names to their default API base URLs.
// These are used when models.yaml does not specify a base_url.
var DefaultBaseURLs = map[string]string{
	"openai":    "https://api.openai.com/v1",
	"deepseek":  "https://api.deepseek.com/v1",
	"zhipu":     "https://open.bigmodel.cn/api/paas/v4",
	"moonshot":  "https://api.moonshot.cn/v1",
	"dashscope": "https://dashscope.aliyuncs.com/compatible-mode/v1",
	"devstral":  "https://api.devstral.com/v1",
}

// Retry configuration defaults.
const (
	DefaultMaxRetries = 3
	DefaultBaseDelay  = 1 * time.Second
)

// openAIClient implements Client for OpenAI-compatible APIs.
type openAIClient struct {
	baseURL    string
	apiKey     string
	modelName  string
	httpClient *http.Client
	maxRetries int
	baseDelay  time.Duration
}

// NewClient creates a new provider Client from a model definition.
// The API key is read from the environment variable specified by def.APIKeyEnv.
// base_url is resolved from def.BaseURL, falling back to DefaultBaseURLs.
func NewClient(def models.ModelDefinition) (Client, error) {
	apiKey := os.Getenv(def.APIKeyEnv)
	if apiKey == "" {
		return nil, fmt.Errorf("provider: required environment variable %s is not set", def.APIKeyEnv)
	}

	baseURL := def.BaseURL
	if baseURL == "" {
		var ok bool
		baseURL, ok = DefaultBaseURLs[def.Provider]
		if !ok {
			return nil, fmt.Errorf("provider: no default base URL for provider %q, configure base_url in models.yaml", def.Provider)
		}
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	modelName := def.Model

	return &openAIClient{
		baseURL:   baseURL,
		apiKey:    apiKey,
		modelName: modelName,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: DefaultMaxRetries,
		baseDelay:  DefaultBaseDelay,
	}, nil
}

// Complete sends a chat completion request and returns the response.
// It retries on transient failures (429, 5xx) with exponential backoff
// up to maxRetries times. Non-retryable errors (4xx except 429) fail
// immediately. Safety: error messages never include the API key.
func (c *openAIClient) Complete(ctx context.Context, req Request) (Response, error) {
	model := c.modelName
	if req.Model != "" {
		model = req.Model
	}

	body := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: req.System},
			{Role: "user", Content: req.Prompt},
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return Response{}, fmt.Errorf("provider: marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Min(
				float64(c.baseDelay)*math.Pow(2, float64(attempt-1)),
				float64(30*time.Second),
			))
			select {
			case <-ctx.Done():
				return Response{}, fmt.Errorf("provider: context cancelled during retry: %w", ctx.Err())
			case <-time.After(delay):
			}
		}

		resp, respErr := c.doRequest(ctx, bytes.NewReader(bodyBytes))

		// Network/timeout errors (no HTTP status) are always retryable.
		if respErr != nil && resp.StatusCode == 0 {
			lastErr = respErr
			continue
		}

		// Content errors at 2xx (bad JSON, empty body) are permanent — don't retry.
		if respErr != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, respErr
		}

		// Permanent client errors (4xx except 429) — don't retry.
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return resp, respErr
		}

		// Success (2xx, no error) — done.
		if respErr == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Retryable: 429 or 5xx (with or without body parse errors).
		lastErr = fmt.Errorf("provider: HTTP %d (attempt %d/%d)",
			resp.StatusCode, attempt+1, c.maxRetries+1)
		if respErr != nil {
			lastErr = respErr
		}
	}

	return Response{}, fmt.Errorf("provider: request failed after %d retries: %w",
		c.maxRetries+1, lastErr)
}

// doRequest executes a single HTTP request and parses the response.
func (c *openAIClient) doRequest(ctx context.Context, bodyReader io.Reader) (Response, error) {
	url := c.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return Response{}, fmt.Errorf("provider: create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("provider: http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("provider: read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{Text: resp.Status, StatusCode: resp.StatusCode}, fmt.Errorf("provider: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	statusCode := resp.StatusCode

	var cr chatResponse
	if err := json.Unmarshal(respBody, &cr); err != nil {
		return Response{StatusCode: statusCode}, fmt.Errorf("provider: parse response: %w", err)
	}

	if len(cr.Choices) == 0 {
		return Response{StatusCode: statusCode}, fmt.Errorf("provider: empty response (no choices)")
	}

	text := cr.Choices[0].Message.Content
	if text == "" {
		return Response{StatusCode: statusCode}, fmt.Errorf("provider: empty response text")
	}

	var usage *Usage
	if cr.Usage.TotalTokens > 0 {
		usage = &Usage{
			PromptTokens:     cr.Usage.PromptTokens,
			CompletionTokens: cr.Usage.CompletionTokens,
			TotalTokens:      cr.Usage.TotalTokens,
		}
	}

	return Response{
		Text:       text,
		Model:      cr.Model,
		Usage:      usage,
		StatusCode: statusCode,
	}, nil
}

// ── OpenAI-compatible JSON types (unexported) ──

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
	Usage   chatUsage    `json:"usage"`
}

type chatChoice struct {
	Index   int         `json:"index"`
	Message chatMessage `json:"message"`
}

type chatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
