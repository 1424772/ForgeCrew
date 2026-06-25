package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/1424772/ForgeCrew/internal/config"
)

// executeAgentsCmd creates a fresh agents command and runs it.
func executeAgentsCmd(args ...string) (string, error) {
	cmd := agentsCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestAgentsValidateAllGood(t *testing.T) {
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	// Write valid agents.yaml and models.yaml.
	agentsYAML := `
agents:
  test_agent:
    agent_id: test_agent
    name: Test Agent
    default_model: gpt_5_5
    fallback_models:
      - glm_5_2
    permission_level: read_only
    tools:
      - read_file
      - search_code
    modes:
      - test
`
	modelsYAML := `
models:
  gpt_5_5:
    provider: openai
    model: gpt-5.5
    api_key_env: OPENAI_API_KEY
    role: reasoning
    cost_tier: high
    supports_tools: true
    supports_vision: true
  glm_5_2:
    provider: zhipu
    model: glm-5.2
    api_key_env: ZHIPU_API_KEY
    role: coding
    cost_tier: medium
    supports_tools: true
    supports_vision: false
`
	os.WriteFile(filepath.Join(config.ConfigDir, config.AgentsYAML), []byte(agentsYAML), 0644)
	os.WriteFile(filepath.Join(config.ConfigDir, config.ModelsYAML), []byte(modelsYAML), 0644)

	out, err := executeAgentsCmd("validate")
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}

	// In zh locale, should show Chinese success messages.
	if out == "" {
		t.Error("validate output should not be empty")
	}
}

func TestAgentsValidateBadDefaultModel(t *testing.T) {
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	// Agent references a model that doesn't exist.
	agentsYAML := `
agents:
  bad_agent:
    agent_id: bad_agent
    name: Bad Agent
    default_model: nonexistent_model_xyz
    permission_level: read_only
    tools:
      - read_file
`
	modelsYAML := `
models:
  gpt_5_5:
    provider: openai
    model: gpt-5.5
    api_key_env: OPENAI_API_KEY
    role: reasoning
    cost_tier: high
    supports_tools: true
    supports_vision: true
`
	os.WriteFile(filepath.Join(config.ConfigDir, config.AgentsYAML), []byte(agentsYAML), 0644)
	os.WriteFile(filepath.Join(config.ConfigDir, config.ModelsYAML), []byte(modelsYAML), 0644)

	_, err := executeAgentsCmd("validate")
	if err == nil {
		t.Fatal("validate should fail with bad default_model")
	}
}

func TestAgentsValidateBadFallbackModel(t *testing.T) {
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	agentsYAML := `
agents:
  bad_fb:
    agent_id: bad_fb
    name: Bad Fallback
    default_model: gpt_5_5
    fallback_models:
      - nonexistent_fallback_xyz
    permission_level: read_only
    tools:
      - read_file
`
	modelsYAML := `
models:
  gpt_5_5:
    provider: openai
    model: gpt-5.5
    api_key_env: OPENAI_API_KEY
    role: reasoning
    cost_tier: high
    supports_tools: true
    supports_vision: true
`
	os.WriteFile(filepath.Join(config.ConfigDir, config.AgentsYAML), []byte(agentsYAML), 0644)
	os.WriteFile(filepath.Join(config.ConfigDir, config.ModelsYAML), []byte(modelsYAML), 0644)

	_, err := executeAgentsCmd("validate")
	if err == nil {
		t.Fatal("validate should fail with bad fallback_model")
	}
}

func TestAgentsValidateBadTool(t *testing.T) {
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	agentsYAML := `
agents:
  bad_tool_agent:
    agent_id: bad_tool_agent
    name: Bad Tool
    default_model: gpt_5_5
    permission_level: read_only
    tools:
      - evil_hack_tool
`
	modelsYAML := `
models:
  gpt_5_5:
    provider: openai
    model: gpt-5.5
    api_key_env: OPENAI_API_KEY
    role: reasoning
    cost_tier: high
    supports_tools: true
    supports_vision: true
`
	os.WriteFile(filepath.Join(config.ConfigDir, config.AgentsYAML), []byte(agentsYAML), 0644)
	os.WriteFile(filepath.Join(config.ConfigDir, config.ModelsYAML), []byte(modelsYAML), 0644)

	_, err := executeAgentsCmd("validate")
	if err == nil {
		t.Fatal("validate should fail with unknown tool")
	}
}

func TestAgentsValidateMissingFiles(t *testing.T) {
	chdirTempDir(t)

	// No files at all.
	_, err := executeAgentsCmd("validate")
	if err == nil {
		t.Error("validate should fail when config files are missing")
	}
}
