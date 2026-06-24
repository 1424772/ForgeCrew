package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information — set via ldflags at build time.
var (
	Version   = "0.1.0"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of forgecrew",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("forgecrew version %s\n", Version)
			fmt.Printf("  commit: %s\n", GitCommit)
			fmt.Printf("  built:  %s\n", BuildDate)
		},
	}
}
