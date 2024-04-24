package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configAppIDCmd = &cobra.Command{
	Use:     "app-id",
	Example: "algoid config app-id [app-id] [network]",
	Short:   "Adjust the `app-id` setting for the active network profile",
	RunE: func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("appID name is required")
		}

		if len(args) == 1 {
			return fmt.Errorf("network name is required")
		}

		conf := new(appConf)
		if err := viper.Unmarshal(&conf); err != nil {
			return err
		}
		appID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid app-id: '%s'", args[0])
		}
		conf.setAppID(args[1], uint(appID))
		err = conf.save()
		if err == nil {
			log.Info("configuration updated")
		}
		return err
	},
}

func init() {
	configCmd.AddCommand(configAppIDCmd)
}
