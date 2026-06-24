package main

import (
	"fmt"

	"github.com/1424772/ForgeCrew/internal/scanner"
	"github.com/spf13/cobra"
)

var scanOutputFormat string

func scanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan the current directory and output a project profile",
		Long:  "Scan the current directory to identify languages, frameworks, tools, and generate a project profile.",
		RunE:  runScan,
	}
	cmd.Flags().StringVarP(&scanOutputFormat, "output", "o", "yaml", "Output format: yaml or json")
	return cmd
}

func runScan(cmd *cobra.Command, args []string) error {
	s := scanner.New()
	profile, err := s.Scan(".")
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	output, err := profile.Format(scanOutputFormat)
	if err != nil {
		return err
	}
	fmt.Print(output)
	return nil
}
