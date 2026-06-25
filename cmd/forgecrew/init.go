package main

import (
	"fmt"
	"path/filepath"

	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/1424772/ForgeCrew/internal/i18n"
	"github.com/1424772/ForgeCrew/internal/settings"
	"github.com/spf13/cobra"
)

var initForce bool

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize ForgeCrew configuration in the current directory",
		Long: `Initialize ForgeCrew configuration in the current directory.

Creates AGENTS.md and .forgecrew/ with default configuration files.
Re-running will not overwrite existing files unless --force is provided.`,
		RunE: runInit,
	}
	cmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing files")
	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	force := initForce

	// Read locale from settings, default to zh.
	locale := "zh"
	settingsPath := filepath.Join(config.ConfigDir, settings.SettingsYAML)
	if s, err := settings.Load(settingsPath); err == nil {
		locale = s.Language
	}
	loc := i18n.Locale(locale)

	if err := config.EnsureDir(config.ConfigDir); err != nil {
		return err
	}
	if err := config.EnsureDir(config.ConfigDir + "/" + config.MemoryDir); err != nil {
		return err
	}
	if err := config.EnsureDir(config.ConfigDir + "/" + config.EvalsDir); err != nil {
		return err
	}
	if err := config.EnsureDir(config.ConfigDir + "/" + config.CheckpointsDir); err != nil {
		return err
	}

	// Write AGENTS.md
	w := cmd.OutOrStdout()
	if err := writeIfMissing(w, "AGENTS.md", config.DefaultAgentsMD, force, loc); err != nil {
		return err
	}

	// Write .forgecrew/*.yaml
	cfgFiles := map[string]string{
		config.ConfigDir + "/" + config.AgentsYAML:    config.DefaultAgentsYAML,
		config.ConfigDir + "/" + config.ModelsYAML:    config.DefaultModelsYAML,
		config.ConfigDir + "/" + config.WorkflowsYAML: config.DefaultWorkflowsYAML,
	}
	for path, content := range cfgFiles {
		if err := writeIfMissing(w, path, content, force, loc); err != nil {
			return err
		}
	}

	// Write settings.yaml (always create if missing, overwrite if --force).
	if !config.FileExists(settingsPath) || force {
		if err := settings.Save(settingsPath, settings.Default()); err != nil {
			return fmt.Errorf("write settings.yaml: %w", err)
		}
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), i18n.T("init.skipping", loc)+settingsPath)
	}

	fmt.Fprintln(cmd.OutOrStdout(), i18n.T("init.success", loc))
	fmt.Fprintln(cmd.OutOrStdout(), "  AGENTS.md"+i18n.T("init.created", loc))
	fmt.Fprintln(cmd.OutOrStdout(), "  "+config.ConfigDir+"/"+config.AgentsYAML+i18n.T("init.created", loc))
	fmt.Fprintln(cmd.OutOrStdout(), "  "+config.ConfigDir+"/"+config.ModelsYAML+i18n.T("init.created", loc))
	fmt.Fprintln(cmd.OutOrStdout(), "  "+config.ConfigDir+"/"+config.WorkflowsYAML+i18n.T("init.created", loc))
	fmt.Fprintln(cmd.OutOrStdout(), "  "+config.ConfigDir+"/"+config.SettingsYAML+i18n.T("init.created", loc))
	fmt.Fprintln(cmd.OutOrStdout(), "  "+config.ConfigDir+"/"+config.MemoryDir+"/"+i18n.T("init.created", loc))
	fmt.Fprintln(cmd.OutOrStdout(), "  "+config.ConfigDir+"/"+config.EvalsDir+"/"+i18n.T("init.created", loc))
	fmt.Fprintln(cmd.OutOrStdout(), "  "+config.ConfigDir+"/"+config.CheckpointsDir+"/"+i18n.T("init.created", loc))
	return nil
}
