package gitops

import (
	"testing"
)

func TestIsGitRepo(t *testing.T) {
	// The project root should be a git repo (we initialized it).
	if !isGitRepo(".") {
		t.Skip("current dir is not a git repo, skipping")
	}
	if !IsGitRepo(".") {
		t.Error("should detect current dir as git repo")
	}
}

func TestIsNotGitRepo(t *testing.T) {
	tmp := t.TempDir()
	if isGitRepo(tmp) {
		t.Error("temp dir should not be a git repo")
	}
}

func TestGetCurrentHash(t *testing.T) {
	if !isGitRepo(".") {
		t.Skip("not a git repo")
	}
	hash, err := GetCurrentHash(".")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "" {
		t.Error("hash should not be empty")
	}
}

func TestReadDiffNotRepo(t *testing.T) {
	tmp := t.TempDir()
	_, err := ReadDiff(tmp)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}
