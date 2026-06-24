package main

import (
	"fmt"

	"github.com/1424772/ForgeCrew/internal/agents"
	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/spf13/cobra"
)

func agentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage agent definitions",
		Long:  "List, show, and manage agent definitions.",
	}
	cmd.AddCommand(agentsListCmd())
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
		fmt.Println("No agents registered. Run 'forgecrew init' first.")
		return nil
	}
	for _, a := range list {
		fmt.Printf("  %-20s %s\n", a.AgentID, a.Name)
	}
	return nil
}
