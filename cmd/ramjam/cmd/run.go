package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/michaelmccabe/ramjam/pkg/runner"
	"github.com/spf13/cobra"
)

var cliVars []string

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

		// Parse CLI variables overrides
		varMap := make(map[string]string)
		for _, v := range cliVars {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) == 2 {
				varMap[parts[0]] = parts[1]
			}
		}
		r.SetVars(varMap)

		err := r.RunPaths(args)
		if err == nil {
			fmt.Println("All steps were run successfully")
			return nil
		}

		if errs, ok := err.(interface{ Unwrap() []error }); ok {
			for _, e := range errs.Unwrap() {
				if se, ok := e.(*runner.StepError); ok {
					fmt.Printf("Failed step: %s\n", se.Step)
					if verbose {
						fmt.Printf("Description: %s\n", se.Description)
						fmt.Printf("Error: %v\n", se.Err)
					}
				} else {
					fmt.Printf("Error: %v\n", e)
				}
			}
			return fmt.Errorf("workflow failed with %d errors", len(errs.Unwrap()))
		}

		return fmt.Errorf("run failed: %w", err)
	},
}

func init() {
	runCmd.Flags().StringSliceVar(&cliVars, "var", nil, "Set variables at runtime (e.g. --var key=val)")
	rootCmd.AddCommand(runCmd)
}

