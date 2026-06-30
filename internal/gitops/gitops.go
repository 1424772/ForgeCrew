// Package gitops provides controlled git operations.
// Direct shell/git access should go through this package.
package gitops

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ReadDiff runs git diff on the given repository root and returns the output.
// Returns a friendly error if the directory is not a git repository.
func ReadDiff(root string) (string, error) {
	if !isGitRepo(root) {
		return "", fmt.Errorf("not a git repository: %s (run 'git init' first)", root)
	}

	var result strings.Builder

	// Run git diff for tracked changes.
	diffCmd := exec.Command("git", "diff")
	diffCmd.Dir = root
	diffOut, err := diffCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	diff := string(diffOut)

	// Run git status --short for untracked files.
	statusCmd := exec.Command("git", "status", "--short")
	statusCmd.Dir = root
	statusOut, err := statusCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}
	status := strings.TrimSpace(string(statusOut))

	hasUntracked := false
	var untrackedLines []string
	for _, line := range strings.Split(status, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Untracked files start with "??".
		if strings.HasPrefix(line, "??") {
			hasUntracked = true
			untrackedLines = append(untrackedLines, "  "+line)
		}
	}

	if diff != "" {
		result.WriteString(diff)
	}
	if hasUntracked {
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString("Untracked files:\n")
		for _, line := range untrackedLines {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	if result.Len() == 0 {
		return "No changes (working tree clean)\n", nil
	}
	return result.String(), nil
}

// GetCurrentHash returns the current HEAD git hash.
func GetCurrentHash(root string) (string, error) {
	if !isGitRepo(root) {
		return "", fmt.Errorf("not a git repository: %s", root)
	}
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// IsGitRepo returns true if the directory is inside a git repository.
func IsGitRepo(root string) bool {
	return isGitRepo(root)
}

func isGitRepo(root string) bool {
	gitDir := root + string(os.PathSeparator) + ".git"
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ChangedFiles returns tracked changed files (git diff --name-only) merged
// with untracked files (git status --short ?? lines). Results are deduplicated.
// Returns nil slice and nil error if not in a git repo.
func ChangedFiles(root string) ([]string, error) {
	if !isGitRepo(root) {
		return nil, nil
	}

	seen := make(map[string]bool)
	var files []string

	// Collect tracked changes.
	if out, err := exec.Command("git", "diff", "--name-only").Output(); err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !seen[line] {
				seen[line] = true
				files = append(files, line)
			}
		}
	}

	// Collect untracked files from git status --short.
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return files, nil
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "??") {
			// Format: "?? path/to/file" — extract path.
			path := strings.TrimSpace(line[2:])
			if path != "" && !seen[path] {
				seen[path] = true
				files = append(files, path)
			}
		}
	}

	return files, nil
}

// ReadDiffFull returns the full git diff plus untracked file listing.
// This delegates to ReadDiff which already handles both tracked and untracked.
func ReadDiffFull(root string) (string, error) {
	return ReadDiff(root)
}
