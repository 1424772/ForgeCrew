package main

import (
	"fmt"

	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/1424772/ForgeCrew/internal/models"
	"github.com/spf13/cobra"
)

func modelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "Manage model definitions",
		Long:  "List, show, and manage model definitions.",
	}
	cmd.AddCommand(modelsListCmd())
	return cmd
}

func modelsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered models",
		RunE:  runModelsList,
	}
}

func runModelsList(cmd *cobra.Command, args []string) error {
	registry := models.NewRegistry()
	path := config.ConfigDir + "/" + config.ModelsYAML
	if err := registry.Load(path); err != nil {
		return fmt.Errorf("load models: %w", err)
	}
	list := registry.List()
	if len(list) == 0 {
		fmt.Println("No models registered. Run 'forgecrew init' first.")
		return nil
	}
	for _, m := range list {
		fmt.Printf("  %-20s %s/%s  tier=%s\n", m.ModelID, m.Provider, m.Model, m.CostTier)
	}
	return nil
}
