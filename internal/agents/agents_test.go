package agents

import (
	"os"
	"path/filepath"
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
