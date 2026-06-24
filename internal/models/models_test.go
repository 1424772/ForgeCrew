package models

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
}

func TestLoadModels(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "models.yaml")

	yamlContent := `
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
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewRegistry()
	if err := r.Load(path); err != nil {
		t.Fatal(err)
	}

	m, err := r.Get("gpt_5_5")
	if err != nil {
		t.Fatal(err)
	}
	if m.Provider != "openai" {
		t.Errorf("provider = %q, want openai", m.Provider)
	}
	if m.CostTier != "high" {
		t.Errorf("cost_tier = %q, want high", m.CostTier)
	}
	if !m.SupportsTools {
		t.Error("supports_tools should be true")
	}
}

func TestLoadMissingFile(t *testing.T) {
	r := NewRegistry()
	err := r.Load("/nonexistent/models.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestGetMissingModel(t *testing.T) {
	r := NewRegistry()
	_, err := r.Get("nonexistent")
	if err == nil {
		t.Error("expected error for missing model")
	}
}

func TestValidateModel(t *testing.T) {
	tests := []struct {
		name    string
		model   ModelDefinition
		wantErr bool
	}{
		{
			name: "valid",
			model: ModelDefinition{
				ModelID: "gpt_5_5", Provider: "openai", Model: "gpt-5.5",
				APIKeyEnv: "OPENAI_API_KEY", CostTier: "high",
			},
			wantErr: false,
		},
		{
			name: "missing provider",
			model: ModelDefinition{
				ModelID: "gpt_5_5", Model: "gpt-5.5",
				APIKeyEnv: "OPENAI_API_KEY", CostTier: "high",
			},
			wantErr: true,
		},
		{
			name: "invalid cost tier",
			model: ModelDefinition{
				ModelID: "gpt_5_5", Provider: "openai", Model: "gpt-5.5",
				APIKeyEnv: "OPENAI_API_KEY", CostTier: "premium",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateModel(tt.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateModel() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
