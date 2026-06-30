package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1424772/ForgeCrew/internal/config"
)

// executeTaskCmd creates a fresh task command (not via rootCmd) to avoid
// state pollution from cobra's internal arg/flag handling across tests.
func executeTaskCmd(args ...string) (string, error) {
	taskDryRun = false
	taskExecute = false
	taskModel = ""
	cmd := taskCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestTaskDryRun(t *testing.T) {
	// Isolate in temp dir — no settings.yaml means default zh.
	chdirTempDir(t)
	out, err := executeTaskCmd("test-goal", "--dry-run")
	if err != nil {
		t.Fatalf("task --dry-run failed: %v", err)
	}
	// Default locale is zh (no settings.yaml in test temp dir).
	if !strings.Contains(out, "test-goal") {
		t.Error("output should contain the goal")
	}
	if !strings.Contains(out, "[演习]") {
		t.Error("output should mark [演习] (zh)")
	}
	// In zh: "第 1 轮:"
	if !strings.Contains(out, "第") || !strings.Contains(out, "轮") {
		t.Error("output should show Chinese iteration headers")
	}
	// zh output must NOT contain English placeholder text.
	if strings.Contains(out, "would execute") {
		t.Error("zh output should not contain 'would execute'")
	}
}

func TestTaskRequiresArg(t *testing.T) {
	chdirTempDir(t)
	_, err := executeTaskCmd()
	if err == nil {
		t.Error("task without goal should fail")
	}
}

func TestTaskHelp(t *testing.T) {
	chdirTempDir(t)
	out, err := executeTaskCmd("--help")
	if err != nil {
		t.Fatalf("task --help failed: %v", err)
	}
	if !strings.Contains(out, "--dry-run") {
		t.Error("help should document --dry-run flag")
	}
	if !strings.Contains(out, "goal") {
		t.Error("help should mention goal argument")
	}
}

func TestTaskAllStates(t *testing.T) {
	chdirTempDir(t)
	out, err := executeTaskCmd("verify-states", "--dry-run")
	if err != nil {
		t.Fatalf("task --dry-run failed: %v", err)
	}
	if out == "" {
		t.Fatal("empty output from task --dry-run")
	}
	// All executed states should appear (plan through commit_memory).
	expectedStates := []string{"plan", "retrieve", "act", "observe", "reflect", "improve", "review", "commit_memory"}
	for _, state := range expectedStates {
		if !strings.Contains(out, state) {
			t.Errorf("output should contain state %q", state)
		}
	}
}

func TestTaskExecuteMissingConfig(t *testing.T) {
	chdirTempDir(t)
	_, err := executeTaskCmd("test goal", "--execute", "--model", "gpt_5_5")
	if err == nil {
		t.Fatal("expected error when models.yaml is missing")
	}
	if !strings.Contains(err.Error(), "init") {
		t.Errorf("error should mention 'init', got: %v", err)
	}
}

func TestTaskExecuteNoModelFlag(t *testing.T) {
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	modelsYAML := `
models:
  gpt_5_5:
    provider: openai
    model: gpt-5.5
    api_key_env: TEST_KEY
    role: reasoning
    cost_tier: high
    supports_tools: true
    supports_vision: true
`
	os.WriteFile(filepath.Join(config.ConfigDir, config.ModelsYAML), []byte(modelsYAML), 0644)

	_, err := executeTaskCmd("test goal", "--execute")
	if err == nil {
		t.Fatal("expected error when --model is not specified")
	}
	if !strings.Contains(err.Error(), "model") {
		t.Errorf("error should mention --model, got: %v", err)
	}
}

func TestTaskExecuteUnknownModel(t *testing.T) {
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	modelsYAML := `
models:
  gpt_5_5:
    provider: openai
    model: gpt-5.5
    api_key_env: TEST_KEY
    role: reasoning
    cost_tier: high
    supports_tools: true
    supports_vision: true
`
	os.WriteFile(filepath.Join(config.ConfigDir, config.ModelsYAML), []byte(modelsYAML), 0644)

	_, err := executeTaskCmd("test goal", "--execute", "--model", "nonexistent_model")
	if err == nil {
		t.Fatal("expected error for unknown model")
	}
	if !strings.Contains(err.Error(), "未在") {
		t.Errorf("error should indicate model not found, got: %v", err)
	}
}

func TestTaskExecuteWithMockProvider(t *testing.T) {
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	// Start a mock server that returns a valid plan.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"model": "test-model",
			"choices": [{"index": 0, "message": {"role": "assistant", "content": "Step 1: Analyze\nStep 2: Implement\nStep 3: Test"}}],
			"usage": {"prompt_tokens": 5, "completion_tokens": 15, "total_tokens": 20}
		}`))
	}))
	t.Cleanup(srv.Close)

	// Write models.yaml pointing to the mock server.
	modelsYAML := `
models:
  test_model:
    provider: test
    model: test-model
    api_key_env: FORGECREW_TEST_KEY
    base_url: ` + srv.URL + `
    role: coding
    cost_tier: low
    supports_tools: false
    supports_vision: false
`
	os.WriteFile(filepath.Join(config.ConfigDir, config.ModelsYAML), []byte(modelsYAML), 0644)

	// Set the fake API key.
	os.Setenv("FORGECREW_TEST_KEY", "sk-test-fake")
	t.Cleanup(func() { os.Unsetenv("FORGECREW_TEST_KEY") })

	out, err := executeTaskCmd("analyze the project", "--execute", "--model", "test_model")
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if !strings.Contains(out, "Step 1: Analyze") {
		t.Errorf("output should contain plan text, got: %s", out)
	}
	if !strings.Contains(out, "LLM 生成的计划") {
		t.Errorf("output should contain plan header, got: %s", out)
	}
	// Must NOT contain dry-run orchestrator state names.
	if strings.Contains(out, "plan") && !strings.Contains(out, "Step 1") {
		t.Error("output should not contain orchestrator state names")
	}
}

func TestTaskDryRunStillProducesOutput(t *testing.T) {
	// Verify that running without --execute and without --dry-run still works.
	chdirTempDir(t)
	out, err := executeTaskCmd("some goal")
	if err != nil {
		t.Fatalf("task without --execute failed: %v", err)
	}
	if !strings.Contains(out, "[演习]") {
		t.Error("default (no --execute) should produce dry-run output with [演习]")
	}
}
