package cmd

import (
	"net/http"
	"os"
	"syscall"

	"github.com/algorandfoundation/did-algo/client/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	pkgHttp "go.bryk.io/pkg/net/http"
	"google.golang.org/grpc"
)

var resolverCmd = &cobra.Command{
	Use:   "resolver",
	Short: "Start a standalone resolver server",
	RunE:  runResolverServer,
	Long: `Resolver server

Resolver server provides a DIF compatible endpoint for DID
document resolution. The server endpoint can be used as a
standalone Universal Resolver Driver.

More information:
https://github.com/decentralized-identity/universal-resolver`,
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
		{
			Name:      "tls",
			Usage:     "enable secure communications using TLS with provided credentials",
			FlagKey:   "resolver.tls.enabled",
			ByDefault: false,
		},
		{
			Name:      "tls-ca",
			Usage:     "TLS custom certificate authority (path to PEM file)",
			FlagKey:   "resolver.tls.ca",
			ByDefault: "",
		},
		{
			Name:      "tls-cert",
			Usage:     "TLS certificate (path to PEM file)",
			FlagKey:   "resolver.tls.cert",
			ByDefault: "/etc/algoid/tls/tls.crt",
		},
		{
			Name:      "tls-key",
			Usage:     "TLS private key (path to PEM file)",
			FlagKey:   "resolver.tls.key",
			ByDefault: "/etc/algoid/tls/tls.key",
		},
		{
			Name:      "agent",
			Usage:     "Network agent to communicate with",
			FlagKey:   "resolver.client.node",
			ByDefault: internal.DefaultAgentEndpoint,
			Short:     "a",
		},
		{
			Name:      "agent-insecure",
			Usage:     "Use an insecure connection to the network agent",
			FlagKey:   "resolver.client.insecure",
			ByDefault: false,
		},
	}
	if err := cli.SetupCommandParams(resolverCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(resolverCmd)
}

func runResolverServer(cmd *cobra.Command, args []string) error {
	// Load settings
	conf := new(internal.ResolverSettings)
	conf.Load(viper.GetViper())

	// Resolver instance
	conn, err := getClientConnection(conf.Client)
	if err != nil {
		return err
	}
	rr, err := conf.Resolver(conn)
	if err != nil {
		return err
	}

	// Start server
	mux := http.NewServeMux()
	mux.HandleFunc("/1.0/ping", pingHandler)
	mux.HandleFunc("/1.0/ready", resolverReadyHandler(conn))
	mux.HandleFunc("/1.0/identifiers/", rr.ResolutionHandler)
	srv, err := pkgHttp.NewServer(conf.ServerOpts(mux, releaseCode())...)
	if err != nil {
		return err
	}
	go func() {
		_ = srv.Start()
	}()

	// wait for system signals
	log.Info("waiting for incoming requests")
	<-cli.SignalsHandler([]os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	})

	// stop server
	log.Info("preparing to exit")
	err = srv.Stop(true) // prevent further requests
	_ = conn.Close()     // close internal API client connection
	return err
}

// Basic reachability test.
func pingHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("pong"))
}

// Status reported is based on the connection to the agent being used.
func resolverReadyHandler(conn *grpc.ClientConn) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		st := conn.GetState().String()
		if st != "READY" && st != "IDLE" {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(st))
			return
		}
	}
	return fn
}
