package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/1424772/ForgeCrew/internal/i18n"
	"github.com/1424772/ForgeCrew/internal/settings"
	"github.com/spf13/cobra"
)

func langCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "language",
		Aliases: []string{"lang"},
		Short:   "Show or set the current language (zh/en)",
		Long:    "Show, set, or list supported languages for ForgeCrew output.",
	}
	cmd.AddCommand(langShowCmd())
	cmd.AddCommand(langSetCmd())
	cmd.AddCommand(langListCmd())
	return cmd
}

func langShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the current language setting",
		RunE:  runLangShow,
	}
}

func runLangShow(cmd *cobra.Command, args []string) error {
	locale, err := loadLocale()
	if err != nil {
		return err
	}
	loc := i18n.Locale(locale)
	fmt.Fprintln(cmd.OutOrStdout(), i18n.T("lang.current", loc)+locale)
	return nil
}

func langSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <zh|en>",
		Short: "Set the current language (zh or en)",
		Args:  cobra.ExactArgs(1),
		RunE:  runLangSet,
	}
}

func runLangSet(cmd *cobra.Command, args []string) error {
	lang := strings.ToLower(strings.TrimSpace(args[0]))
	loc, err := i18n.ValidLocale(lang)
	if err != nil {
		return err
	}

	settingsPath := filepath.Join(config.ConfigDir, config.SettingsYAML)
	// Ensure .forgecrew/ exists.
	if err := config.EnsureDir(config.ConfigDir); err != nil {
		return err
	}

	s := settings.Default()
	s.Language = string(loc)
	if err := settings.Save(settingsPath, s); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), i18n.T("lang.set_to", loc)+string(loc))
	return nil
}

func langListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all supported languages",
		RunE:  runLangList,
	}
}

func runLangList(cmd *cobra.Command, args []string) error {
	locale, err := loadLocale()
	if err != nil {
		return err
	}
	loc := i18n.Locale(locale)
	fmt.Fprintln(cmd.OutOrStdout(), i18n.T("lang.available", loc))
	for _, l := range i18n.SupportedLocales() {
		marker := " "
		if l == string(loc) {
			marker = "*"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %s %s\n", marker, l)
	}
	return nil
}
