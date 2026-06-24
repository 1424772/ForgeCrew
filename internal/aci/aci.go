// Package aci implements the Agent-Computer Interface (ACI) layer.
//
// ACI provides structured, permission-controlled actions that agents
// use instead of directly accessing the shell, filesystem, or git.
// Based on SWE-agent's ACI design philosophy.
package aci

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/1424772/ForgeCrew/internal/gitops"
)

// ErrNotImplemented is returned for actions that exist as interfaces but
// are not yet implemented in this version.
var ErrNotImplemented = fmt.Errorf("not implemented")

// Action represents a single ACI action result.
type Action struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	RiskLevel   string `json:"risk_level"`
}

// AllActions returns the complete list of ACI actions with metadata.
func AllActions() []Action {
	return []Action{
		{Name: "ReadFile", Description: "Read a file from disk", RiskLevel: "low"},
		{Name: "SearchCode", Description: "Search code with ripgrep", RiskLevel: "low"},
		{Name: "GeneratePatch", Description: "Generate a code patch/diff", RiskLevel: "medium"},
		{Name: "RunTest", Description: "Run project tests", RiskLevel: "medium"},
		{Name: "ReadGitDiff", Description: "Read current git diff", RiskLevel: "low"},
		{Name: "CreateCheckpoint", Description: "Create a checkpoint", RiskLevel: "low"},
		{Name: "Rollback", Description: "Rollback to a checkpoint", RiskLevel: "high"},
	}
}

// ReadFile reads a file and returns its content.
// path must be within the project root.
func ReadFile(root, path string) (string, error) {
	fullPath := filepath.Join(root, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", path, err)
	}
	return string(data), nil
}

// SearchCode performs a ripgrep search in the project root.
func SearchCode(root, pattern string) (string, error) {
	// Use ripgrep if available; this is the interface layer.
	return "", fmt.Errorf("SearchCode: %w (will use ripgrep integration)", ErrNotImplemented)
}

// GeneratePatch is the controlled interface for generating a code diff/patch.
func GeneratePatch(root string, changes []FileChange) (string, error) {
	return "", fmt.Errorf("GeneratePatch: %w", ErrNotImplemented)
}

// FileChange represents a change to a single file.
type FileChange struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Action  string `json:"action"` // create, update, delete
}

// RunTest executes the project's test suite.
func RunTest(root string) (string, error) {
	return "", fmt.Errorf("RunTest: %w", ErrNotImplemented)
}

// ReadGitDiff returns the current git diff.
func ReadGitDiff(root string) (string, error) {
	return gitops.ReadDiff(root)
}

// Rollback reverts to a previous checkpoint.
func Rollback(checkpointID string) error {
	return fmt.Errorf("Rollback: %w", ErrNotImplemented)
}
