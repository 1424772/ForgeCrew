package main

import (
	"fmt"

	"github.com/1424772/ForgeCrew/internal/orchestrator"
	"github.com/spf13/cobra"
)

var taskDryRun bool

func taskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task <goal>",
		Short: "Execute a task through the Loop Engineering state machine",
		Long: `Execute a task through the full Loop Engineering cycle.

In --dry-run mode, prints the state sequence without making any changes.
The Loop Engineering cycle runs: Goal -> Plan -> Retrieve -> Act ->
Observe -> Reflect -> Improve -> Review -> CommitMemory`,
		Args: cobra.ExactArgs(1),
		RunE: runTask,
	}
	cmd.Flags().BoolVar(&taskDryRun, "dry-run", false, "Run in dry-run mode (no real execution)")
	return cmd
}

func runTask(cmd *cobra.Command, args []string) error {
	goal := args[0]
	sm := orchestrator.New(3, taskDryRun)
	result := sm.RunFull(goal)

	// Read locale from settings.
	locale, err := loadLocale()
	if err != nil {
		return err
	}

	fmt.Fprint(cmd.OutOrStdout(), result.FormatText(locale))
	return nil
}
