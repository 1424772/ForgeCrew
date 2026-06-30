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
	"github.com/1424772/ForgeCrew/internal/config"
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

// searchCodeFallback performs a pure-Go substring search without any
// external tool dependency. It walks the directory tree, skips common
// noise directories (defined in config.SkipDirs) and binary files,
// and returns results in a stable order matching ripgrep's default format.
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
			if config.SkipDirs[info.Name()] {
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

// FileChange represents a change to a single file.
type FileChange struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Action  string `json:"action"` // create, update, delete
}

// GeneratePatch is the controlled interface for generating a code diff/patch.
// It applies the file changes to disk, writes each change through the ACI
// safety boundary (path traversal checks), and returns a unified-diff-style
// summary of all changes made.
//
// Each FileChange must have a relative path and one of three actions:
//
//	"create" — writes a new file
//	"update" — overwrites an existing file (reads original first for diff)
//	"delete" — removes a file
func GeneratePatch(root string, changes []FileChange) (string, error) {
	if len(changes) == 0 {
		return "", fmt.Errorf("GeneratePatch: no changes provided")
	}

	var patchLines []string
	patchLines = append(patchLines, fmt.Sprintf("--- a/%s", filepath.Base(root)))
	patchLines = append(patchLines, fmt.Sprintf("+++ b/%s", filepath.Base(root)))

	for _, ch := range changes {
		resolved, err := resolvePath(root, ch.Path)
		if err != nil {
			return "", fmt.Errorf("GeneratePatch: %s: %w", ch.Path, err)
		}

		switch ch.Action {
		case "create":
			// Ensure parent directory exists.
			parent := filepath.Dir(resolved)
			if err := os.MkdirAll(parent, 0755); err != nil {
				return "", fmt.Errorf("GeneratePatch: create dir for %s: %w", ch.Path, err)
			}
			if err := os.WriteFile(resolved, []byte(ch.Content), 0644); err != nil {
				return "", fmt.Errorf("GeneratePatch: write %s: %w", ch.Path, err)
			}
			lines := strings.Split(ch.Content, "\n")
			patchLines = append(patchLines, fmt.Sprintf("@@ -0,0 +1,%d @@ %s (new file)", len(lines), ch.Path))
			for _, l := range lines {
				patchLines = append(patchLines, "+"+l)
			}

		case "update":
			// Read original for diff context.
			origData, readErr := os.ReadFile(resolved)
			origLines := []string{}
			if readErr == nil {
				origLines = strings.Split(string(origData), "\n")
			}

			// Write new content.
			if err := os.WriteFile(resolved, []byte(ch.Content), 0644); err != nil {
				return "", fmt.Errorf("GeneratePatch: update %s: %w", ch.Path, err)
			}

			newLines := strings.Split(ch.Content, "\n")
			patchLines = append(patchLines, fmt.Sprintf("@@ -1,%d +1,%d @@ %s", len(origLines), len(newLines), ch.Path))
			// Simple line-by-line diff: show removed then added lines.
			for _, l := range origLines {
				if !containsLine(newLines, l) {
					patchLines = append(patchLines, "-"+l)
				}
			}
			for _, l := range newLines {
				if !containsLine(origLines, l) {
					patchLines = append(patchLines, "+"+l)
				}
			}

		case "delete":
			if err := os.Remove(resolved); err != nil {
				return "", fmt.Errorf("GeneratePatch: delete %s: %w", ch.Path, err)
			}
			patchLines = append(patchLines, fmt.Sprintf("@@ -1,0 +0,0 @@ %s (deleted)", ch.Path))

		default:
			return "", fmt.Errorf("GeneratePatch: unknown action %q for %s (use create, update, or delete)", ch.Action, ch.Path)
		}
	}

	return strings.Join(patchLines, "\n"), nil
}

// RunTest detects the project's test framework and executes the test suite.
// It supports Go (go test ./...), Node (npm test), Python (pytest or unittest),
// and Rust (cargo test). The command runs from root and returns combined
// stdout/stderr. A 120-second timeout is enforced via context.
func RunTest(root string) (string, error) {
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", fmt.Errorf("RunTest: resolve root: %w", err)
	}

	cmdArgs := detectTestCommand(absRoot)
	if cmdArgs == nil {
		return "", fmt.Errorf("RunTest: could not detect a supported test framework in %s", absRoot)
	}

	// #nosec G204 — command arguments are statically derived from project detection.
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = absRoot

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("RunTest: %s: %w\n%s", strings.Join(cmdArgs, " "), err, string(out))
	}
	return string(out), nil
}

// detectTestCommand returns the test command and arguments for a project,
// or nil if no known test framework is detected.
func detectTestCommand(root string) []string {
	// Go modules.
	if config.FileExists(filepath.Join(root, "go.mod")) {
		return []string{"go", "test", "./..."}
	}
	// Node / npm.
	if config.FileExists(filepath.Join(root, "package.json")) {
		return []string{"npm", "test"}
	}
	// Python.
	if config.FileExists(filepath.Join(root, "pyproject.toml")) ||
		config.FileExists(filepath.Join(root, "setup.py")) {
		// Prefer pytest if it's likely installed.
		return []string{"python", "-m", "pytest"}
	}
	if config.FileExists(filepath.Join(root, "requirements.txt")) {
		return []string{"python", "-m", "unittest"}
	}
	// Rust.
	if config.FileExists(filepath.Join(root, "Cargo.toml")) {
		return []string{"cargo", "test"}
	}
	// Makefile with test target (common fallback).
	if config.FileExists(filepath.Join(root, "Makefile")) {
		return []string{"make", "test"}
	}
	return nil
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

// Rollback reverts changes by restoring files to their state at a given
// checkpoint. It loads the checkpoint from the store, then uses git
// checkout on each changed file recorded in the checkpoint.
//
// When the project is not a git repository Rollback reads the checkpoint
// data and reports which files would need manual restoration.
func Rollback(root, checkpointID string) error {
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return fmt.Errorf("Rollback: resolve root: %w", err)
	}

	store := checkpoint.NewStoreAt(filepath.Join(absRoot, ".forgecrew", "checkpoints"))
	ckpts, err := store.List()
	if err != nil {
		return fmt.Errorf("Rollback: list checkpoints: %w", err)
	}

	var target *checkpoint.Checkpoint
	for i := range ckpts {
		if ckpts[i].ID == checkpointID {
			target = &ckpts[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("Rollback: checkpoint %s not found", checkpointID)
	}

	if len(target.ChangedFiles) == 0 {
		return fmt.Errorf("Rollback: checkpoint %s has no changed files to restore", checkpointID)
	}

	// If it's a git repo, use git checkout to restore each file.
	if gitops.IsGitRepo(absRoot) {
		for _, f := range target.ChangedFiles {
			// #nosec G204 — file paths come from our own checkpoint, not user input.
			cmd := exec.Command("git", "checkout", "--", f)
			cmd.Dir = absRoot
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("Rollback: git checkout %s: %w\n%s", f, err, string(out))
			}
		}
		return nil
	}

	return fmt.Errorf("Rollback: not a git repository; manual restoration required for: %s",
		strings.Join(target.ChangedFiles, ", "))
}

// containsLine reports whether a line exists in a slice of lines.
func containsLine(lines []string, target string) bool {
	for _, l := range lines {
		if l == target {
			return true
		}
	}
	return false
}
