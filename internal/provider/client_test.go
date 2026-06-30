package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/1424772/ForgeCrew/internal/models"
)

// ── Helpers ──

func testModelDef(provider, model, apiKeyEnv, baseURL string) models.ModelDefinition {
	return models.ModelDefinition{
		Provider:  provider,
		Model:     model,
		APIKeyEnv: apiKeyEnv,
		BaseURL:   baseURL,
	}
}

func startTestServer(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv.URL
}

// ── Tests ──

func TestCompleteSuccess(t *testing.T) {
	srvURL := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "chat-123",
			"model": "test-model",
			"choices": [{"index": 0, "message": {"role": "assistant", "content": "Here is the plan."}}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}
		}`))
	})

	os.Setenv("TEST_KEY", "sk-fake")
	t.Cleanup(func() { os.Unsetenv("TEST_KEY") })

	def := testModelDef("test", "test-model", "TEST_KEY", srvURL)
	client, err := NewClient(def)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	resp, err := client.Complete(context.Background(), Request{
		System: "You are helpful.",
		Prompt: "Plan this task.",
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}

	if resp.Text != "Here is the plan." {
		t.Errorf("Text = %q, want %q", resp.Text, "Here is the plan.")
	}
	if resp.Model != "test-model" {
		t.Errorf("Model = %q, want %q", resp.Model, "test-model")
	}
	if resp.Usage == nil {
		t.Error("Usage should not be nil")
	} else if resp.Usage.TotalTokens != 30 {
		t.Errorf("TotalTokens = %d, want 30", resp.Usage.TotalTokens)
	}
}

func TestCompleteSuccessNoUsage(t *testing.T) {
	// Usage field with 0 total_tokens should result in nil Usage in response.
	srvURL := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"model": "test-model",
			"choices": [{"index": 0, "message": {"role": "assistant", "content": "OK"}}],
			"usage": {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}
		}`))
	})

	os.Setenv("TEST_KEY", "sk-fake")
	t.Cleanup(func() { os.Unsetenv("TEST_KEY") })

	def := testModelDef("test", "test-model", "TEST_KEY", srvURL)
	client, _ := NewClient(def)
	resp, err := client.Complete(context.Background(), Request{
		System: "sys",
		Prompt: "do",
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Usage != nil {
		t.Errorf("Usage should be nil when total_tokens is 0, got %+v", resp.Usage)
	}
}

func TestNewClientMissingAPIKey(t *testing.T) {
	os.Unsetenv("MISSING_KEY_VAR")
	def := testModelDef("openai", "gpt-5.5", "MISSING_KEY_VAR", "https://api.example.com/v1")
	_, err := NewClient(def)
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "not set") {
		t.Errorf("error should mention 'not set', got: %v", err)
	}
	// Error should NOT contain a fake key value.
	if strings.Contains(err.Error(), "sk-") {
		t.Error("error should not contain any key value")
	}
}

func TestNewClientUnknownProvider(t *testing.T) {
	os.Setenv("SOME_KEY", "sk-fake")
	t.Cleanup(func() { os.Unsetenv("SOME_KEY") })

	def := testModelDef("unknown_provider_xyz", "some-model", "SOME_KEY", "")
	_, err := NewClient(def)
	if err == nil {
		t.Fatal("expected error for unknown provider without base_url")
	}
	if !strings.Contains(err.Error(), "no default base URL") {
		t.Errorf("error should mention 'no default base URL', got: %v", err)
	}
}

func TestCompleteHTTP500(t *testing.T) {
	srvURL := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	})

	os.Setenv("TEST_KEY", "sk-fake")
	t.Cleanup(func() { os.Unsetenv("TEST_KEY") })

	def := testModelDef("test", "test-model", "TEST_KEY", srvURL)
	client, _ := NewClient(def)
	_, err := client.Complete(context.Background(), Request{
		System: "sys",
		Prompt: "do",
	})
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should contain status code 500, got: %v", err)
	}
}

func TestCompleteEmptyChoices(t *testing.T) {
	srvURL := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"model": "x", "choices": []}`))
	})

	os.Setenv("TEST_KEY", "sk-fake")
	t.Cleanup(func() { os.Unsetenv("TEST_KEY") })

	def := testModelDef("test", "test-model", "TEST_KEY", srvURL)
	client, _ := NewClient(def)
	_, err := client.Complete(context.Background(), Request{
		System: "sys",
		Prompt: "do",
	})
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error should mention 'empty', got: %v", err)
	}
}

func TestCompleteEmptyContent(t *testing.T) {
	srvURL := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"model": "x", "choices": [{"index": 0, "message": {"role": "assistant", "content": ""}}]}`))
	})

	os.Setenv("TEST_KEY", "sk-fake")
	t.Cleanup(func() { os.Unsetenv("TEST_KEY") })

	def := testModelDef("test", "test-model", "TEST_KEY", srvURL)
	client, _ := NewClient(def)
	_, err := client.Complete(context.Background(), Request{
		System: "sys",
		Prompt: "do",
	})
	if err == nil {
		t.Fatal("expected error for empty content")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error should mention 'empty', got: %v", err)
	}
}

func TestCompleteInvalidJSON(t *testing.T) {
	srvURL := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not json at all {{{`))
	})

	os.Setenv("TEST_KEY", "sk-fake")
	t.Cleanup(func() { os.Unsetenv("TEST_KEY") })

	def := testModelDef("test", "test-model", "TEST_KEY", srvURL)
	client, _ := NewClient(def)
	_, err := client.Complete(context.Background(), Request{
		System: "sys",
		Prompt: "do",
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("error should mention 'parse', got: %v", err)
	}
}

func TestCompleteTimeout(t *testing.T) {
	srvURL := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than the client's context timeout to trigger a deadline error.
		time.Sleep(2 * time.Second)
	})

	os.Setenv("TEST_KEY", "sk-fake")
	t.Cleanup(func() { os.Unsetenv("TEST_KEY") })

	def := testModelDef("test", "test-model", "TEST_KEY", srvURL)
	client, err := NewClient(def)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// Create a very short timeout context.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = client.Complete(ctx, Request{
		System: "sys",
		Prompt: "do",
	})
	if err == nil {
		t.Fatal("expected timeout/context error")
	}
}

func TestDefaultBaseURLResolution(t *testing.T) {
	// Verify that known providers get their default base URLs.
	os.Setenv("OPENAI_API_KEY", "sk-fake")
	t.Cleanup(func() { os.Unsetenv("OPENAI_API_KEY") })

	def := models.ModelDefinition{
		Provider:  "openai",
		Model:     "gpt-5.5",
		APIKeyEnv: "OPENAI_API_KEY",
		// BaseURL intentionally empty — should use default.
	}
	client, err := NewClient(def)
	if err != nil {
		t.Fatalf("NewClient should resolve default base URL: %v", err)
	}
	if client == nil {
		t.Fatal("client should not be nil")
	}
}

func TestExplicitBaseURL(t *testing.T) {
	os.Setenv("TEST_KEY", "sk-fake")
	t.Cleanup(func() { os.Unsetenv("TEST_KEY") })

	customURL := "https://custom-proxy.example.com/v1"
	def := models.ModelDefinition{
		Provider:  "openai",
		Model:     "gpt-5.5",
		APIKeyEnv: "TEST_KEY",
		BaseURL:   customURL,
	}
	client, err := NewClient(def)
	if err != nil {
		t.Fatalf("NewClient with explicit base URL: %v", err)
	}
	if client == nil {
		t.Fatal("client should not be nil")
	}
}
