package main

import (
	"fmt"

	"github.com/1424772/ForgeCrew/internal/scanner"
	"github.com/1424772/ForgeCrew/internal/teamarchitect"
	"github.com/spf13/cobra"
)

func teamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage team configuration",
		Long:  "Suggest, show, and apply team configurations based on project analysis.",
	}
	cmd.AddCommand(teamSuggestCmd())
	return cmd
}

func teamSuggestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "suggest",
		Short: "Suggest a team configuration based on project scan",
		Long:  "Scan the project and suggest an appropriate team composition.",
		RunE:  runTeamSuggest,
	}
}

func runTeamSuggest(cmd *cobra.Command, args []string) error {
	s := scanner.New()
	profile, err := s.Scan(".")
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	ta := teamarchitect.New()
	teamConfig := ta.Suggest(profile)
	output, err := teamConfig.FormatYAML()
	if err != nil {
		return err
	}
	fmt.Print(output)
	return nil
}
