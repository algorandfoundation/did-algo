package cmd

import "github.com/spf13/cobra"

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage services enabled for the identifier",
}

func init() {
	editCmd.AddCommand(serviceCmd)
}
