package aci

import (
	"os"
	"path/filepath"
	"testing"
)

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
	tmp := t.TempDir()
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
	tmp := t.TempDir()
	_, err := ReadFile(tmp, "nonexistent.txt")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestSearchCodeNotImplemented(t *testing.T) {
	_, err := SearchCode(".", "pattern")
	if err == nil {
		t.Error("expected ErrNotImplemented for SearchCode")
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
