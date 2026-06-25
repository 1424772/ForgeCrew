package main

import (
	"fmt"
	"path/filepath"

	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/1424772/ForgeCrew/internal/i18n"
	"github.com/1424772/ForgeCrew/internal/settings"
	"github.com/spf13/cobra"
)

// Version information — set via ldflags at build time.
var (
	Version   = "0.1.0"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of forgecrew",
		RunE: func(cmd *cobra.Command, args []string) error {
			locale, err := loadLocale()
			if err != nil {
				return err
			}
			loc := i18n.Locale(locale)

			fmt.Fprintln(cmd.OutOrStdout(), i18n.T("version.header", loc)+Version)
			fmt.Fprintln(cmd.OutOrStdout(), i18n.T("version.commit", loc)+GitCommit)
			fmt.Fprintln(cmd.OutOrStdout(), i18n.T("version.built", loc)+BuildDate)
			return nil
		},
	}
}

// loadLocale reads the current language from settings.
// Returns an error if settings.yaml has an invalid language value.
func loadLocale() (string, error) {
	settingsPath := filepath.Join(config.ConfigDir, settings.SettingsYAML)
	s, err := settings.Load(settingsPath)
	if err != nil {
		return "", err
	}
	return s.Language, nil
}
