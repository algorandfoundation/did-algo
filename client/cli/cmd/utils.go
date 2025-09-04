package cmd

import (
	"fmt"

	"github.com/algorandfoundation/did-algo/client/internal"
	"github.com/algorandfoundation/did-algo/client/store"
	"github.com/spf13/viper"
)

// When reading contents from standard input a maximum of 4MB is expected.
const maxPipeInputSize = 4096

// Accessor to the local storage handler.
func getClientStore() (*store.LocalStore, error) {
	return store.NewLocalStore(viper.GetString("home"))
}

// Retrieve the application identifier for the active network profile.
func getStorageAppID(network string) (uint, error) {
	conf := new(internal.ClientSettings)
	if err := viper.UnmarshalKey("network", &conf); err != nil {
		return 0, err
	}
	var profile *internal.NetworkProfile
	for _, p := range conf.Profiles {
		if p.Name == network {
			profile = p
			break
		}
	}
	if profile == nil {
		return 0, fmt.Errorf("no active profile found")
	}
	return profile.AppID, nil
}

// Get network client instance.
func getAlgoClient() (*internal.DIDAlgoStorageClient, error) {
	conf := new(internal.ClientSettings)
	if err := viper.UnmarshalKey("network", &conf); err != nil {
		return nil, err
	}

	return internal.NewAlgoClient(conf.Profiles, log)
}
