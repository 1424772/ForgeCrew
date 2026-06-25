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

	// Create a tests directory with a test file inside.
	testsDir := filepath.Join(tmp, "tests")
	os.MkdirAll(testsDir, 0755)
	os.WriteFile(filepath.Join(testsDir, "test_example.py"), []byte("def test(): pass"), 0644)

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

func TestDetectTestsRecursive(t *testing.T) {
	tmp := t.TempDir()

	// Deeply nested test file.
	nestedDir := filepath.Join(tmp, "pkg", "service", "auth")
	os.MkdirAll(nestedDir, 0755)
	os.WriteFile(filepath.Join(nestedDir, "auth_test.go"), []byte("package auth"), 0644)

	s := New()
	p, err := s.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if !p.HasTests {
		t.Error("should detect deeply nested _test.go files")
	}
}

func TestDetectTestsSkipIgnored(t *testing.T) {
	tmp := t.TempDir()

	// Test file inside .git should be skipped.
	gitDir := filepath.Join(tmp, ".git")
	os.MkdirAll(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "fake_test.go"), []byte("package git"), 0644)

	// Test file inside vendor should be skipped.
	vendorDir := filepath.Join(tmp, "vendor", "pkg")
	os.MkdirAll(vendorDir, 0755)
	os.WriteFile(filepath.Join(vendorDir, "vendor_test.go"), []byte("package vendor"), 0644)

	// Test file inside node_modules should be skipped.
	nmDir := filepath.Join(tmp, "node_modules", "lib")
	os.MkdirAll(nmDir, 0755)
	os.WriteFile(filepath.Join(nmDir, "lib.test.ts"), []byte("test"), 0644)

	s := New()
	p, err := s.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if p.HasTests {
		t.Error("should NOT detect tests from .git, vendor, or node_modules")
	}
}

func TestDetectTestsDirectoryMatch(t *testing.T) {
	tmp := t.TempDir()

	// __tests__ directory with a JS file.
	testsDir := filepath.Join(tmp, "__tests__")
	os.MkdirAll(testsDir, 0755)
	os.WriteFile(filepath.Join(testsDir, "helper.js"), []byte("// test helper"), 0644)

	s := New()
	p, err := s.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if !p.HasTests {
		t.Error("should detect tests from __tests__ directory")
	}
}

func TestDetectTestsSpecFile(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "button.spec.ts"), []byte("describe('button')"), 0644)

	s := New()
	p, err := s.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if !p.HasTests {
		t.Error("should detect .spec.ts files")
	}
}
