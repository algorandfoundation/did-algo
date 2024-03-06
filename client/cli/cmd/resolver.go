package cmd

import (
	"net/http"
	"os"
	"syscall"

	"github.com/algorandfoundation/did-algo/info"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/did/resolver"
	xHttp "go.bryk.io/pkg/net/http"
	mwHeaders "go.bryk.io/pkg/net/middleware/headers"
)

var resolverCmd = &cobra.Command{
	Use:   "resolver",
	Short: "Run a DID resolver service for the Algorand network",
	RunE:  runResolverCmd,
}

func init() {
	params := []cli.Param{
		{
			Name:      "port",
			Usage:     "TCP port to use for the server",
			FlagKey:   "resolver.port",
			ByDefault: 9091,
			Short:     "p",
		},
		{
			Name:      "proxy-protocol",
			Usage:     "enable support for PROXY protocol",
			FlagKey:   "resolver.proxy_protocol",
			ByDefault: false,
			Short:     "P",
		},
	}
	if err := cli.SetupCommandParams(resolverCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(resolverCmd)
}

func runResolverCmd(_ *cobra.Command, _ []string) error {
	// Network client
	cl, err := getAlgoClient()
	if err != nil {
		return err
	}

	rslv, err := resolver.New(resolver.WithProvider("algo", cl))
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/1.0/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})
	mux.HandleFunc("/1.0/ready", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/1.0/identifiers/", rslv.ResolutionHandler)

	// start server
	srvOpts := []xHttp.Option{
		xHttp.WithHandler(mux),
		xHttp.WithPort(viper.GetInt("resolver.port")),
		xHttp.WithMiddleware(mwHeaders.Handler(map[string]string{
			"x-resolver-version": info.CoreVersion,
			"x-resolver-build":   info.BuildCode,
		})),
	}
	srv, err := xHttp.NewServer(srvOpts...)
	if err != nil {
		return err
	}
	go srv.Start() // nolint:errcheck

	// wait for system signals
	log.Info("waiting for incoming requests")
	<-cli.SignalsHandler([]os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	})

	// stop server
	log.Info("closing resolver server")
	return srv.Stop(true)
}
