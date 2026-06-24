package teamarchitect

import (
	"testing"

	"github.com/1424772/ForgeCrew/internal/scanner"
)

func TestSuggestCLITool(t *testing.T) {
	ta := New()
	p := &scanner.Profile{ProjectTypeGuess: "cli_tool"}
	tc := ta.Suggest(p)

	if tc.ProjectType != "cli_tool" {
		t.Errorf("ProjectType = %q, want cli_tool", tc.ProjectType)
	}
	if len(tc.RequiredAgents) != 6 {
		t.Errorf("expected 6 required agents for cli_tool, got %d", len(tc.RequiredAgents))
	}

	// Verify permission levels are conservative.
	for _, a := range tc.RequiredAgents {
		switch a.AgentID {
		case "pm", "architect", "repo_analyst", "code_reviewer":
			if a.PermissionLevel != "read_only" {
				t.Errorf("%s should be read_only, got %s", a.AgentID, a.PermissionLevel)
			}
		case "backend_engineer", "qa_engineer":
			if a.PermissionLevel != "patch_only" {
				t.Errorf("%s should be patch_only, got %s", a.AgentID, a.PermissionLevel)
			}
		}
	}
}

func TestSuggestBackendAPI(t *testing.T) {
	ta := New()
	p := &scanner.Profile{ProjectTypeGuess: "backend_api"}
	tc := ta.Suggest(p)

	if tc.ProjectType != "backend_api" {
		t.Errorf("ProjectType = %q, want backend_api", tc.ProjectType)
	}
	if len(tc.RequiredAgents) != 6 {
		t.Errorf("expected 6 required agents, got %d", len(tc.RequiredAgents))
	}
}

func TestSuggestFrontendApp(t *testing.T) {
	ta := New()
	p := &scanner.Profile{ProjectTypeGuess: "frontend_app"}
	tc := ta.Suggest(p)

	hasFrontend := false
	for _, a := range tc.RequiredAgents {
		if a.AgentID == "frontend_engineer" {
			hasFrontend = true
			break
		}
	}
	if !hasFrontend {
		t.Error("frontend_app should include frontend_engineer")
	}
}

func TestSuggestAgentApp(t *testing.T) {
	ta := New()
	p := &scanner.Profile{ProjectTypeGuess: "agent_app"}
	tc := ta.Suggest(p)

	if tc.ProjectType != "agent_app" {
		t.Errorf("ProjectType = %q, want agent_app", tc.ProjectType)
	}
	// agent_app should have memory_agent as optional.
	hasMemory := false
	for _, a := range tc.OptionalAgents {
		if a.AgentID == "memory_agent" {
			hasMemory = true
			break
		}
	}
	if !hasMemory {
		t.Error("agent_app should include memory_agent as optional")
	}
}

func TestSuggestGeneric(t *testing.T) {
	ta := New()
	p := &scanner.Profile{ProjectTypeGuess: "generic"}
	tc := ta.Suggest(p)

	if tc.ProjectType != "generic" {
		t.Errorf("ProjectType = %q, want generic", tc.ProjectType)
	}
	if len(tc.RequiredAgents) == 0 {
		t.Error("generic should have at least some required agents")
	}
}

func TestFormatYAML(t *testing.T) {
	tc := &TeamConfig{
		ProjectType: "cli_tool",
		RequiredAgents: []AgentAssignment{
			{AgentID: "pm", PermissionLevel: "read_only"},
		},
	}
	out, err := tc.FormatYAML()
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Error("YAML output should not be empty")
	}
}
