package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/1424772/ForgeCrew/internal/i18n"
	"github.com/1424772/ForgeCrew/internal/settings"
)

// mockLineReader implements lineReader from a list of input lines.
type mockLineReader struct {
	lines []string
	pos   int
}

func (m *mockLineReader) ReadLine() (string, error) {
	if m.pos >= len(m.lines) {
		return "", io.EOF
	}
	line := m.lines[m.pos]
	m.pos++
	return line, nil
}

func newMockLineReader(input string) *mockLineReader {
	lines := strings.Split(input, "\n")
	return &mockLineReader{lines: lines}
}

// runInteractiveTest runs interactive mode with the given input in an
// isolated temp directory (to avoid settings.yaml pollution between tests).
func runInteractiveTest(t *testing.T, input string) string {
	t.Helper()
	chdirTempDir(t)
	var buf bytes.Buffer
	br := newMockLineReader(input)
	err := interactiveModeWithIO(br, &buf)
	if err != nil {
		t.Fatalf("interactive mode failed: %v", err)
	}
	return buf.String()
}

// ── Tests ──

func TestInteractiveHelp(t *testing.T) {
	out := runInteractiveTest(t, "/help\n/exit\n")
	if !strings.Contains(out, i18n.T("interactive.help", i18n.ZH)) {
		t.Error("output should contain help text")
	}
}

func TestInteractiveExit(t *testing.T) {
	out := runInteractiveTest(t, "/exit\n")
	if !strings.Contains(out, i18n.T("interactive.goodbye", i18n.ZH)) {
		t.Error("output should contain goodbye message")
	}
}

func TestInteractiveQuit(t *testing.T) {
	out := runInteractiveTest(t, "/quit\n")
	if !strings.Contains(out, i18n.T("interactive.goodbye", i18n.ZH)) {
		t.Error("output should contain goodbye on /quit")
	}
}

func TestInteractiveLangShow(t *testing.T) {
	out := runInteractiveTest(t, "/lang show\n/exit\n")
	if !strings.Contains(out, "zh") {
		t.Errorf("lang show should show zh (default), got: %s", out)
	}
}

func TestInteractiveLangSetZh(t *testing.T) {
	out := runInteractiveTest(t, "/lang set zh\n/lang show\n/exit\n")
	if !strings.Contains(out, "zh") {
		t.Errorf("output should contain zh, got: %s", out)
	}
}

func TestInteractiveLangSetEn(t *testing.T) {
	out := runInteractiveTest(t, "/lang set en\n/lang show\n/exit\n")
	// After setting en, the output should show English.
	if !strings.Contains(out, "Language set to:") {
		t.Errorf("output should show English set message, got: %s", out)
	}
	// The show after set should display en.
	if !strings.Contains(out, "en") {
		t.Errorf("lang show should display en after setting, got: %s", out)
	}
}

func TestInteractiveModeShow(t *testing.T) {
	out := runInteractiveTest(t, "/mode show\n/exit\n")
	if !strings.Contains(out, "plan") {
		t.Errorf("default mode should be plan, got: %s", out)
	}
}

func TestInteractiveModePlan(t *testing.T) {
	out := runInteractiveTest(t, "/mode plan\n/mode show\n/exit\n")
	if !strings.Contains(out, "plan") {
		t.Errorf("output should show plan mode, got: %s", out)
	}
}

func TestInteractiveModeAct(t *testing.T) {
	out := runInteractiveTest(t, "/mode act\n/mode show\n/exit\n")
	if !strings.Contains(out, "act") {
		t.Errorf("output should show act mode, got: %s", out)
	}
}

func TestInteractiveModeReview(t *testing.T) {
	out := runInteractiveTest(t, "/mode review\n/mode show\n/exit\n")
	if !strings.Contains(out, "review") {
		t.Errorf("output should show review mode, got: %s", out)
	}
}

func TestInteractiveUnknownCommand(t *testing.T) {
	out := runInteractiveTest(t, "/blah\n/exit\n")
	if !strings.Contains(out, "未知") || !strings.Contains(out, "/blah") {
		t.Errorf("output should show unknown command message, got: %s", out)
	}
}

func TestInteractiveScan(t *testing.T) {
	out := runInteractiveTest(t, "/scan\n/exit\n")
	if out == "" {
		t.Error("scan should produce output")
	}
	// Scanner output should contain project profile fields.
	if !strings.Contains(out, "project_type_guess") {
		t.Errorf("scan output should contain project_type_guess, got: %s", out)
	}
	if !strings.Contains(out, "root") {
		t.Errorf("scan output should contain root, got: %s", out)
	}
	if !strings.Contains(out, "languages") {
		t.Errorf("scan output should contain languages, got: %s", out)
	}
}

func TestInteractiveTeam(t *testing.T) {
	out := runInteractiveTest(t, "/team\n/exit\n")
	if out == "" {
		t.Error("team should produce output")
	}
	// Team output should contain YAML with team config fields.
	if !strings.Contains(out, "required_agents") {
		t.Errorf("team output should contain required_agents, got: %s", out)
	}
	if !strings.Contains(out, "project_type") {
		t.Errorf("team output should contain project_type, got: %s", out)
	}
}

