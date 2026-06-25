// Package teamarchitect implements the Team Architect rules engine.
// It recommends team compositions based on project profiles.
package teamarchitect

import (
	"github.com/1424772/ForgeCrew/internal/scanner"
	"gopkg.in/yaml.v3"
)

// TeamConfig is the output of the Team Architect — a recommended team composition.
type TeamConfig struct {
	ProjectType    string            `yaml:"project_type" json:"project_type"`
	RequiredAgents []AgentAssignment `yaml:"required_agents" json:"required_agents"`
	OptionalAgents []AgentAssignment `yaml:"optional_agents" json:"optional_agents"`
}

// AgentAssignment assigns an agent role with a permission level for a project.
type AgentAssignment struct {
	AgentID         string `yaml:"agent_id" json:"agent_id"`
	PermissionLevel string `yaml:"permission_level" json:"permission_level"`
}

// TeamArchitect is the rules engine.
type TeamArchitect struct{}

// New creates a new TeamArchitect.
func New() *TeamArchitect {
	return &TeamArchitect{}
}

// Suggest generates a team configuration based on the project profile.
func (ta *TeamArchitect) Suggest(profile *scanner.Profile) *TeamConfig {
	required, optional := ta.selectAgents(profile)

	return &TeamConfig{
		ProjectType:    profile.ProjectTypeGuess,
		RequiredAgents: required,
		OptionalAgents: optional,
	}
}

func (ta *TeamArchitect) selectAgents(p *scanner.Profile) ([]AgentAssignment, []AgentAssignment) {
	switch p.ProjectTypeGuess {
	case "cli_tool":
		return ta.teamForCLITool(p)
	case "backend_api":
		return ta.teamForBackendAPI(p)
	case "frontend_app":
		return ta.teamForFrontendApp(p)
	case "fullstack_saas":
		return ta.teamForFullstackSaaS(p)
	case "agent_app":
		return ta.teamForAgentApp(p)
	default:
		return ta.teamForGeneric(p)
	}
}

func (ta *TeamArchitect) teamForCLITool(p *scanner.Profile) ([]AgentAssignment, []AgentAssignment) {
	required := []AgentAssignment{
		{AgentID: "pm", PermissionLevel: "read_only"},
		{AgentID: "architect", PermissionLevel: "read_only"},
		{AgentID: "repo_analyst", PermissionLevel: "read_only"},
		{AgentID: "backend_engineer", PermissionLevel: "patch_only"},
		{AgentID: "qa_engineer", PermissionLevel: "patch_only"},
		{AgentID: "code_reviewer", PermissionLevel: "read_only"},
	}
	optional := []AgentAssignment{
		{AgentID: "devops", PermissionLevel: "write_with_approval"},
	}
	return required, optional
}

func (ta *TeamArchitect) teamForBackendAPI(p *scanner.Profile) ([]AgentAssignment, []AgentAssignment) {
	required := []AgentAssignment{
		{AgentID: "pm", PermissionLevel: "read_only"},
		{AgentID: "architect", PermissionLevel: "read_only"},
		{AgentID: "repo_analyst", PermissionLevel: "read_only"},
		{AgentID: "backend_engineer", PermissionLevel: "patch_only"},
		{AgentID: "qa_engineer", PermissionLevel: "patch_only"},
		{AgentID: "code_reviewer", PermissionLevel: "read_only"},
	}
	optional := []AgentAssignment{
		{AgentID: "devops", PermissionLevel: "write_with_approval"},
	}
	return required, optional
}

func (ta *TeamArchitect) teamForFrontendApp(p *scanner.Profile) ([]AgentAssignment, []AgentAssignment) {
	required := []AgentAssignment{
		{AgentID: "pm", PermissionLevel: "read_only"},
		{AgentID: "architect", PermissionLevel: "read_only"},
		{AgentID: "repo_analyst", PermissionLevel: "read_only"},
		{AgentID: "frontend_engineer", PermissionLevel: "patch_only"},
		{AgentID: "qa_engineer", PermissionLevel: "patch_only"},
		{AgentID: "code_reviewer", PermissionLevel: "read_only"},
	}
	optional := []AgentAssignment{
		{AgentID: "backend_engineer", PermissionLevel: "patch_only"},
	}
	return required, optional
}

func (ta *TeamArchitect) teamForFullstackSaaS(p *scanner.Profile) ([]AgentAssignment, []AgentAssignment) {
	required := []AgentAssignment{
		{AgentID: "pm", PermissionLevel: "read_only"},
		{AgentID: "architect", PermissionLevel: "read_only"},
		{AgentID: "repo_analyst", PermissionLevel: "read_only"},
		{AgentID: "backend_engineer", PermissionLevel: "patch_only"},
		{AgentID: "frontend_engineer", PermissionLevel: "patch_only"},
		{AgentID: "qa_engineer", PermissionLevel: "patch_only"},
		{AgentID: "code_reviewer", PermissionLevel: "read_only"},
	}
	optional := []AgentAssignment{
		{AgentID: "devops", PermissionLevel: "write_with_approval"},
	}
	return required, optional
}

func (ta *TeamArchitect) teamForAgentApp(p *scanner.Profile) ([]AgentAssignment, []AgentAssignment) {
	required := []AgentAssignment{
		{AgentID: "pm", PermissionLevel: "read_only"},
		{AgentID: "architect", PermissionLevel: "read_only"},
		{AgentID: "repo_analyst", PermissionLevel: "read_only"},
		{AgentID: "backend_engineer", PermissionLevel: "patch_only"},
		{AgentID: "qa_engineer", PermissionLevel: "patch_only"},
		{AgentID: "code_reviewer", PermissionLevel: "read_only"},
	}
	optional := []AgentAssignment{
		{AgentID: "memory_agent", PermissionLevel: "read_only"},
		{AgentID: "devops", PermissionLevel: "write_with_approval"},
	}
	return required, optional
}

func (ta *TeamArchitect) teamForGeneric(p *scanner.Profile) ([]AgentAssignment, []AgentAssignment) {
	required := []AgentAssignment{
		{AgentID: "pm", PermissionLevel: "read_only"},
		{AgentID: "repo_analyst", PermissionLevel: "read_only"},
		{AgentID: "backend_engineer", PermissionLevel: "patch_only"},
		{AgentID: "qa_engineer", PermissionLevel: "patch_only"},
		{AgentID: "code_reviewer", PermissionLevel: "read_only"},
	}
	return required, nil
}

// FormatYAML formats the team config as YAML.
func (tc *TeamConfig) FormatYAML() (string, error) {
	out, err := yaml.Marshal(tc)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
