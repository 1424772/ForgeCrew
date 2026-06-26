package main

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/1424772/ForgeCrew/internal/agents"
	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/1424772/ForgeCrew/internal/i18n"
	"github.com/1424772/ForgeCrew/internal/models"
	"github.com/1424772/ForgeCrew/internal/orchestrator"
	"github.com/1424772/ForgeCrew/internal/scanner"
	"github.com/1424772/ForgeCrew/internal/settings"
	"github.com/1424772/ForgeCrew/internal/teamarchitect"
)

// interactiveMode runs the interactive CLI loop using os.Stdin / os.Stdout.
func interactiveMode() error {
	return interactiveModeWithIO(
		reader{r: bufio.NewReader(osReadCloser{})},
		osStdoutWriter{},
	)
}

// interactiveModeWithIO runs the interactive CLI loop with injectable IO.
// r provides lines of input; w receives all output.
func interactiveModeWithIO(r lineReader, w io.Writer) error {
	locale, loc := loadLoc(w)

	// ── Startup banner ──
	printBanner(w, loc)

	// ── Warn if not initialized ──
	if !config.FileExists(filepath.Join(config.ConfigDir, config.AgentsYAML)) {
		fmt.Fprintln(w, i18n.T("interactive.no_init", loc))
	}

	// ── Hint to use /help ──
	fmt.Fprintln(w, i18n.T("interactive.start_hint", loc))

	// ── Dry-run warning ──
	fmt.Fprintln(w, i18n.T("interactive.dry_run_warning", loc))
	fmt.Fprintln(w)

	// ── Current mode ──
	mode := "plan"

	// ── Read loop ──
	fmt.Fprint(w, i18n.T("interactive.prompt", loc))
	for {
		line, err := r.ReadLine()
		if err == io.EOF {
			fmt.Fprintln(w)
			fmt.Fprintln(w, i18n.T("interactive.goodbye", loc))
			return nil
		}
		if err != nil {
			return fmt.Errorf("interactive input: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			fmt.Fprint(w, i18n.T("interactive.prompt", loc))
			continue
		}

		// Refresh locale (may have changed via /lang set).
		if _, newLoc := loadLocQuiet(); newLoc != loc {
			loc = newLoc
		}
		_ = locale

		switch {
		case line == "/exit" || line == "/quit":
			fmt.Fprintln(w, i18n.T("interactive.goodbye", loc))
			return nil

		case line == "/help":
			fmt.Fprintln(w, i18n.T("interactive.help", loc))

		case line == "/lang show":
			fmt.Fprintf(w, "%s%s\n", i18n.T("lang.current", loc), string(loc))

		case strings.HasPrefix(line, "/lang set "):
			lang := strings.TrimPrefix(line, "/lang set ")
			lang = strings.ToLower(strings.TrimSpace(lang))
			newLoc, err := i18n.ValidLocale(lang)
			if err != nil {
				fmt.Fprintln(w, err)
			} else {
				settingsPath := filepath.Join(config.ConfigDir, settings.SettingsYAML)
				if err := config.EnsureDir(config.ConfigDir); err != nil {
					fmt.Fprintf(w, "保存语言设置失败: %v\n", err)
				} else if err := settings.Save(settingsPath, &settings.Settings{Language: string(newLoc)}); err != nil {
					fmt.Fprintf(w, "保存语言设置失败: %v\n", err)
				} else {
					loc = newLoc
					fmt.Fprintf(w, "%s%s\n", i18n.T("lang.set_to", loc), string(loc))
				}
			}

		case line == "/mode show":
			fmt.Fprintln(w, i18n.T("interactive.mode", loc)+mode)

		case line == "/mode plan":
			mode = "plan"
			fmt.Fprintln(w, i18n.T("interactive.mode", loc)+"plan")

		case line == "/mode act":
			mode = "act"
			fmt.Fprintln(w, i18n.T("interactive.mode", loc)+"act")

		case line == "/mode review":
			mode = "review"
			fmt.Fprintln(w, i18n.T("interactive.mode", loc)+"review")

		case line == "/scan":
			runInteractiveScan(w)

		case line == "/team":
			runInteractiveTeam(w)

		case line == "/validate":
			runInteractiveValidate(w, loc)

		case strings.HasPrefix(line, "/"):
			// Unknown slash command.
			fmt.Fprintf(w, i18n.T("interactive.unknown_cmd", loc), line)

		default:
			// Plain text → dry-run task with mode-specific behavior.
			runInteractiveTask(w, line, mode, loc)
		}

		fmt.Fprint(w, i18n.T("interactive.prompt", loc))
	}
}

// ── IO abstractions for injectable testing ──

// lineReader abstracts reading lines of input.
type lineReader interface {
	ReadLine() (string, error)
}

// reader wraps a bufio.Reader to implement lineReader.
type reader struct {
	r *bufio.Reader
}

func (r reader) ReadLine() (string, error) {
	line, err := r.r.ReadString('\n')
	return strings.TrimSuffix(line, "\n"), err
}

// osReadCloser is an io.Reader that never actually closes — it's just
// a thin wrapper so os.Stdin satisfies io.Reader in our reader wrapper.
type osReadCloser struct{}

func (osReadCloser) Read(p []byte) (int, error) { return osStdin.Read(p) }

// osStdoutWriter writes to os.Stdout.
type osStdoutWriter struct{}

func (osStdoutWriter) Write(p []byte) (int, error) { return osStdout.Write(p) }

// ── loadLoc ──

// loadLoc loads the current locale from settings. If the file is missing,
// it silently defaults to zh. If the language value is invalid, it prints
// a warning to w and falls back to zh (so the user can still enter the CLI).
func loadLoc(w io.Writer) (string, i18n.Locale) {
	s, err := settings.Load(filepath.Join(config.ConfigDir, settings.SettingsYAML))
	if err != nil {
		// Only warn if the file exists but has an invalid value.
		// Missing file is normal (user hasn't run init yet) — silent default.
		if config.FileExists(filepath.Join(config.ConfigDir, settings.SettingsYAML)) {
			fmt.Fprintf(w, "警告: 语言设置无效 (%v)，已回退为 zh。\n", err)
		}
		return "zh", i18n.ZH
	}
	return s.Language, i18n.Locale(s.Language)
}

// loadLocQuiet loads the locale silently, defaulting to zh on any error.
// Used for in-loop refresh where warnings would be noisy.
func loadLocQuiet() (string, i18n.Locale) {
	s, err := settings.Load(filepath.Join(config.ConfigDir, settings.SettingsYAML))
	if err != nil {
		return "zh", i18n.ZH
	}
	return s.Language, i18n.Locale(s.Language)
}

// ── printBanner ──

func printBanner(w io.Writer, loc i18n.Locale) {
	fmt.Fprintf(w, "\n%s\n", fmt.Sprintf(i18n.T("interactive.banner", loc), Version))

	cwd, _ := osGetwd()
	fmt.Fprintf(w, "%s\n", fmt.Sprintf(i18n.T("interactive.project", loc), cwd))
	fmt.Fprintf(w, "%s\n", fmt.Sprintf(i18n.T("interactive.language", loc), string(loc)))
	fmt.Fprintln(w, i18n.T("interactive.mode", loc)+"plan")
}

// ── runInteractiveTask ──

func runInteractiveTask(w io.Writer, goal string, mode string, loc i18n.Locale) {
	switch mode {
	case "plan":
		// Full dry-run through orchestrator.
		fmt.Fprintf(w, "\n%s: %s\n", i18n.T("interactive.mode", loc)+"plan", goal)
		sm := orchestrator.New(3, true)
		result := sm.RunFull(goal)
		fmt.Fprint(w, result.FormatText(string(loc)))
		fmt.Fprintln(w)

	case "review":
		// Review mode: not yet connected to real reviewer.
		fmt.Fprintln(w)
		fmt.Fprintln(w, i18n.T("interactive.review_disabled", loc))

	case "act":
		// Act mode: still dry-run in this phase.
		fmt.Fprintln(w)
		fmt.Fprintln(w, i18n.T("interactive.act_disabled", loc))

	default:
		// Fallback: same as plan.
		sm := orchestrator.New(3, true)
		result := sm.RunFull(goal)
		fmt.Fprint(w, result.FormatText(string(loc)))
		fmt.Fprintln(w)
	}
}

// ── runInteractiveScan ──

func runInteractiveScan(w io.Writer) {
	s := scanner.New()
	profile, err := s.Scan(".")
	if err != nil {
		fmt.Fprintf(w, "扫描失败: %v\n", err)
		return
	}
	yamlOut, err := profile.Format("yaml")
	if err != nil {
		fmt.Fprintf(w, "YAML 序列化失败: %v\n", err)
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, yamlOut)
}

// ── runInteractiveTeam ──

func runInteractiveTeam(w io.Writer) {
	s := scanner.New()
	profile, err := s.Scan(".")
	if err != nil {
		fmt.Fprintf(w, "扫描失败: %v\n", err)
		return
	}
	ta := teamarchitect.New()
	tc := ta.Suggest(profile)
	yamlOut, err := tc.FormatYAML()
	if err != nil {
		fmt.Fprintf(w, "YAML 序列化失败: %v\n", err)
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, yamlOut)
}

// ── runInteractiveValidate ──

func runInteractiveValidate(w io.Writer, loc i18n.Locale) {
	// Load agents.
	agentReg := agents.NewRegistry()
	agentsPath := filepath.Join(config.ConfigDir, config.AgentsYAML)
	if err := agentReg.Load(agentsPath); err != nil {
		fmt.Fprintf(w, "加载 agents 失败: %v\n", err)
		return
	}

	// Load models.
	modelReg := models.NewRegistry()
	modelsPath := filepath.Join(config.ConfigDir, config.ModelsYAML)
	if err := modelReg.Load(modelsPath); err != nil {
		fmt.Fprintf(w, "加载 models 失败: %v\n", err)
		return
	}

	// Build known tool names.
	knownTools := buildKnownToolNames()

	fmt.Fprintln(w)
	hadError := false

	// Validate model refs.
	if err := agentReg.ValidateModelRefs(modelReg); err != nil {
		fmt.Fprintf(w, "%v\n", err)
		hadError = true
	} else {
		fmt.Fprintln(w, i18n.T("validate.model_ok", loc))
	}

	// Validate tools.
	if err := agentReg.ValidateTools(knownTools); err != nil {
		fmt.Fprintf(w, "%v\n", err)
		hadError = true
	} else {
		fmt.Fprintln(w, i18n.T("validate.tool_ok", loc))
	}

	if hadError {
		fmt.Fprintln(w, "校验存在错误。")
	} else {
		fmt.Fprintln(w, i18n.T("validate.all_ok", loc))
	}
}
