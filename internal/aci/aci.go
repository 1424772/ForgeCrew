// Package aci implements the Agent-Computer Interface (ACI) layer.
//
// ACI provides structured, permission-controlled actions that agents
// use instead of directly accessing the shell, filesystem, or git.
// Based on SWE-agent's ACI design philosophy.
package aci

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/1424772/ForgeCrew/internal/checkpoint"
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

// resolvePath resolves and validates a file path within a project root.
// It returns an error if:
//   - path is absolute (must be relative)
//   - the resolved path escapes outside root
func resolvePath(root, path string) (string, error) {
	// Clean and make both paths absolute for comparison.
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}

	// Reject absolute paths.
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("absolute path not allowed: %s (must be relative)", path)
	}

	// Clean the path to normalize .. and .
	cleanPath := filepath.Clean(path)

	// Resolve against the root.
	resolved := filepath.Join(absRoot, cleanPath)
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	// Check that the resolved path is within root (no traversal escape).
	// filepath.Rel should produce a relative path without ".." prefix.
	rel, err := filepath.Rel(absRoot, resolved)
	if err != nil {
		return "", fmt.Errorf("path escapes root: %s", path)
	}
	if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path escape detected: %s resolves outside root", path)
	}

	return resolved, nil
}

// ReadFile reads a file and returns its content.
// path must be relative and within the project root.
// Symlink escape attempts are detected and rejected.
func ReadFile(root, path string) (string, error) {
	resolved, err := resolvePath(root, path)
	if err != nil {
		return "", err
	}

	// Resolve symlinks to prevent symlink escape attacks.
	realPath, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", path, err)
	}

	// Re-validate that the real path (after symlink resolution) is still within root.
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", path, err)
	}
	rel, err := filepath.Rel(absRoot, realPath)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path escape detected via symlink: %s resolves outside root", path)
	}

	data, err := os.ReadFile(realPath)
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", path, err)
	}
	return string(data), nil
}

// SearchCode performs a code search within the project root.
// It tries ripgrep first; if rg is unavailable or cannot execute,
// it falls back to a pure-Go substring search.
func SearchCode(root, pattern string) (string, error) {
	// Try rg first.
	result, err := tryRg(root, pattern)
	if err == nil {
		return result, nil
	}
	// rg failed or not found; use pure Go fallback.
	return searchCodeFallback(root, pattern)
}

// tryRg runs ripgrep and returns its output.
// Any error (including rg not found or execution failure) signals
// the caller to use the fallback.
func tryRg(root, pattern string) (string, error) {
	rgPath, err := exec.LookPath("rg")
	if err != nil {
		return "", err
	}

	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", err
	}

	cmd := exec.Command(rgPath, "--no-heading", "--line-number", "--color", "never", "--sort", "path", pattern, ".")
	cmd.Dir = absRoot

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// rg exits 1 when no matches found.
			if exitErr.ExitCode() == 1 {
				return "", nil
			}
		}
		// Any other error (Access denied, etc.) → fallback.
		return "", err
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// skipDirs lists directory names the pure-Go fallback should skip.
var skipDirs = map[string]bool{
	".git":         true,
	".forgecrew":   true,
	"vendor":       true,
	"node_modules": true,
	".claude":      true,
}

// searchCodeFallback performs a pure-Go substring search without any
// external tool dependency. It walks the directory tree, skips common
// noise directories and binary files, and returns results in a stable
// order matching ripgrep's default format.
func searchCodeFallback(root, pattern string) (string, error) {
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}

	var results []string
	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip files we cannot access
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		// Skip likely-binary files.
		if isBinaryFile(path) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if strings.Contains(line, pattern) {
				results = append(results, fmt.Sprintf("%s:%d:%s", relPath, i+1, line))
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("search fallback: %w", err)
	}

	// Stable, deterministic output.
	sort.Strings(results)
	return strings.Join(results, "\n"), nil
}

// isBinaryFile checks whether a file is likely binary by looking for
// a zero byte within its first 8000 bytes.
func isBinaryFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true // cannot open → skip
	}
	defer f.Close()

	buf := make([]byte, 8000)
	n, _ := f.Read(buf)
	for _, b := range buf[:n] {
		if b == 0 {
			return true
		}
	}
	return false
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

// CreateCheckpoint creates a checkpoint via the checkpoint store.
// This is the ACI-controlled path; agents must use this, never the store directly.
// The checkpoint is stored under root/.forgecrew/checkpoints/.
func CreateCheckpoint(root, taskID, agentID string, changedFiles []string) (*checkpoint.Checkpoint, error) {
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return nil, fmt.Errorf("resolve root: %w", err)
	}
	store := checkpoint.NewStoreAt(filepath.Join(absRoot, ".forgecrew", "checkpoints"))
	return store.Create(taskID, agentID, changedFiles)
}

// Rollback reverts to a previous checkpoint.
func Rollback(checkpointID string) error {
	return fmt.Errorf("Rollback: %w", ErrNotImplemented)
}
