package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "0.0.1-development" // Version is the version of the CLI tool

	rootCmd = &cobra.Command{
		Use:   "ramjam",
		Short: "ramjam - CLI tool to execute HTTP API workflows via YAML",
		Long: `ramjam is a command-line tool for executing HTTP API workflows defined in YAML files.

All HTTP requests are made through declarative YAML workflow files, providing:
- Reproducible API testing
- Variable substitution and captures
- Response validation with JSONPath
- Support for all HTTP methods`,

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
