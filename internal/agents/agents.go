// Package agents provides the Agent Registry — definitions, loading, and querying.
package agents

import (
	"fmt"
	"os"
	"strings"

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

// ModelLookup is the interface the agent registry needs from a model
// registry for cross-validation.
type ModelLookup interface {
	Exists(id string) bool
}

// ValidateModelRefs checks that every agent's DefaultModel and FallbackModels
// reference models that exist in the model registry. It returns nil if all
// references are valid, or an error listing every broken reference.
func (r *Registry) ValidateModelRefs(models ModelLookup) error {
	var errs []string
	for _, a := range r.agents {
		if !models.Exists(a.DefaultModel) {
			errs = append(errs, fmt.Sprintf("agent %q: default_model %q does not exist in model registry", a.AgentID, a.DefaultModel))
		}
		for _, fm := range a.FallbackModels {
			if !models.Exists(fm) {
				errs = append(errs, fmt.Sprintf("agent %q: fallback_model %q does not exist in model registry", a.AgentID, fm))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("agent-model cross-validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

// ValidateTools checks that every agent's Tools entries match known names
// (ACI actions or system-reserved tools). Comparison is case-insensitive
// with underscores stripped so snake_case tool names match PascalCase ACI
// actions. Pass nil for knownNames to skip validation.
func (r *Registry) ValidateTools(knownNames map[string]bool) error {
	if knownNames == nil {
		return nil
	}
	// Build normalized lookup.
	normalized := make(map[string]bool, len(knownNames))
	for name := range knownNames {
		normalized[normalizeName(name)] = true
	}
	var errs []string
	for _, a := range r.agents {
		for _, tool := range a.Tools {
			if !normalized[normalizeName(tool)] {
				errs = append(errs, fmt.Sprintf("agent %q: tool %q is not a known ACI action or system-reserved tool", a.AgentID, tool))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("agent-tool validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

// normalizeName folds a name for comparison: lowercase + remove underscores.
func normalizeName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, "_", ""))
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
