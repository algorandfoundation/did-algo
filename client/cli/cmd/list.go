package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List registered DIDs",
	Example: "algoid list",
	Aliases: []string{"ls"},
	RunE:    runListCmd,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runListCmd(_ *cobra.Command, _ []string) error {
	// Get store handler
	st, err := getClientStore()
	if err != nil {
		return err
	}

	// Get list of entries
	list := st.List()
	if len(list) == 0 {
		log.Warning("no DIDs registered for the moment")
		return nil
	}

	// Show list of registered entries
	table := tabwriter.NewWriter(os.Stdout, 8, 0, 4, ' ', tabwriter.TabIndent)
	_, _ = fmt.Fprintf(table, "%s\t%s\n", "Name", "DID")
	for k, id := range list {
		_, _ = fmt.Fprintf(table, "%s\t%s\n", k, id.DID())
	}
	return table.Flush()
}
