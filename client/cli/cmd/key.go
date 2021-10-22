package cmd

import "github.com/spf13/cobra"

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage cryptographic keys associated with the DID",
}

func init() {
	editCmd.AddCommand(keyCmd)
}
