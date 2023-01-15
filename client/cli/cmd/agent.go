package cmd

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/algorandfoundation/did-algo/agent"
	"github.com/algorandfoundation/did-algo/agent/storage"
	"github.com/algorandfoundation/did-algo/info"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	mwHeaders "go.bryk.io/pkg/net/middleware/headers"
	mwProxy "go.bryk.io/pkg/net/middleware/proxy"
	"go.bryk.io/pkg/net/rpc"
	"go.bryk.io/pkg/otel"
)

var agentCmd = &cobra.Command{
	Use:     "agent",
	Short:   "Start a network agent supporting the DID method requirements",
	Example: "algoid agent --port 8080",
	Aliases: []string{"server", "node"},
	RunE:    runMethodServer,
}

func init() {
	params := []cli.Param{
		{
			Name:      "port",
			Usage:     "TCP port to use for the server",
			FlagKey:   "agent.port",
			ByDefault: 9090,
			Short:     "p",
		},
		{
			Name:      "pow",
			Usage:     "set the required request ticket difficulty level",
			FlagKey:   "agent.pow",
			ByDefault: 24,
		},
		{
			Name:      "proxy-protocol",
			Usage:     "enable support for PROXY protocol",
			FlagKey:   "agent.proxy_protocol",
			ByDefault: false,
			Short:     "P",
		},
		{
			Name:      "http",
			Usage:     "enable the HTTP interface",
			FlagKey:   "agent.http",
			ByDefault: false,
		},
		{
			Name:      "env",
			Usage:     "agent environment identifier",
			FlagKey:   "agent.env",
			ByDefault: "dev",
			Short:     "e",
		},
		{
			Name:      "tls",
			Usage:     "enable secure communications using TLS with provided credentials",
			FlagKey:   "agent.tls.enabled",
			ByDefault: false,
		},
		{
			Name:      "tls-ca",
			Usage:     "TLS custom certificate authority (path to PEM file)",
			FlagKey:   "agent.tls.ca",
			ByDefault: "",
		},
		{
			Name:      "tls-cert",
			Usage:     "TLS certificate (path to PEM file)",
			FlagKey:   "agent.tls.cert",
			ByDefault: "/etc/algoid/tls/tls.crt",
		},
		{
			Name:      "tls-key",
			Usage:     "TLS private key (path to PEM file)",
			FlagKey:   "agent.tls.key",
			ByDefault: "/etc/algoid/tls/tls.key",
		},
		{
			Name:      "method",
			Usage:     "specify a supported DID method (can be provided multiple times)",
			FlagKey:   "agent.method",
			ByDefault: []string{"algo"},
			Short:     "m",
		},
		{
			Name:      "storage-kind",
			Usage:     "storage mechanism to use",
			FlagKey:   "agent.storage.kind",
			ByDefault: "ephemeral",
		},
		{
			Name:      "storage-address",
			Usage:     "storage connection endpoint",
			FlagKey:   "agent.storage.addr",
			ByDefault: "",
		},
	}
	if err := cli.SetupCommandParams(agentCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(agentCmd)
}

func runMethodServer(_ *cobra.Command, _ []string) error {
	// Observability operator
	oop, err := otel.NewOperator([]otel.OperatorOption{
		otel.WithLogger(log),
		otel.WithServiceName("algoid"),
		otel.WithServiceVersion(info.CoreVersion),
		otel.WithHostMetrics(),
		otel.WithRuntimeMetrics(5 * time.Second),
		otel.WithResourceAttributes(otel.Attributes{
			"environment": viper.GetString("agent.env"),
		}),
	}...)
	if err != nil {
		return err
	}

	// Prepare API handler
	handler, err := getAgentHandler(oop)
	if err != nil {
		return err
	}

	// Base server configuration
	opts := []rpc.ServerOption{
		rpc.WithPanicRecovery(),
		rpc.WithPort(viper.GetInt("agent.port")),
		rpc.WithNetworkInterface(rpc.NetworkInterfaceAll),
		rpc.WithServiceProvider(handler),
		rpc.WithObservability(oop),
		rpc.WithResourceLimits(rpc.ResourceLimits{
			Connections: 1000,
			Requests:    10,
			Rate:        10000,
		}),
	}

	// TLS configuration
	if viper.GetBool("agent.tls.enabled") {
		log.Debug("TLS enabled")
		opt, err := loadAgentCredentials()
		if err != nil {
			return err
		}
		opts = append(opts, opt)
	}

	// Initialize HTTP gateway
	if viper.GetBool("agent.http") {
		log.Debug("HTTP gateway available")
		gw, err := getAgentGateway(handler)
		if err != nil {
			return err
		}
		opts = append(opts, rpc.WithHTTPGateway(gw))
	}

	// Start server and wait for it to be ready
	log.Debugf("difficulty level: %d", viper.GetInt("agent.pow"))
	log.Debugf("TCP port: %d", viper.GetInt("agent.port"))
	log.Info("starting network agent")
	if viper.GetBool("agent.tls.enabled") {
		log.Debugf("certificate: %s", viper.GetString("agent.tls.cert"))
		log.Debugf("private key: %s", viper.GetString("agent.tls.key"))
	}
	server, err := rpc.NewServer(opts...)
	if err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}
	ready := make(chan bool)
	go func() {
		_ = server.Start(ready)
	}()
	<-ready

	// Wait for system signals
	log.Info("waiting for incoming requests")
	<-cli.SignalsHandler([]os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	})

	// Close handler
	log.Info("preparing to exit")
	if err := server.Stop(true); err != nil {
		log.WithField("error", err).Warning("error stopping server")
	}
	return handler.Close()
}

