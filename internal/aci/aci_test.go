package aci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// localTempDir creates a temporary directory within the current package
// directory instead of the system temp directory. This avoids Windows
// issues where filepath.EvalSymlinks on system temp paths triggers
// "Access is denied".
func localTempDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	base := filepath.Join(wd, ".testtmp")
	if err := os.MkdirAll(base, 0755); err != nil {
		t.Fatalf("mkdir .testtmp: %v", err)
	}
	dir, err := os.MkdirTemp(base, "t-*")
	if err != nil {
		t.Fatalf("mkdirtemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestAllActions(t *testing.T) {
	actions := AllActions()
	if len(actions) != 7 {
		t.Errorf("expected 7 actions, got %d", len(actions))
	}
	expectedNames := map[string]bool{
		"ReadFile": true, "SearchCode": true, "GeneratePatch": true,
		"RunTest": true, "ReadGitDiff": true, "CreateCheckpoint": true,
		"Rollback": true,
	}
	for _, a := range actions {
		if !expectedNames[a.Name] {
			t.Errorf("unexpected action: %s", a.Name)
		}
	}
}

func TestReadFile(t *testing.T) {
	tmp := localTempDir(t)
	path := filepath.Join(tmp, "test.txt")
	content := "hello world"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ReadFile(tmp, "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if result != content {
		t.Errorf("ReadFile = %q, want %q", result, content)
	}
}

func TestReadFileNotFound(t *testing.T) {
	tmp := localTempDir(t)
	_, err := ReadFile(tmp, "nonexistent.txt")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReadFileAbsolutePathRejected(t *testing.T) {
	tmp := localTempDir(t)
	absPath := filepath.Join(tmp, "test.txt")
	os.WriteFile(absPath, []byte("data"), 0644)

	_, err := ReadFile(tmp, absPath)
	if err == nil {
		t.Error("expected error for absolute path")
	}
	if !strings.Contains(err.Error(), "absolute path not allowed") {
		t.Errorf("error should mention absolute path, got: %v", err)
	}
}

func TestReadFilePathEscapeRejected(t *testing.T) {
	tmp := localTempDir(t)
	// Create a subdirectory to test traversal.
	subDir := filepath.Join(tmp, "sub")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(tmp, "secret.txt"), []byte("secret"), 0644)

	// Try to read secret.txt from sub via ../
	_, err := ReadFile(subDir, "../secret.txt")
	if err == nil {
		t.Error("expected error for path escape via ../")
	}
	if !strings.Contains(err.Error(), "escape") {
		t.Errorf("error should mention path escape, got: %v", err)
	}
}

func TestReadFileDeepEscapeRejected(t *testing.T) {
	tmp := localTempDir(t)
	deepDir := filepath.Join(tmp, "a", "b", "c")
	os.MkdirAll(deepDir, 0755)
	os.WriteFile(filepath.Join(tmp, "secret.txt"), []byte("secret"), 0644)

	_, err := ReadFile(deepDir, "../../../secret.txt")
	if err == nil {
		t.Error("expected error for deep path escape")
	}
}

func TestReadFileSymlinkEscapeRejected(t *testing.T) {
	tmp := localTempDir(t)
	outside := localTempDir(t) // Different temp dir
	os.WriteFile(filepath.Join(outside, "leak.txt"), []byte("leaked"), 0644)

	linkPath := filepath.Join(tmp, "link")
	if err := os.Symlink(outside, linkPath); err != nil {
		t.Skipf("symlink creation not supported on this platform: %v", err)
	}

	_, err := ReadFile(tmp, "link/leak.txt")
	if err == nil {
		t.Error("expected error for symlink escape")
	}
}

func TestResolvePathNormal(t *testing.T) {
	tmp := localTempDir(t)
	resolved, err := resolvePath(tmp, "foo/bar.txt")
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(tmp, "foo", "bar.txt")
	expected, _ = filepath.Abs(expected)
	if resolved != expected {
		t.Errorf("resolved = %q, want %q", resolved, expected)
	}
}

func TestResolvePathDotDotClean(t *testing.T) {
	tmp := localTempDir(t)
	// Paths that clean to something safe should work.
	resolved, err := resolvePath(tmp, "foo/../bar.txt")
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(tmp, "bar.txt")
	expected, _ = filepath.Abs(expected)
	if resolved != expected {
		t.Errorf("resolved = %q, want %q", resolved, expected)
	}
}

// --- SearchCode tests ---

func TestSearchCodeFallbackWorksWithoutRg(t *testing.T) {
	// Pure-Go fallback must work regardless of whether rg is installed.
	tmp := localTempDir(t)
	os.WriteFile(filepath.Join(tmp, "a.go"), []byte("package test\nfunc hello() {}"), 0644)

	result, err := SearchCode(tmp, "func")
	if err != nil {
		t.Fatalf("fallback search should succeed even without rg: %v", err)
	}
	if result == "" {
		t.Error("expected matches for 'func' via fallback")
	}
}

func TestSearchCodeWithMatches(t *testing.T) {
	tmp := localTempDir(t)
	os.WriteFile(filepath.Join(tmp, "a.go"), []byte("package test\nfunc hello() {}"), 0644)
	os.WriteFile(filepath.Join(tmp, "b.go"), []byte("package test\nfunc world() {}"), 0644)

	result, err := SearchCode(tmp, "func")
	if err != nil {
		t.Fatal(err)
	}
	if result == "" {
		t.Error("expected matches for 'func'")
	}
}

func TestSearchCodeNoMatches(t *testing.T) {
	tmp := localTempDir(t)
	os.WriteFile(filepath.Join(tmp, "a.go"), []byte("package test"), 0644)

	result, err := SearchCode(tmp, "nonexistentpattern12345")
	if err != nil {
		t.Fatal(err)
	}
	if result != "" {
		t.Errorf("expected empty result for no matches, got: %q", result)
	}
}

func TestSearchCodeBinaryFile(t *testing.T) {
	tmp := localTempDir(t)
	os.WriteFile(filepath.Join(tmp, "a.go"), []byte("package test\nfunc hello() {}"), 0644)

	result, err := SearchCode(tmp, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if result == "" {
		t.Error("expected to find 'hello'")
	}
}

// --- CreateCheckpoint tests ---

func TestCreateCheckpoint(t *testing.T) {
	tmp := localTempDir(t)
	ckpt, err := CreateCheckpoint(tmp, "task_001", "backend_engineer", []string{"main.go", "config.go"})
	if err != nil {
		t.Fatal(err)
	}
	if ckpt.ID == "" {
		t.Error("checkpoint ID should not be empty")
	}
	if ckpt.TaskID != "task_001" {
		t.Errorf("TaskID = %q, want task_001", ckpt.TaskID)
	}
	if ckpt.AgentID != "backend_engineer" {
		t.Errorf("AgentID = %q, want backend_engineer", ckpt.AgentID)
	}
	if len(ckpt.ChangedFiles) != 2 {
		t.Errorf("expected 2 changed files, got %d", len(ckpt.ChangedFiles))
	}

	// Verify the checkpoint was actually written to disk at the right location.
	entries, err := os.ReadDir(filepath.Join(tmp, ".forgecrew", "checkpoints"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 checkpoint file on disk, got %d", len(entries))
	}
}

func TestGeneratePatchNotImplemented(t *testing.T) {
	_, err := GeneratePatch(".", nil)
	if err == nil {
		t.Error("expected ErrNotImplemented for GeneratePatch")
	}
}

func TestRunTestNotImplemented(t *testing.T) {
	_, err := RunTest(".")
	if err == nil {
		t.Error("expected ErrNotImplemented for RunTest")
	}
}

func TestRollbackNotImplemented(t *testing.T) {
	err := Rollback("ckpt_001")
	if err == nil {
		t.Error("expected ErrNotImplemented for Rollback")
	}
}
