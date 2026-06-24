package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigConstants(t *testing.T) {
	if AppName != "ForgeCrew" {
		t.Errorf("AppName = %q, want %q", AppName, "ForgeCrew")
	}
	if ConfigDir != ".forgecrew" {
		t.Errorf("ConfigDir = %q, want %q", ConfigDir, ".forgecrew")
	}
	if CLIName != "forgecrew" {
		t.Errorf("CLIName = %q, want %q", CLIName, "forgecrew")
	}
}

func TestFileExists(t *testing.T) {
	tmp := t.TempDir()

	// File doesn't exist.
	if FileExists(filepath.Join(tmp, "nonexistent")) {
		t.Error("FileExists should return false for non-existent file")
	}

	// Create a file.
	path := filepath.Join(tmp, "test.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if !FileExists(path) {
		t.Error("FileExists should return true for existing file")
	}

	// Directory should return false.
	if FileExists(tmp) {
		t.Error("FileExists should return false for directories")
	}
}

func TestEnsureDir(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "newdir")
	if err := EnsureDir(dir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("EnsureDir should create a directory")
	}

	// Idempotent.
	if err := EnsureDir(dir); err != nil {
		t.Error("EnsureDir should be idempotent")
	}
}

func TestSaveAndLoadYAML(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.yaml")

	type TestConfig struct {
		Name  string `yaml:"name"`
		Value int    `yaml:"value"`
	}
	original := TestConfig{Name: "test", Value: 42}

	if err := SaveYAML(path, &original); err != nil {
		t.Fatal(err)
	}

	var loaded TestConfig
	if err := LoadYAML(path, &loaded); err != nil {
		t.Fatal(err)
	}
	if loaded.Name != "test" || loaded.Value != 42 {
		t.Errorf("loaded = %+v, want name=test value=42", loaded)
	}
}
