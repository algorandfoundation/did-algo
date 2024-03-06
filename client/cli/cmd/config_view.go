package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var configViewCmd = &cobra.Command{
	Use:     "view",
	Example: "algoid config view",
	Short:   "View current configuration settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := new(appConf)
		if err := viper.Unmarshal(&conf); err != nil {
			return err
		}
		output, _ := yaml.Marshal(conf)
		log.Info("configuration file: ", viper.ConfigFileUsed())
		fmt.Printf("%s\n", output)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configViewCmd)
}
