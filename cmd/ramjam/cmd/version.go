package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ramjam",
	Long:  `All software has versions. This is ramjam's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ramjam version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
