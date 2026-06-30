// Package config provides configuration constants and utilities for ForgeCrew.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Core configuration constants.
const (
	// AppName is the human-readable product name.
	AppName = "ForgeCrew"
	// ConfigDir is the configuration directory name, prefixed with dot.
	ConfigDir = ".forgecrew"
	// CLIName is the CLI binary/command name.
	CLIName = "forgecrew"
)

// Config file names within .forgecrew/.
const (
	AgentsYAML     = "agents.yaml"
	ModelsYAML     = "models.yaml"
	WorkflowsYAML  = "workflows.yaml"
	SettingsYAML   = "settings.yaml"
	TeamsYAML      = "teams.yaml"
	MemoryDir      = "memory"
	EvalsDir       = "evals"
	CheckpointsDir = "checkpoints"
)

// Project-level files of interest.
const (
	AgentsMDFile = "AGENTS.md"
)

// PrintError formats and prints an error to stderr.
func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "%s error: %v\n", CLIName, err)
}

// LoadYAML reads and unmarshals a YAML file into the provided target.
func LoadYAML(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

// SaveYAML marshals and writes a YAML file.
func SaveYAML(path string, data interface{}) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}
	if err := os.WriteFile(path, out, 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// EnsureDir creates a directory if it does not exist.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create dir %s: %w", path, err)
	}
	return nil
}

// FileExists checks if a file exists and is not a directory.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// SkipDirs lists directory names that tools should skip during traversal.
// These are common noise directories: VCS, package managers, build artifacts.
var SkipDirs = map[string]bool{
	".git":         true,
	".forgecrew":   true,
	"vendor":       true,
	"node_modules": true,
	".claude":      true,
}
