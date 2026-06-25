// Package models provides the Model Registry — definitions, loading, and querying.
package models

import (
	"fmt"
	"os"

	"github.com/1424772/ForgeCrew/internal/config"
	"gopkg.in/yaml.v3"
)

// ModelDefinition describes a single model's configuration.
type ModelDefinition struct {
	ModelID        string `yaml:"-" json:"model_id"`
	Provider       string `yaml:"provider" json:"provider"`
	Model          string `yaml:"model" json:"model"`
	APIKeyEnv      string `yaml:"api_key_env" json:"api_key_env"`
	Role           string `yaml:"role" json:"role"`
	CostTier       string `yaml:"cost_tier" json:"cost_tier"`
	SupportsTools  bool   `yaml:"supports_tools" json:"supports_tools"`
	SupportsVision bool   `yaml:"supports_vision" json:"supports_vision"`
}

// modelsFile is the YAML file structure for models.
type modelsFile struct {
	Models map[string]ModelDefinition `yaml:"models"`
}

// Registry stores and manages model definitions.
type Registry struct {
	models map[string]ModelDefinition
}

// NewRegistry creates a new empty model registry.
func NewRegistry() *Registry {
	return &Registry{
		models: make(map[string]ModelDefinition),
	}
}

// Load reads model definitions from a YAML file.
func (r *Registry) Load(path string) error {
	if !config.FileExists(path) {
		return fmt.Errorf("models file not found: %s (run 'forgecrew init' first)", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read models file: %w", err)
	}
	var mf modelsFile
	if err := yaml.Unmarshal(data, &mf); err != nil {
		return fmt.Errorf("parse models file: %w", err)
	}

	for id, def := range mf.Models {
		def.ModelID = id
		if err := validateModel(def); err != nil {
			return fmt.Errorf("model %s: %w", id, err)
		}
		r.models[id] = def
	}
	return nil
}

// Get returns a single model definition by ID.
func (r *Registry) Get(id string) (*ModelDefinition, error) {
	m, ok := r.models[id]
	if !ok {
		return nil, fmt.Errorf("model not found: %s", id)
	}
	return &m, nil
}

// Exists returns true if a model with the given ID is registered.
func (r *Registry) Exists(id string) bool {
	_, ok := r.models[id]
	return ok
}

// ModelIDs returns the set of all registered model IDs.
func (r *Registry) ModelIDs() map[string]bool {
	ids := make(map[string]bool, len(r.models))
	for id := range r.models {
		ids[id] = true
	}
	return ids
}

// List returns all model definitions.
func (r *Registry) List() []ModelDefinition {
	result := make([]ModelDefinition, 0, len(r.models))
	for _, m := range r.models {
		result = append(result, m)
	}
	return result
}

func validateModel(m ModelDefinition) error {
	if m.ModelID == "" {
		return fmt.Errorf("model id is required")
	}
	if m.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if m.Model == "" {
		return fmt.Errorf("model is required")
	}
	if m.APIKeyEnv == "" {
		return fmt.Errorf("api_key_env is required")
	}
	validCostTiers := map[string]bool{"high": true, "medium": true, "low": true}
	if !validCostTiers[m.CostTier] {
		return fmt.Errorf("invalid cost_tier: %s (must be high/medium/low)", m.CostTier)
	}
	return nil
}
