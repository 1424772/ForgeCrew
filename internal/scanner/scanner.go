// Package scanner provides project scanning capabilities to identify
// languages, frameworks, tools, and generate a project profile.
package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Profile represents the result of a project scan.
type Profile struct {
	Root             string   `json:"root" yaml:"root"`
	Languages        []string `json:"languages" yaml:"languages"`
	Frameworks       []string `json:"frameworks" yaml:"frameworks"`
	HasTests         bool     `json:"has_tests" yaml:"has_tests"`
	HasDocker        bool     `json:"has_docker" yaml:"has_docker"`
	HasAgentsMD      bool     `json:"has_agents_md" yaml:"has_agents_md"`
	HasGitHubActions bool     `json:"has_github_actions" yaml:"has_github_actions"`
	ProjectTypeGuess string   `json:"project_type_guess" yaml:"project_type_guess"`
}

// Scanner performs project scanning.
type Scanner struct{}

// New creates a new Scanner.
func New() *Scanner {
	return &Scanner{}
}

// Scan analyzes the given root directory and returns a project profile.
func (s *Scanner) Scan(root string) (*Profile, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root: %w", err)
	}

	p := &Profile{
		Root: absRoot,
	}

	// Detect languages and frameworks from config files.
	s.detectFromFiles(absRoot, p)

	// Check for tests directory patterns.
	s.detectTests(absRoot, p)

	// Check for Docker.
	s.detectDocker(absRoot, p)

	// Check for AGENTS.md.
	s.detectAgentsMD(absRoot, p)

	// Check for GitHub Actions.
	s.detectGitHubActions(absRoot, p)

	// Guess project type.
	p.ProjectTypeGuess = guessProjectType(p)

	return p, nil
}

func (s *Scanner) detectFromFiles(root string, p *Profile) {
	indicators := map[string]struct {
		lang      string
		framework string
	}{
		"go.mod":           {lang: "go", framework: "go modules"},
		"go.sum":           {lang: "go", framework: ""},
		"package.json":     {lang: "javascript/typescript", framework: "node"},
		"tsconfig.json":    {lang: "typescript", framework: ""},
		"pyproject.toml":   {lang: "python", framework: ""},
		"setup.py":         {lang: "python", framework: ""},
		"requirements.txt": {lang: "python", framework: ""},
		"Pipfile":          {lang: "python", framework: ""},
		"Cargo.toml":       {lang: "rust", framework: "cargo"},
		"pom.xml":          {lang: "java", framework: "maven"},
		"build.gradle":     {lang: "java/kotlin", framework: "gradle"},
		"Gemfile":          {lang: "ruby", framework: "bundler"},
		"CMakeLists.txt":   {lang: "c/c++", framework: "cmake"},
		"Makefile":         {lang: "", framework: "make"},
	}

	seenLangs := map[string]bool{}
	seenFrameworks := map[string]bool{}

	for file, info := range indicators {
		if fileExists(filepath.Join(root, file)) {
			if info.lang != "" && !seenLangs[info.lang] {
				p.Languages = append(p.Languages, info.lang)
				seenLangs[info.lang] = true
			}
			if info.framework != "" && !seenFrameworks[info.framework] {
				p.Frameworks = append(p.Frameworks, info.framework)
				seenFrameworks[info.framework] = true
			}
		}
	}

	// Detect more frameworks from go.mod content.
	if fileExists(filepath.Join(root, "go.mod")) {
		data, err := os.ReadFile(filepath.Join(root, "go.mod"))
		if err == nil {
			content := string(data)
			if strings.Contains(content, "cobra") && !seenFrameworks["cobra"] {
				p.Frameworks = append(p.Frameworks, "cobra")
				seenFrameworks["cobra"] = true
			}
			if strings.Contains(content, "gin-gonic") && !seenFrameworks["gin"] {
				p.Frameworks = append(p.Frameworks, "gin")
				seenFrameworks["gin"] = true
			}
			if strings.Contains(content, "fiber") && !seenFrameworks["fiber"] {
				p.Frameworks = append(p.Frameworks, "fiber")
				seenFrameworks["fiber"] = true
			}
		}
	}
}

func (s *Scanner) detectTests(root string, p *Profile) {
	testPatterns := []string{
		"*_test.go", "*.test.ts", "*.test.js", "*.test.tsx", "*.spec.ts",
		"*.spec.js", "test_*.py", "*_test.py", "tests/", "__tests__/",
		"src/test/", "spec/",
	}
	for _, pattern := range testPatterns {
		matches, _ := filepath.Glob(filepath.Join(root, pattern))
		if len(matches) > 0 {
			p.HasTests = true
			return
		}
		// Check if directory exists
		if strings.HasSuffix(pattern, "/") {
			if info, err := os.Stat(filepath.Join(root, pattern)); err == nil && info.IsDir() {
				p.HasTests = true
				return
			}
		}
	}
}

func (s *Scanner) detectDocker(root string, p *Profile) {
	p.HasDocker = fileExists(filepath.Join(root, "Dockerfile")) ||
		fileExists(filepath.Join(root, "docker-compose.yml")) ||
		fileExists(filepath.Join(root, "docker-compose.yaml")) ||
		fileExists(filepath.Join(root, ".dockerignore"))
}

func (s *Scanner) detectAgentsMD(root string, p *Profile) {
	p.HasAgentsMD = fileExists(filepath.Join(root, "AGENTS.md"))
}

func (s *Scanner) detectGitHubActions(root string, p *Profile) {
	ghaDir := filepath.Join(root, ".github", "workflows")
	if info, err := os.Stat(ghaDir); err == nil && info.IsDir() {
		p.HasGitHubActions = true
	}
}

func guessProjectType(p *Profile) string {
	hasGo := contains(p.Languages, "go")
	hasTS := contains(p.Languages, "typescript") || contains(p.Languages, "javascript/typescript")
	hasPython := contains(p.Languages, "python")
	hasDocker := p.HasDocker
	hasTests := p.HasTests

	// CLI tool detection.
	if hasGo && contains(p.Frameworks, "cobra") {
		return "cli_tool"
	}

	// Backend API detection.
	if hasGo && (contains(p.Frameworks, "gin") || contains(p.Frameworks, "fiber")) {
		return "backend_api"
	}

	// Frontend detection.
	if hasTS && !hasGo && !hasPython {
		return "frontend_app"
	}

	// Fullstack detection.
	if hasTS && (hasGo || hasPython) {
		return "fullstack_saas"
	}

	// Agent/AI app detection.
	if hasPython && hasTests {
		return "agent_app"
	}

	// Docker-only.
	if hasDocker && len(p.Languages) == 0 {
		return "devops"
	}

	if len(p.Languages) > 0 {
		return "generic"
	}
	return "unknown"
}

// Format outputs the profile in the requested format.
func (p *Profile) Format(format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		return p.toJSON()
	case "yaml", "yml":
		return p.toYAML()
	default:
		return "", fmt.Errorf("unsupported format: %s (use yaml or json)", format)
	}
}

func (p *Profile) toYAML() (string, error) {
	out, err := yaml.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (p *Profile) toJSON() (string, error) {
	out, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out) + "\n", nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
