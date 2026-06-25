// Package settings manages ForgeCrew project settings stored in
// .forgecrew/settings.yaml.
package settings

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SettingsYAML is the filename within .forgecrew/.
const SettingsYAML = "settings.yaml"

// Settings represents the project-level ForgeCrew settings.
type Settings struct {
	Language string `yaml:"language"`
}

// Default returns the default settings (language: zh).
func Default() *Settings {
	return &Settings{Language: "zh"}
}

// Load reads settings from path. If the file does not exist, it returns
// the default settings without error (so callers always have a valid config).
// It validates the language field; invalid values produce an error.
func Load(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("read settings: %w", err)
	}
	var s Settings
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}
	// Default to zh if empty.
	if s.Language == "" {
		s.Language = "zh"
		return &s, nil
	}
	// Validate: only zh and en are supported.
	switch s.Language {
	case "zh", "en":
		return &s, nil
	default:
		return nil, fmt.Errorf("unsupported language: %q (supported: zh, en)", s.Language)
	}
}

// Save writes settings to path as YAML.
func Save(path string, s *Settings) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	return nil
}
