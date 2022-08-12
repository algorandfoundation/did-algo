package cmd

import (
	"fmt"
	"runtime"
	"time"

	"github.com/algorandfoundation/did-algo/info"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		var components = map[string]string{
			"Version":    info.CoreVersion,
			"Build code": info.BuildCode,
			"OS/Arch":    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"Go version": runtime.Version(),
			"Release":    releaseCode(),
		}
		if info.BuildTimestamp != "" {
			rd, err := time.Parse(time.RFC3339, info.BuildTimestamp)
			if err == nil {
				components["Release Date"] = rd.Format(time.RFC822)
			}
		}
		for k, v := range components {
			fmt.Printf("\033[21;37m%-13s:\033[0m %s\n", k, v)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func releaseCode() string {
	release := "algoid"
	if info.CoreVersion != "" {
		release += "@" + info.CoreVersion
	}
	if info.BuildCode != "" {
		release += "+" + info.BuildCode
	}
	return release
}