func getAgentHandler(oop *otel.Operator) (*agent.Handler, error) {
	// Storage
	ss := &storageSettings{}
	if err := viper.UnmarshalKey("agent.storage", ss); err != nil {
		return nil, err
	}
	store, err := getStorage(ss)
	if err != nil {
		return nil, err
	}
	log.Infof("storage: %s", store.Description())

	// Network clients
	algodClient, err := algodClient()
	if err != nil {
		return nil, err
	}
	indexerClient, err := indexerClient()
	if err != nil {
		return nil, err
	}

	// Prepare API handler
	handler, err := agent.NewHandler(agent.HandlerOptions{
		Methods:     viper.GetStringSlice("agent.method"),
		Difficulty:  uint(viper.GetInt("agent.pow")),
		Store:       store,
		OOP:         oop,
		AlgoNode:    algodClient,
		AlgoIndexer: indexerClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start service handler: %w", err)
	}
	return handler, nil
}

func loadAgentCredentials() (rpc.ServerOption, error) {
	var err error
	tlsConf := rpc.ServerTLSConfig{
		IncludeSystemCAs: true,
	}
	tlsConf.Cert, err = os.ReadFile(viper.GetString("agent.tls.cert"))
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate file: %w", err)
	}
	tlsConf.PrivateKey, err = os.ReadFile(viper.GetString("agent.tls.key"))
	if err != nil {
		return nil, fmt.Errorf("failed to load private key file: %w", err)
	}
	if viper.GetString("agent.tls.ca") != "" {
		caPEM, err := os.ReadFile(viper.GetString("agent.tls.ca"))
		if err != nil {
			return nil, fmt.Errorf("failed to load CA file: %w", err)
		}
		tlsConf.CustomCAs = append(tlsConf.CustomCAs, caPEM)
	}
	return rpc.WithTLS(tlsConf), nil
}

func getAgentGateway(handler *agent.Handler) (*rpc.Gateway, error) {
	var gwCl []rpc.ClientOption
	if viper.GetBool("agent.tls.enabled") {
		tlsConf := rpc.ClientTLSConfig{IncludeSystemCAs: true}
		if viper.GetString("agent.tls.ca") != "" {
			caPEM, err := os.ReadFile(viper.GetString("agent.tls.ca"))
			if err != nil {
				return nil, fmt.Errorf("failed to load CA file: %w", err)
			}
			tlsConf.CustomCAs = append(tlsConf.CustomCAs, caPEM)
		}
		gwCl = append(gwCl, rpc.WithClientTLS(tlsConf))
		gwCl = append(gwCl, rpc.WithInsecureSkipVerify()) // Internally the gateway proxy accept any certificate
	}

	gwOpts := []rpc.GatewayOption{
		rpc.WithClientOptions(gwCl...),
		rpc.WithInterceptor(handler.QueryResponseFilter()),
		rpc.WithGatewayMiddleware(mwHeaders.Handler(map[string]string{
			"x-agent-version":        info.CoreVersion,
			"x-agent-build-code":     info.BuildCode,
			"x-agent-release":        releaseCode(),
			"x-content-type-options": "nosniff",
		})),
	}
	if viper.GetBool("agent.proxy_protocol") {
		log.Debug("enable PROXY protocol support")
		gwOpts = append(gwOpts, rpc.WithGatewayMiddleware(mwProxy.Handler()))
	}
	gw, err := rpc.NewGateway(gwOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP gateway: %w", err)
	}
	return gw, nil
}

// Return the proper storage handler instance based on the connection
// details provided.
func getStorage(info *storageSettings) (agent.Storage, error) {
	switch info.Kind {
	case "ephemeral":
		store := &storage.Ephemeral{}
		return store, store.Open("no-op")
	case "mongodb":
		store := &storage.MongoStore{}
		return store, store.Open(info.Addr)
	case "ipfs":
		store := &storage.IPFS{}
		return store, store.Open(info.Addr)
	default:
		return nil, errors.New("non supported storage")
	}
}

type storageSettings struct {
	Kind string `json:"kind" mapstructure:"kind"`
	Addr string `json:"addr" mapstructure:"addr"`
}
