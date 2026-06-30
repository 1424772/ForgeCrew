package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	s := Default()
	if s.Language != "zh" {
		t.Errorf("Default language = %q, want zh", s.Language)
	}
}

func TestLoadValidZH(t *testing.T) {
	path := writeTemp(t, "language: zh\n")
	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Language != "zh" {
		t.Errorf("language = %q, want zh", s.Language)
	}
}

func TestLoadValidEN(t *testing.T) {
	path := writeTemp(t, "language: en\n")
	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Language != "en" {
		t.Errorf("language = %q, want en", s.Language)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.yaml")
	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Language != "zh" {
		t.Errorf("language = %q, want zh (default)", s.Language)
	}
}

func TestLoadEmptyLanguageDefaultsToZH(t *testing.T) {
	path := writeTemp(t, "language: \"\"\n")
	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Language != "zh" {
		t.Errorf("language = %q, want zh", s.Language)
	}
}

func TestLoadMissingLanguageFieldDefaultsToZH(t *testing.T) {
	path := writeTemp(t, "{}\n")
	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Language != "zh" {
		t.Errorf("language = %q, want zh", s.Language)
	}
}

func TestLoadInvalidLanguage(t *testing.T) {
	path := writeTemp(t, "language: fr\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unsupported language 'fr'")
	}
}

func TestLoadCorruptedYAML(t *testing.T) {
	path := writeTemp(t, "language: [unclosed\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected parse error for corrupted YAML")
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.yaml")

	s := &Settings{Language: "en"}
	if err := Save(path, s); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Language != "en" {
		t.Errorf("language = %q, want en", loaded.Language)
	}
}

func TestSaveDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.yaml")

	s := Default()
	if err := Save(path, s); err != nil {
		t.Fatal(err)
	}

	// Verify file exists and is readable.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("saved file is empty")
	}
}

func TestLoadWithIndent(t *testing.T) {
	path := writeTemp(t, "language: zh\n# comment\n")
	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Language != "zh" {
		t.Errorf("language = %q, want zh", s.Language)
	}
}

// writeTemp writes YAML content to a temp file and returns the path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "settings.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