func TestInteractiveValidate(t *testing.T) {
	// Set up valid agents.yaml and models.yaml for validation to pass.
	t.Helper()
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)

	agentsYAML := `
agents:
  test_agent:
    agent_id: test_agent
    name: Test Agent
    default_model: gpt_5_5
    permission_level: read_only
    tools:
      - read_file
`
	modelsYAML := `
models:
  gpt_5_5:
    provider: openai
    model: gpt-5.5
    api_key_env: OPENAI_API_KEY
    role: reasoning
    cost_tier: high
    supports_tools: true
    supports_vision: true
`
	os.WriteFile(filepath.Join(config.ConfigDir, config.AgentsYAML), []byte(agentsYAML), 0644)
	os.WriteFile(filepath.Join(config.ConfigDir, config.ModelsYAML), []byte(modelsYAML), 0644)

	var buf bytes.Buffer
	br := newMockLineReader("/validate\n/exit\n")
	err := interactiveModeWithIO(br, &buf)
	if err != nil {
		t.Fatalf("interactive mode failed: %v", err)
	}
	out := buf.String()

	if out == "" {
		t.Error("validate should produce output")
	}
	// Should contain validation passed messages (zh locale default).
	if !strings.Contains(out, "Agent-Model") || !strings.Contains(out, "校验通过") {
		t.Errorf("validate output should contain Agent-Model validation passed text, got: %s", out)
	}
	if !strings.Contains(out, "Agent-Tool") || !strings.Contains(out, "校验通过") {
		t.Errorf("validate output should contain Agent-Tool validation passed text, got: %s", out)
	}
	if !strings.Contains(out, "所有校验通过") {
		t.Errorf("validate output should contain all-ok message, got: %s", out)
	}
}

func TestInteractiveValidateMissingFiles(t *testing.T) {
	out := runInteractiveTest(t, "/validate\n/exit\n")
	if out == "" {
		t.Error("validate should produce output even when files missing")
	}
	// Should show a load error, not crash.
	if !strings.Contains(out, "加载") && !strings.Contains(out, "失败") {
		t.Errorf("validate should show load error for missing files, got: %s", out)
	}
}

func TestInteractiveLoadLocInvalidLanguage(t *testing.T) {
	// Write settings.yaml with invalid language (fr).
	t.Helper()
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)
	os.WriteFile(
		filepath.Join(config.ConfigDir, settings.SettingsYAML),
		[]byte("language: fr\n"),
		0644,
	)

	var buf bytes.Buffer
	br := newMockLineReader("/exit\n")
	err := interactiveModeWithIO(br, &buf)
	if err != nil {
		t.Fatalf("interactive mode failed: %v", err)
	}
	out := buf.String()

	// Should output warning about unsupported language.
	if !strings.Contains(out, "警告") || !strings.Contains(out, "语言设置无效") {
		t.Errorf("should warn about invalid language, got: %s", out)
	}
	// Should fallback to zh and still enter CLI.
	if !strings.Contains(out, "zh") {
		t.Errorf("should fallback to zh and show zh in banner, got: %s", out)
	}
	if !strings.Contains(out, i18n.T("interactive.goodbye", i18n.ZH)) {
		t.Errorf("should enter CLI with zh fallback, got: %s", out)
	}
}

func TestInteractivePlainTextTriggersDryRun(t *testing.T) {
	// Ensure zh locale for this test.
	t.Helper()
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)
	settings.Save(
		config.ConfigDir+"/"+settings.SettingsYAML,
		&settings.Settings{Language: "zh"},
	)

	var buf bytes.Buffer
	br := newMockLineReader("分析这个项目\n/exit\n")
	err := interactiveModeWithIO(br, &buf)
	if err != nil {
		t.Fatalf("interactive mode failed: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "分析这个项目") {
		t.Error("output should contain the task goal")
	}
	if !strings.Contains(out, "[演习]") {
		t.Error("output should contain [演习] (zh dry-run marker)")
	}
	if !strings.Contains(out, "第") || !strings.Contains(out, "轮") {
		t.Error("output should show iteration headers")
	}
}

func TestInteractivePlainTextNoWouldExecuteInZh(t *testing.T) {
	t.Helper()
	chdirTempDir(t)
	config.EnsureDir(config.ConfigDir)
	settings.Save(
		config.ConfigDir+"/"+settings.SettingsYAML,
		&settings.Settings{Language: "zh"},
	)

	var buf bytes.Buffer
	br := newMockLineReader("测试\n/exit\n")
	err := interactiveModeWithIO(br, &buf)
	if err != nil {
		t.Fatalf("interactive mode failed: %v", err)
	}
	out := buf.String()

	if strings.Contains(out, "would execute") {
		t.Error("zh output should not contain 'would execute'")
	}
}

func TestInteractiveEmptyInput(t *testing.T) {
	out := runInteractiveTest(t, "\n\n/exit\n")
	if !strings.Contains(out, i18n.T("interactive.goodbye", i18n.ZH)) {
		t.Error("output should contain goodbye after empty lines")
	}
}

func TestInteractiveBanner(t *testing.T) {
	out := runInteractiveTest(t, "/exit\n")
	if !strings.Contains(out, Version) {
		t.Error("banner should contain version")
	}
	if !strings.Contains(out, "语言") {
		t.Error("banner should show language (zh default)")
	}
	if !strings.Contains(out, "模式") {
		t.Error("banner should show mode")
	}
}

func TestInteractiveReviewMode(t *testing.T) {
	out := runInteractiveTest(t, "/mode review\n分析一下\n/exit\n")
	if !strings.Contains(out, i18n.T("interactive.review_disabled", i18n.ZH)) {
		t.Error("review mode should show disabled notice, got: " + out)
	}
}

func TestInteractiveActMode(t *testing.T) {
	out := runInteractiveTest(t, "/mode act\n写代码\n/exit\n")
	if !strings.Contains(out, i18n.T("interactive.act_disabled", i18n.ZH)) {
		t.Error("act mode should show disabled notice, got: " + out)
	}
}
