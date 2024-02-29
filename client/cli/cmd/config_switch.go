package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configSwithProfileCmd = &cobra.Command{
	Use:     "switch",
	Example: "algoid config switch [profile]",
	Short:   "Modify the active network profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("profile name is required")
		}
		conf := new(appConf)
		if err := viper.Unmarshal(&conf); err != nil {
			return err
		}
		if !conf.isProfileAvailable(args[0]) {
			return fmt.Errorf("profile '%s' not found", args[0])
		}
		conf.Network.Active = args[0]
		err := conf.save()
		if err == nil {
			log.Info("configuration updated")
		}
		return err
	},
}

func init() {
	configCmd.AddCommand(configSwithProfileCmd)
}
