package main

import (
	"fmt"

	"github.com/1424772/ForgeCrew/internal/gitops"
	"github.com/spf13/cobra"
)

func diffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Show git diff of the current repository",
		Long:  "Run git diff and display the output. Returns a friendly error if not in a git repository.",
		RunE:  runDiff,
	}
}

func runDiff(cmd *cobra.Command, args []string) error {
	diff, err := gitops.ReadDiff(".")
	if err != nil {
		return fmt.Errorf("diff failed: %w", err)
	}
	fmt.Print(diff)
	return nil
}
