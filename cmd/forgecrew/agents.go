package main

import (
	"fmt"
	"path/filepath"

	"github.com/1424772/ForgeCrew/internal/aci"
	"github.com/1424772/ForgeCrew/internal/agents"
	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/1424772/ForgeCrew/internal/i18n"
	"github.com/1424772/ForgeCrew/internal/models"
	"github.com/spf13/cobra"
)

func agentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage agent definitions",
		Long:  "List, show, and validate agent definitions.",
	}
	cmd.AddCommand(agentsListCmd())
	cmd.AddCommand(agentsValidateCmd())
	return cmd
}

func agentsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered agents",
		RunE:  runAgentsList,
	}
}

func runAgentsList(cmd *cobra.Command, args []string) error {
	registry := agents.NewRegistry()
	path := config.ConfigDir + "/" + config.AgentsYAML
	if err := registry.Load(path); err != nil {
		return fmt.Errorf("load agents: %w", err)
	}
	list := registry.List()
	if len(list) == 0 {
		locale, err := loadLocale()
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), i18n.T("agents.empty", i18n.Locale(locale)))
		return nil
	}
	for _, a := range list {
		fmt.Fprintf(cmd.OutOrStdout(), "  %-20s %s\n", a.AgentID, a.Name)
	}
	return nil
}

func agentsValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate agent definitions against model registry and ACI tools",
		Long: `Cross-validate agent definitions against the model registry and known ACI tools.

Checks:
  1. Every agent's default_model and fallback_models exist in models.yaml
  2. Every agent's tools match known ACI actions or system-reserved names`,
		RunE: runAgentsValidate,
	}
}

func runAgentsValidate(cmd *cobra.Command, args []string) error {
	locale, err := loadLocale()
	if err != nil {
		return err
	}
	loc := i18n.Locale(locale)

	// Load agents.
	agentReg := agents.NewRegistry()
	agentsPath := filepath.Join(config.ConfigDir, config.AgentsYAML)
	if err := agentReg.Load(agentsPath); err != nil {
		return fmt.Errorf("load agents: %w", err)
	}

	// Load models.
	modelReg := models.NewRegistry()
	modelsPath := filepath.Join(config.ConfigDir, config.ModelsYAML)
	if err := modelReg.Load(modelsPath); err != nil {
		return fmt.Errorf("load models: %w", err)
	}

	// Build known tool names: ACI actions + system-reserved tools.
	knownTools := buildKnownToolNames()

	hadError := false
	w := cmd.OutOrStdout()
	ew := cmd.ErrOrStderr()

	// Validate model refs.
	if err := agentReg.ValidateModelRefs(modelReg); err != nil {
		fmt.Fprintf(ew, "%v\n", err)
		hadError = true
	} else {
		fmt.Fprintln(w, i18n.T("validate.model_ok", loc))
	}

	// Validate tools.
	if err := agentReg.ValidateTools(knownTools); err != nil {
		fmt.Fprintf(ew, "%v\n", err)
		hadError = true
	} else {
		fmt.Fprintln(w, i18n.T("validate.tool_ok", loc))
	}

	if hadError {
		return fmt.Errorf("validation failed")
	}
	fmt.Fprintln(w, i18n.T("validate.all_ok", loc))
	return nil
}

// buildKnownToolNames returns the set of all valid tool names:
// ACI actions + system-reserved tools.
func buildKnownToolNames() map[string]bool {
	known := make(map[string]bool)

	// ACI actions.
	for _, a := range aci.AllActions() {
		known[a.Name] = true
	}

	// System-reserved tool names.
	systemTools := []string{
		"scan_project",
		"list_agents",
		"list_models",
		"read_project_profile",
		"read_team_config",
		"dispatch_agent",
		"check_state",
		"build_repo_map",
		"read_memory",
		"write_memory",
		"read_eval",
		"suggest",
	}
	for _, t := range systemTools {
		known[t] = true
	}

	return known
}
