package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/1424772/ForgeCrew/internal/settings"
)

// chdirTempDir creates a temp dir, chdirs into it, and returns the path.
// The test is run inside this directory so that commands operating on "."
// (like init) see an isolated filesystem.
func chdirTempDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	return tmp
}

func TestInitFirstGeneration(t *testing.T) {
	initForce = false
	tmp := chdirTempDir(t)

	_, err := executeCommand("init", "--force")
	if err != nil {
		t.Fatalf("init --force failed: %v", err)
	}

	// AGENTS.md
	data, err := os.ReadFile(filepath.Join(tmp, "AGENTS.md"))
	if err != nil {
		t.Fatal("AGENTS.md not created:", err)
	}
	if len(data) == 0 {
		t.Error("AGENTS.md is empty")
	}

	// agents.yaml
	agentsPath := filepath.Join(tmp, ".forgecrew", "agents.yaml")
	agentsData, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatal("agents.yaml not created:", err)
	}
	if len(agentsData) == 0 {
		t.Error("agents.yaml is empty")
	}

	// models.yaml
	modelsPath := filepath.Join(tmp, ".forgecrew", "models.yaml")
	modelsData, err := os.ReadFile(modelsPath)
	if err != nil {
		t.Fatal("models.yaml not created:", err)
	}
	if len(modelsData) == 0 {
		t.Error("models.yaml is empty")
	}

	// workflows.yaml
	wfPath := filepath.Join(tmp, ".forgecrew", "workflows.yaml")
	wfData, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatal("workflows.yaml not created:", err)
	}
	if len(wfData) == 0 {
		t.Error("workflows.yaml is empty")
	}

	// settings.yaml
	settingsPath := filepath.Join(tmp, ".forgecrew", settings.SettingsYAML)
	settingsData, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal("settings.yaml not created:", err)
	}
	if len(settingsData) == 0 {
		t.Error("settings.yaml is empty")
	}

	// subdirectories
	for _, dir := range []string{"memory", "evals", "checkpoints"} {
		info, err := os.Stat(filepath.Join(tmp, ".forgecrew", dir))
		if err != nil || !info.IsDir() {
			t.Errorf(".forgecrew/%s/ directory not created", dir)
		}
	}
}

func TestInitIdempotent(t *testing.T) {
	initForce = false
	chdirTempDir(t)

	// First init.
	_, err := executeCommand("init")
	if err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	// Modify a generated file.
	modifiedContent := "<!-- modified comment -->\n"
	if err := os.WriteFile("AGENTS.md", []byte(modifiedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Second init without --force.
	_, err = executeCommand("init")
	if err != nil {
		t.Fatalf("second init failed: %v", err)
	}

	// AGENTS.md should still have our modified content.
	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != modifiedContent {
		t.Errorf("AGENTS.md was overwritten without --force.\n got: %q\nwant: %q", string(data), modifiedContent)
	}

	// settings.yaml should NOT be overwritten without --force either.
	settingsPath := filepath.Join(config.ConfigDir, settings.SettingsYAML)
	if !config.FileExists(settingsPath) {
		t.Error("settings.yaml should exist after first init")
	}
}

func TestInitForceOverwrite(t *testing.T) {
	initForce = false
	chdirTempDir(t)

	// First init.
	_, err := executeCommand("init")
	if err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	// Modify a generated file.
	modifiedContent := "<!-- modified comment -->\n"
	if err := os.WriteFile("AGENTS.md", []byte(modifiedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Second init with --force.
	out, err := executeCommand("init", "--force")
	if err != nil {
		t.Fatalf("init --force failed: %v", err)
	}

	// AGENTS.md should be overwritten back to the default template.
	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == modifiedContent {
		t.Error("AGENTS.md was NOT overwritten even with --force")
	}
	if !strings.Contains(string(data), "ForgeCrew") {
		t.Error("overwritten AGENTS.md should contain default template content")
	}

	// Output should NOT mention skipping for forced writes (zh: 跳过, en: skipping).
	if strings.Contains(out, "跳过") || strings.Contains(out, "skipping") {
		t.Error("output should not mention 'skipping'/'跳过' when --force is used")
	}
}
