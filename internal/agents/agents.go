// Package agents provides the Agent Registry — definitions, loading, and querying.
package agents

import (
	"fmt"
	"os"

	"github.com/1424772/ForgeCrew/internal/config"
	"gopkg.in/yaml.v3"
)

// AgentDefinition describes a single agent's configuration.
type AgentDefinition struct {
	AgentID         string   `yaml:"agent_id" json:"agent_id"`
	Name            string   `yaml:"name" json:"name"`
	Description     string   `yaml:"description" json:"description"`
	DefaultModel    string   `yaml:"default_model" json:"default_model"`
	FallbackModels  []string `yaml:"fallback_models" json:"fallback_models"`
	PermissionLevel string   `yaml:"permission_level" json:"permission_level"`
	RequireApproval bool     `yaml:"require_approval" json:"require_approval"`
	Tools           []string `yaml:"tools" json:"tools"`
	Modes           []string `yaml:"modes" json:"modes"`
}

// agentsFile is the YAML file structure for agents.
type agentsFile struct {
	Agents map[string]AgentDefinition `yaml:"agents"`
}

// Registry stores and manages agent definitions.
type Registry struct {
	agents map[string]AgentDefinition
}

// NewRegistry creates a new empty agent registry.
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]AgentDefinition),
	}
}

// Load reads agent definitions from a YAML file.
func (r *Registry) Load(path string) error {
	if !config.FileExists(path) {
		return fmt.Errorf("agents file not found: %s (run 'forgecrew init' first)", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read agents file: %w", err)
	}
	var af agentsFile
	if err := yaml.Unmarshal(data, &af); err != nil {
		return fmt.Errorf("parse agents file: %w", err)
	}

	for id, def := range af.Agents {
		if def.AgentID == "" {
			def.AgentID = id
		}
		if err := validateAgent(def); err != nil {
			return fmt.Errorf("agent %s: %w", id, err)
		}
		r.agents[id] = def
	}
	return nil
}

// Get returns a single agent definition by ID.
func (r *Registry) Get(id string) (*AgentDefinition, error) {
	a, ok := r.agents[id]
	if !ok {
		return nil, fmt.Errorf("agent not found: %s", id)
	}
	return &a, nil
}

// List returns all agent definitions.
func (r *Registry) List() []AgentDefinition {
	result := make([]AgentDefinition, 0, len(r.agents))
	for _, a := range r.agents {
		result = append(result, a)
	}
	return result
}

func validateAgent(a AgentDefinition) error {
	if a.AgentID == "" {
		return fmt.Errorf("agent_id is required")
	}
	if a.Name == "" {
		return fmt.Errorf("name is required")
	}
	if a.DefaultModel == "" {
		return fmt.Errorf("default_model is required")
	}
	validPermissions := map[string]bool{
		"read_only":                   true,
		"patch_only":                  true,
		"write_with_approval":         true,
		"auto_write_in_sandbox":       true,
		"commit_with_approval":        true,
		"deploy_with_manual_approval": true,
	}
	if !validPermissions[a.PermissionLevel] {
		return fmt.Errorf("invalid permission_level: %s", a.PermissionLevel)
	}
	return nil
}
