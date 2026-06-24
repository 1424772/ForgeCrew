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
	cmd := exec.Command("git", "diff")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	diff := string(out)
	if diff == "" {
		return "No changes (working tree clean)\n", nil
	}
	return diff, nil
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
