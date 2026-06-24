package main

import (
	"fmt"

	"github.com/1424772/ForgeCrew/internal/config"
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
	if err := writeIfMissing("AGENTS.md", config.DefaultAgentsMD, force); err != nil {
		return err
	}

	// Write .forgecrew/*.yaml
	cfgFiles := map[string]string{
		config.ConfigDir + "/" + config.AgentsYAML:    config.DefaultAgentsYAML,
		config.ConfigDir + "/" + config.ModelsYAML:    config.DefaultModelsYAML,
		config.ConfigDir + "/" + config.WorkflowsYAML: config.DefaultWorkflowsYAML,
	}
	for path, content := range cfgFiles {
		if err := writeIfMissing(path, content, force); err != nil {
			return err
		}
	}

	fmt.Println("ForgeCrew initialized successfully.")
	fmt.Println("  AGENTS.md created")
	fmt.Println("  .forgecrew/agents.yaml created")
	fmt.Println("  .forgecrew/models.yaml created")
	fmt.Println("  .forgecrew/workflows.yaml created")
	fmt.Println("  .forgecrew/memory/ created")
	fmt.Println("  .forgecrew/evals/ created")
	fmt.Println("  .forgecrew/checkpoints/ created")
	return nil
}
