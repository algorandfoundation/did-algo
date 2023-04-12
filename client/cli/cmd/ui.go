package cmd

import (
	"net/http"
	"os"
	"os/exec"
	"syscall"

	"github.com/algorandfoundation/did-algo/client/internal"
	"github.com/algorandfoundation/did-algo/client/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	xhttp "go.bryk.io/pkg/net/http"
)

var uiCmd = &cobra.Command{
	Use:     "ui",
	Aliases: []string{"gui"},
	Short:   "Start the local graphical client",
	RunE:    runLocalUI,
	Long: `Graphical client

Starts a local graphical user interface that can be used
to create and manage local identifiers and connect your
wallet for more advanced features.`,
}

func init() {
	rootCmd.AddCommand(uiCmd)
}

func runLocalUI(cmd *cobra.Command, args []string) error {
	// Load client configuration
	conf := new(internal.ClientSettings)
	if err := viper.UnmarshalKey("client", conf); err != nil {
		return err
	}
	if err := conf.Validate(); err != nil {
		return err
	}

	// Get store handler
	st, err := getClientStore()
	if err != nil {
		return err
	}

	// Local API server
	srv, err := ui.LocalAPIServer(st, conf, log)
	if err != nil {
		return err
	}
	log.Info("starting local API server")
	go func() {
		_ = srv.Start()
	}()

	log.Info("starting local app server")
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(ui.AppContents)))
	appSrv, _ := xhttp.NewServer(xhttp.WithHandler(mux), xhttp.WithPort(8080))
	go func() {
		_ = appSrv.Start()
	}()
	if err = exec.Command("open", "http://localhost:8080/").Run(); err != nil {
		log.Info("open: http://localhost:8080/")
	}

	// Wait for system signals
	log.Info("waiting for incoming requests")
	<-cli.SignalsHandler([]os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	})

	// Close handler
	log.Info("stopping local API server")
	_ = appSrv.Stop(true)
	return srv.Stop()
}
