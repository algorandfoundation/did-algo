package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// zshCmd represents the zsh command
var zshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Generates zsh completion scripts",
	RunE: func(_ *cobra.Command, _ []string) error {
		return rootCmd.GenZshCompletion(os.Stdout)
	},
}

func init() {
	completionCmd.AddCommand(zshCmd)
}
