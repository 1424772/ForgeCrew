package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/1424772/ForgeCrew/internal/i18n"
	"github.com/1424772/ForgeCrew/internal/orchestrator"
	"github.com/1424772/ForgeCrew/internal/settings"
)

// interactiveMode runs the interactive CLI loop.
func interactiveMode() error {
	locale, loc := loadLoc()
	w := os.Stdout

	// ── Startup banner ──
	printBanner(w, loc)

	// ── Warn if not initialized ──
	if !config.FileExists(filepath.Join(config.ConfigDir, config.AgentsYAML)) {
		fmt.Fprintln(w, i18n.T("interactive.no_init", loc))
		fmt.Fprintln(w)
	}

	// ── Dry-run warning ──
	fmt.Fprintln(w, i18n.T("interactive.dry_run_warning", loc))
	fmt.Fprintln(w)

	// ── Current mode ──
	mode := "plan"

	// ── Read loop ──
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Fprint(w, i18n.T("interactive.prompt", loc))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			fmt.Fprint(w, i18n.T("interactive.prompt", loc))
			continue
		}

		// Refresh locale (may have changed via /lang set).
		if _, newLoc := loadLoc(); newLoc != loc {
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
			fmt.Fprintf(w, "%s%s\n", i18n.T("lang.current", loc), loc)

		case strings.HasPrefix(line, "/lang set "):
			lang := strings.TrimPrefix(line, "/lang set ")
			lang = strings.ToLower(strings.TrimSpace(lang))
			newLoc, err := i18n.ValidLocale(lang)
			if err != nil {
				fmt.Fprintln(w, err)
			} else {
				loc = newLoc
				settingsPath := filepath.Join(config.ConfigDir, settings.SettingsYAML)
				config.EnsureDir(config.ConfigDir)
				settings.Save(settingsPath, &settings.Settings{Language: string(loc)})
				fmt.Fprintf(w, "%s%s\n", i18n.T("lang.set_to", loc), string(loc))
			}

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

		default:
			// Plain text → dry-run task.
			runInteractiveTask(w, line, mode, loc)
		}

		fmt.Fprint(w, i18n.T("interactive.prompt", loc))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("interactive input: %w", err)
	}
	return nil
}

// loadLoc loads the current locale, defaulting to zh.
func loadLoc() (string, i18n.Locale) {
	locale := "zh"
	s, err := settings.Load(filepath.Join(config.ConfigDir, settings.SettingsYAML))
	if err == nil {
		locale = s.Language
	}
	return locale, i18n.Locale(locale)
}

// printBanner prints the startup banner to w.
func printBanner(w *os.File, loc i18n.Locale) {
	// Version.
	fmt.Fprintf(w, "\n%s\n", fmt.Sprintf(i18n.T("interactive.banner", loc), Version))
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", 50))

	// Project path.
	cwd, _ := os.Getwd()
	fmt.Fprintf(w, "%s\n", fmt.Sprintf(i18n.T("interactive.project", loc), cwd))

	// Language.
	fmt.Fprintf(w, "%s\n", fmt.Sprintf(i18n.T("interactive.language", loc), string(loc)))

	// Mode (always starts in plan).
	fmt.Fprintln(w, i18n.T("interactive.mode", loc)+"plan")
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", 50))
}

// runInteractiveTask runs a dry-run task from interactive input.
func runInteractiveTask(w *os.File, goal string, mode string, loc i18n.Locale) {
	_ = mode
	// Show plan mode note.
	fmt.Fprintf(w, "\n%s\n\n", i18n.T("interactive.plan_mode_note", loc))

	// Run through orchestrator.
	sm := orchestrator.New(3, true)
	result := sm.RunFull(goal)

	fmt.Fprint(w, result.FormatText(string(loc)))
	fmt.Fprintln(w)
}

// runInteractiveScan runs a project scan from interactive mode.
func runInteractiveScan(w *os.File) {
	// Delegate to scan command logic — just print a message for now.
	fmt.Fprintln(w, "  Run 'forgecrew scan' for project profiling.")
}

// runInteractiveTeam runs team suggestion from interactive mode.
func runInteractiveTeam(w *os.File) {
	fmt.Fprintln(w, "  Run 'forgecrew team suggest' for team recommendations.")
}
