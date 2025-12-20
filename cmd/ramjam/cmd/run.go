package cmd

import (
	"fmt"
	"time"

	"github.com/michaelmccabe/ramjam/pkg/runner"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <files-or-folders...>",
	Short: "Execute YAML-defined API workflows",
	Long: `Execute one or more YAML workflow files, or all YAML files in a directory.
Examples:
  ramjam run test-get.yaml
  ramjam run ./tests/integration/
  ramjam run login.yaml signup.yaml profile.yaml`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		r := runner.New(30*time.Second, verbose)
		if err := r.RunPaths(args); err != nil {
			return fmt.Errorf("run failed: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
