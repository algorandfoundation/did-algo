package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generates bash completion scripts",
	RunE: func(_ *cobra.Command, _ []string) error {
		return rootCmd.GenBashCompletion(os.Stdout)
	},
}

func init() {
	completionCmd.AddCommand(bashCmd)
}
