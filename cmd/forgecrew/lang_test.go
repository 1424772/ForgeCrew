package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/1424772/ForgeCrew/internal/settings"
	"gopkg.in/yaml.v3"
)

// executeLangCmd creates a fresh lang command and runs it.
func executeLangCmd(args ...string) (string, error) {
	cmd := langCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestLangShowDefault(t *testing.T) {
	chdirTempDir(t)

	// Without settings.yaml, should default to zh.
	out, err := executeLangCmd("show")
	if err != nil {
		t.Fatalf("lang show failed: %v", err)
	}
	if !strings.Contains(out, "zh") {
		t.Errorf("default lang should be zh, got: %s", out)
	}
}

func TestLangSetZh(t *testing.T) {
	tmp := chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	out, err := executeLangCmd("set", "zh")
	if err != nil {
		t.Fatalf("lang set zh failed: %v", err)
	}
	if !strings.Contains(out, "zh") {
		t.Errorf("output should contain zh, got: %s", out)
	}

	// Verify settings.yaml was written.
	verifySettingsFile(t, tmp, "zh")
}

func TestLangSetEn(t *testing.T) {
	tmp := chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	out, err := executeLangCmd("set", "en")
	if err != nil {
		t.Fatalf("lang set en failed: %v", err)
	}
	if !strings.Contains(out, "en") {
		t.Errorf("output should contain en, got: %s", out)
	}

	verifySettingsFile(t, tmp, "en")
}

func TestLangSetInvalid(t *testing.T) {
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	_, err := executeLangCmd("set", "fr")
	if err == nil {
		t.Error("lang set fr should fail")
	}
}

func TestLangShowAfterSet(t *testing.T) {
	tmp := chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	// Set to en.
	_, err := executeLangCmd("set", "en")
	if err != nil {
		t.Fatalf("lang set en failed: %v", err)
	}

	// Show should display en.
	out, err := executeLangCmd("show")
	if err != nil {
		t.Fatalf("lang show failed: %v", err)
	}
	if !strings.Contains(out, "en") {
		t.Errorf("show after set en should show en, got: %s", out)
	}
	_ = tmp
}

func TestLangList(t *testing.T) {
	chdirTempDir(t)

	out, err := executeLangCmd("list")
	if err != nil {
		t.Fatalf("lang list failed: %v", err)
	}
	if !strings.Contains(out, "zh") || !strings.Contains(out, "en") {
		t.Errorf("lang list should contain zh and en, got: %s", out)
	}
}

func TestLangAlias(t *testing.T) {
	// "language" should work as alias for "lang".
	rootCmd.SetArgs([]string{"language", "show"})
	// Just verify the alias is registered.
	cmd, _, _ := rootCmd.Find([]string{"language"})
	if cmd == nil {
		t.Error("'language' alias should be registered")
	}
}

func TestInitGeneratesSettingsYAML(t *testing.T) {
	initForce = false
	tmp := chdirTempDir(t)

	_, err := executeCommand("init")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	settingsPath := filepath.Join(tmp, config.ConfigDir, settings.SettingsYAML)
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("settings.yaml not created by init: %v", err)
	}

	var s settings.Settings
	if err := yaml.Unmarshal(data, &s); err != nil {
		t.Fatalf("invalid settings.yaml: %v", err)
	}
	if s.Language != "zh" {
		t.Errorf("default language should be zh, got: %s", s.Language)
	}
}

func TestTaskZhOutput(t *testing.T) {
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	// Set language to zh.
	settingsPath := filepath.Join(config.ConfigDir, settings.SettingsYAML)
	settings.Save(settingsPath, &settings.Settings{Language: "zh"})

	out, err := executeTaskCmd("测试任务", "--dry-run")
	if err != nil {
		t.Fatalf("task --dry-run failed: %v", err)
	}
	if !strings.Contains(out, "任务") {
		t.Error("zh output should contain 任务")
	}
	if !strings.Contains(out, "[演习]") {
		t.Error("zh output should contain [演习]")
	}
	if !strings.Contains(out, "第") {
		t.Error("zh output should contain 第")
	}
	if strings.Contains(out, "would execute") {
		t.Error("zh output should not contain 'would execute'")
	}
}

func TestTaskEnOutput(t *testing.T) {
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	// Set language to en.
	settingsPath := filepath.Join(config.ConfigDir, settings.SettingsYAML)
	settings.Save(settingsPath, &settings.Settings{Language: "en"})

	out, err := executeTaskCmd("test task", "--dry-run")
	if err != nil {
		t.Fatalf("task --dry-run failed: %v", err)
	}
	if !strings.Contains(out, "Task") {
		t.Error("en output should contain Task")
	}
	if !strings.Contains(out, "[dry-run]") {
		t.Error("en output should contain [dry-run]")
	}
	if !strings.Contains(out, "Iteration") {
		t.Error("en output should contain Iteration")
	}
}

func TestLangShowInvalidInSettings(t *testing.T) {
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	// Write settings.yaml with an invalid language value.
	settingsPath := filepath.Join(config.ConfigDir, settings.SettingsYAML)
	os.WriteFile(settingsPath, []byte("language: fr\n"), 0644)

	_, err := executeLangCmd("show")
	if err == nil {
		t.Fatal("lang show should fail when settings.yaml has invalid language")
	}
}

func TestSettingsLoadRejectsInvalid(t *testing.T) {
	tmp := chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	settingsPath := filepath.Join(tmp, config.ConfigDir, settings.SettingsYAML)
	os.WriteFile(settingsPath, []byte("language: fr\n"), 0644)

	_, err := settings.Load(settingsPath)
	if err == nil {
		t.Fatal("settings.Load should reject invalid language")
	}
	if !strings.Contains(err.Error(), "fr") {
		t.Errorf("error should mention the invalid value fr, got: %v", err)
	}
}

func verifySettingsFile(t *testing.T, tmp, expectedLang string) {
	t.Helper()
	settingsPath := filepath.Join(tmp, config.ConfigDir, settings.SettingsYAML)
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("settings.yaml not found: %v", err)
	}
	var s settings.Settings
	if err := yaml.Unmarshal(data, &s); err != nil {
		t.Fatalf("invalid settings.yaml: %v", err)
	}
	if s.Language != expectedLang {
		t.Errorf("settings.yaml language = %q, want %q", s.Language, expectedLang)
	}
}
