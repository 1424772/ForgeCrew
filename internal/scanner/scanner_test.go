package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanGoProject(t *testing.T) {
	tmp := t.TempDir()

	// Create a simulated Go CLI project.
	os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example\nrequire github.com/spf13/cobra"), 0644)
	os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmp, "main_test.go"), []byte("package main\nimport \"testing\""), 0644)
	os.WriteFile(filepath.Join(tmp, "AGENTS.md"), []byte("# AGENTS.md"), 0644)

	s := New()
	p, err := s.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if !contains(p.Languages, "go") {
		t.Error("should detect go language")
	}
	if !contains(p.Frameworks, "go modules") {
		t.Error("should detect go modules")
	}
	if !contains(p.Frameworks, "cobra") {
		t.Error("should detect cobra framework")
	}
	if !p.HasTests {
		t.Error("should detect tests")
	}
	if !p.HasAgentsMD {
		t.Error("should detect AGENTS.md")
	}
	if p.ProjectTypeGuess != "cli_tool" {
		t.Errorf("project_type_guess = %q, want cli_tool", p.ProjectTypeGuess)
	}
}

func TestScanEmptyProject(t *testing.T) {
	tmp := t.TempDir()
	s := New()
	p, err := s.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Languages) != 0 {
		t.Error("empty project should have no languages")
	}
	if p.ProjectTypeGuess != "unknown" {
		t.Errorf("project_type_guess = %q, want unknown", p.ProjectTypeGuess)
	}
}

func TestScanWithDocker(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "Dockerfile"), []byte("FROM scratch"), 0644)

	s := New()
	p, err := s.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if !p.HasDocker {
		t.Error("should detect Dockerfile")
	}
	if p.ProjectTypeGuess != "devops" {
		t.Errorf("project_type_guess = %q, want devops", p.ProjectTypeGuess)
	}
}

func TestFormatYAML(t *testing.T) {
	p := &Profile{
		Root:             "/test",
		Languages:        []string{"go"},
		Frameworks:       []string{"cobra"},
		HasTests:         true,
		ProjectTypeGuess: "cli_tool",
	}
	out, err := p.Format("yaml")
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Error("YAML output should not be empty")
	}
}

func TestFormatJSON(t *testing.T) {
	p := &Profile{
		Root:             "/test",
		Languages:        []string{"go"},
		ProjectTypeGuess: "cli_tool",
	}
	out, err := p.Format("json")
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Error("JSON output should not be empty")
	}
}

func TestFormatInvalid(t *testing.T) {
	p := &Profile{}
	_, err := p.Format("xml")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestDetectPython(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "pyproject.toml"), []byte("[tool.poetry]"), 0644)
	os.WriteFile(filepath.Join(tmp, "tests"), nil, 0755)
	os.MkdirAll(filepath.Join(tmp, "tests"), 0755)

	s := New()
	p, err := s.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if !contains(p.Languages, "python") {
		t.Error("should detect python")
	}
	if !p.HasTests {
		t.Error("should detect tests directory")
	}
}
