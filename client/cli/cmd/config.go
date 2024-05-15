package cmd

import (
	"os"

	"github.com/algorandfoundation/did-algo/client/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"cfg", "conf", "settings"},
	Short:   "Adjust configuration settings for the CLI client",
}

func init() {
	rootCmd.AddCommand(configCmd)
}

type appConf struct {
	Home    string                   `json:"home" yaml:"home" mapstructure:"home"`
	Network *internal.ClientSettings `json:"network" yaml:"network" mapstructure:"network"`
}

func (ac *appConf) isProfileAvailable(name string) bool {
	for _, p := range ac.Network.Profiles {
		if p.Name == name {
			return true
		}
	}
	return false
}

func (ac *appConf) setAppID(network string, appID uint) {
	for _, p := range ac.Network.Profiles {
		if p.Name == network {
			p.AppID = appID
		}
	}
}

func (ac *appConf) save() error {
	file := viper.ConfigFileUsed()
	output, _ := yaml.Marshal(ac)
	return os.WriteFile(file, output, 0644) // nolint
}
