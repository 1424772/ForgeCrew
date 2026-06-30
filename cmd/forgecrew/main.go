// Package main is the entry point for the forgecrew CLI.
//
// ForgeCrew is a Go-native, CLI-first AI Software Company Runtime.
// It provides dynamic team assembly for multi-agent engineering workflows.
package main

import (
	"os"

	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "forgecrew",
	Short: "ForgeCrew - AI Software Company Runtime",
	Long: `ForgeCrew is a Go-native, CLI-first AI Software Company Runtime.
It dynamically assembles AI agent teams for software engineering tasks.

Not a single coding agent — a virtual software company that builds teams
based on your project's needs.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	// When no subcommand is given, enter interactive mode instead of
	// printing help. Pass --help explicitly to see help.
	RunE: func(cmd *cobra.Command, args []string) error {
		return interactiveMode()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(scanCmd())
	rootCmd.AddCommand(agentsCmd())
	rootCmd.AddCommand(modelsCmd())
	rootCmd.AddCommand(teamCmd())
	rootCmd.AddCommand(diffCmd())
	rootCmd.AddCommand(checkpointCmd())
	rootCmd.AddCommand(taskCmd())
	rootCmd.AddCommand(langCmd())
	rootCmd.AddCommand(runsCmd())
	rootCmd.AddCommand(handoffCmd())
	rootCmd.AddCommand(ctoCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		config.PrintError(err)
		os.Exit(1)
	}
}
