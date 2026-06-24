package main

import (
	"fmt"

	"github.com/1424772/ForgeCrew/internal/checkpoint"
	"github.com/spf13/cobra"
)

func checkpointCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkpoint",
		Short: "Manage checkpoints",
		Long:  "List and manage checkpoints for rollback safety.",
	}
	cmd.AddCommand(checkpointListCmd())
	return cmd
}

func checkpointListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all checkpoints",
		RunE:  runCheckpointList,
	}
}

func runCheckpointList(cmd *cobra.Command, args []string) error {
	store := checkpoint.NewStore()
	ckpts, err := store.List()
	if err != nil {
		return fmt.Errorf("list checkpoints: %w", err)
	}
	if len(ckpts) == 0 {
		fmt.Println("No checkpoints found.")
		return nil
	}
	for _, c := range ckpts {
		fmt.Printf("  %s  task=%s  agent=%s  time=%s\n", c.ID, c.TaskID, c.AgentID, c.Timestamp)
		if len(c.ChangedFiles) > 0 {
			for _, f := range c.ChangedFiles {
				fmt.Printf("    changed: %s\n", f)
			}
		}
	}
	return nil
}
