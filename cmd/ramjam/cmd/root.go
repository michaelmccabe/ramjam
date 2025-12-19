package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is the version of the CLI tool
	Version = "dev"

	rootCmd = &cobra.Command{
		Use:   "ramjam",
		Short: "ramjam - CLI tool to test HTTP APIs",
		Long: `ramjam is a command-line tool designed to test and interact with HTTP APIs.
		
It provides a simple and intuitive interface for making HTTP requests,
inspecting responses, and validating API behavior.`,
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			// If no subcommand is provided, show help
			cmd.Help()
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}
