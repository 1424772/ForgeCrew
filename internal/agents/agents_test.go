package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if len(r.agents) != 0 {
		t.Error("new registry should be empty")
	}
}

func TestLoadAgents(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "agents.yaml")

	yamlContent := `
agents:
  test_agent:
    agent_id: test_agent
    name: Test Agent
    description: A test agent
    default_model: gpt_5_5
    fallback_models:
      - glm_5_2
    permission_level: read_only
    require_approval: false
    tools:
      - read_file
    modes:
      - test
`
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewRegistry()
	if err := r.Load(path); err != nil {
		t.Fatal(err)
	}

	if len(r.agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(r.agents))
	}

	a, err := r.Get("test_agent")
	if err != nil {
		t.Fatal(err)
	}
	if a.Name != "Test Agent" {
		t.Errorf("name = %q, want %q", a.Name, "Test Agent")
	}
	if a.DefaultModel != "gpt_5_5" {
		t.Errorf("default_model = %q, want %q", a.DefaultModel, "gpt_5_5")
	}
}

func TestLoadMissingFile(t *testing.T) {
	r := NewRegistry()
	err := r.Load("/nonexistent/path/agents.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestGetMissingAgent(t *testing.T) {
	r := NewRegistry()
	_, err := r.Get("nonexistent")
	if err == nil {
		t.Error("expected error for missing agent")
	}
}

func TestValidateAgent(t *testing.T) {
	tests := []struct {
		name    string
		agent   AgentDefinition
		wantErr bool
	}{
		{
			name: "valid",
			agent: AgentDefinition{
				AgentID: "test", Name: "Test", DefaultModel: "gpt_5_5",
				PermissionLevel: "read_only",
			},
			wantErr: false,
		},
		{
			name: "missing agent_id",
			agent: AgentDefinition{
				Name: "Test", DefaultModel: "gpt_5_5",
				PermissionLevel: "read_only",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			agent: AgentDefinition{
				AgentID: "test", DefaultModel: "gpt_5_5",
				PermissionLevel: "read_only",
			},
			wantErr: true,
		},
		{
			name: "missing default_model",
			agent: AgentDefinition{
				AgentID: "test", Name: "Test",
				PermissionLevel: "read_only",
			},
			wantErr: true,
		},
		{
			name: "invalid permission",
			agent: AgentDefinition{
				AgentID: "test", Name: "Test", DefaultModel: "gpt_5_5",
				PermissionLevel: "admin",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAgent(tt.agent)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAgent() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestList(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "agents.yaml")

	yamlContent := `
agents:
  agent_a:
    agent_id: agent_a
    name: Agent A
    default_model: gpt_5_5
    permission_level: read_only
  agent_b:
    agent_id: agent_b
    name: Agent B
    default_model: glm_5_2
    permission_level: patch_only
`
	os.WriteFile(path, []byte(yamlContent), 0644)

	r := NewRegistry()
	r.Load(path)

	list := r.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(list))
	}
}

// stubModelLookup implements ModelLookup for tests.
type stubModelLookup struct {
	ids map[string]bool
}

func (s stubModelLookup) Exists(id string) bool { return s.ids[id] }

func TestValidateModelRefsAllValid(t *testing.T) {
	r := NewRegistry()
	r.agents["a"] = AgentDefinition{
		AgentID: "a", DefaultModel: "gpt_5_5",
		FallbackModels: []string{"glm_5_2"},
	}
	models := stubModelLookup{ids: map[string]bool{"gpt_5_5": true, "glm_5_2": true}}

	if err := r.ValidateModelRefs(models); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateModelRefsMissingDefault(t *testing.T) {
	r := NewRegistry()
	r.agents["a"] = AgentDefinition{
		AgentID: "a", DefaultModel: "missing_model",
	}
	models := stubModelLookup{ids: map[string]bool{}}

	err := r.ValidateModelRefs(models)
	if err == nil {
		t.Fatal("expected error for missing default_model")
	}
	if !strings.Contains(err.Error(), "missing_model") {
		t.Errorf("error should mention missing model, got: %v", err)
	}
}

func TestValidateModelRefsMissingFallback(t *testing.T) {
	r := NewRegistry()
	r.agents["a"] = AgentDefinition{
		AgentID: "a", DefaultModel: "gpt_5_5",
		FallbackModels: []string{"bad_fallback"},
	}
	models := stubModelLookup{ids: map[string]bool{"gpt_5_5": true}}

	err := r.ValidateModelRefs(models)
	if err == nil {
		t.Fatal("expected error for missing fallback_model")
	}
	if !strings.Contains(err.Error(), "bad_fallback") {
		t.Errorf("error should mention bad_fallback, got: %v", err)
	}
}

func TestValidateModelRefsEmptyRegistry(t *testing.T) {
	r := NewRegistry()
	models := stubModelLookup{ids: map[string]bool{}}
	if err := r.ValidateModelRefs(models); err != nil {
		t.Errorf("empty registry should produce no errors, got: %v", err)
	}
}

func TestValidateToolsAllValid(t *testing.T) {
	r := NewRegistry()
	r.agents["a"] = AgentDefinition{
		AgentID: "a", Tools: []string{"read_file", "search_code"},
	}
	known := map[string]bool{"ReadFile": true, "SearchCode": true}

	if err := r.ValidateTools(known); err != nil {
		t.Errorf("expected no error for valid tools, got: %v", err)
	}
}

func TestValidateToolsUnknown(t *testing.T) {
	r := NewRegistry()
	r.agents["a"] = AgentDefinition{
		AgentID: "a", Tools: []string{"evil_tool"},
	}
	known := map[string]bool{"ReadFile": true}

	err := r.ValidateTools(known)
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
	if !strings.Contains(err.Error(), "evil_tool") {
		t.Errorf("error should mention evil_tool, got: %v", err)
	}
}

func TestValidateToolsNilKnownNames(t *testing.T) {
	r := NewRegistry()
	r.agents["a"] = AgentDefinition{
		AgentID: "a", Tools: []string{"anything"},
	}
	if err := r.ValidateTools(nil); err != nil {
		t.Errorf("nil knownNames should skip validation, got: %v", err)
	}
}

func TestValidateToolsEmptyTools(t *testing.T) {
	r := NewRegistry()
	r.agents["a"] = AgentDefinition{
		AgentID: "a", Tools: []string{},
	}
	known := map[string]bool{"ReadFile": true}
	if err := r.ValidateTools(known); err != nil {
		t.Errorf("empty tools should produce no errors, got: %v", err)
	}
}

func TestValidateToolsCaseInsensitive(t *testing.T) {
	r := NewRegistry()
	r.agents["a"] = AgentDefinition{
		AgentID: "a", Tools: []string{"read_file", "SEARCH_CODE", "ReadFile"},
	}
	// ACI actions use PascalCase; tools in YAML use snake_case.
	known := map[string]bool{"ReadFile": true, "SearchCode": true}

	if err := r.ValidateTools(known); err != nil {
		t.Errorf("case-insensitive match should pass, got: %v", err)
	}
}
