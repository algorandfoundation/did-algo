package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"

	"github.com/algorandfoundation/did-algo/agent"
	"github.com/algorandfoundation/did-algo/agent/storage"
	"github.com/algorandfoundation/did-algo/info"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
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
			Name:      "http",
			Usage:     "enable the HTTP interface",
			FlagKey:   "agent.http",
			ByDefault: false,
		},
		{
			Name:      "monitoring",
			Usage:     "publish metrics for instrumentation and monitoring",
			FlagKey:   "agent.monitoring",
			ByDefault: false,
		},
		{
			Name:      "debug",
			Usage:     "run agent in debug mode to generate profiling information",
			FlagKey:   "agent.debug",
			ByDefault: false,
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
			ByDefault: "/etc/algoid/agent/tls.crt",
		},
		{
			Name:      "tls-key",
			Usage:     "TLS private key (path to PEM file)",
			FlagKey:   "agent.tls.key",
			ByDefault: "/etc/algoid/agent/tls.key",
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
	if err := cli.SetupCommandParams(agentCmd, params); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(agentCmd)
}

func runMethodServer(_ *cobra.Command, _ []string) error {
	// CPU profile
	if viper.GetBool("agent.debug") {
		cpuProfileHook, err := cpuProfile()
		if err != nil {
			return err
		}
		defer cpuProfileHook()
	}

	// Observability operator
	oop, err := otel.NewOperator([]otel.OperatorOption{
		otel.WithLogger(log),
		otel.WithServiceName("algoid"),
		otel.WithServiceVersion(info.CoreVersion),
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
		rpc.WithService(handler.ServiceDefinition()),
		rpc.WithObservability(oop),
	}

	// TLS configuration
	if viper.GetBool("agent.tls.enabled") {
		log.Info("TLS enabled")
		opt, err := loadAgentCredentials()
		if err != nil {
			return err
		}
		opts = append(opts, opt)
	}

	// Initialize HTTP gateway
	if viper.GetBool("agent.http") {
		log.Info("HTTP gateway available")
		gw, err := getAgentGateway(handler)
		if err != nil {
			return err
		}
		opts = append(opts, rpc.WithHTTPGateway(gw))
	}

	// Start server and wait for it to be ready
	log.Infof("difficulty level: %d", viper.GetInt("agent.pow"))
	log.Infof("TCP port: %d", viper.GetInt("agent.port"))
	log.Info("starting network agent")
	if viper.GetBool("agent.tls.enabled") {
		log.Infof("certificate: %s", viper.GetString("agent.tls.cert"))
		log.Infof("private key: %s", viper.GetString("agent.tls.key"))
	}
	server, err := rpc.NewServer(opts...)
	if err != nil {
		return fmt.Errorf("failed to start node: %s", err)
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
	if err = handler.Close(); err != nil && !strings.Contains(err.Error(), "closed network connection") {
		return err
	}

	// Dump memory profile and exit
	if viper.GetBool("agent.debug") {
		if err := memoryProfile(); err != nil {
			return err
		}
	}
	return nil
}

func getAgentHandler(oop *otel.Operator) (*agent.Handler, error) {
	// Get handler settings
	ss := &storageSettings{}
	if err := viper.UnmarshalKey("agent.storage", ss); err != nil {
		return nil, err
	}
	methods := viper.GetStringSlice("agent.method")
	pow := uint(viper.GetInt("agent.pow"))
	store, err := getStorage(ss)
	if err != nil {
		return nil, err
	}

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
		Methods:     methods,
		Difficulty:  pow,
		Store:       store,
		OOP:         oop,
		AlgoNode:    algodClient,
		AlgoIndexer: indexerClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start method handler: %s", err)
	}
	log.Infof("storage: %s", store.Description())
	return handler, nil
}

func loadAgentCredentials() (rpc.ServerOption, error) {
	var err error
	tlsConf := rpc.ServerTLSConfig{
		IncludeSystemCAs: true,
	}
	tlsConf.Cert, err = ioutil.ReadFile(viper.GetString("agent.tls.cert"))
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate file: %s", err)
	}
	tlsConf.PrivateKey, err = ioutil.ReadFile(viper.GetString("agent.tls.key"))
	if err != nil {
		return nil, fmt.Errorf("failed to load private key file: %s", err)
	}
	if viper.GetString("agent.tls.ca") != "" {
		caPEM, err := ioutil.ReadFile(viper.GetString("agent.tls.ca"))
		if err != nil {
			return nil, fmt.Errorf("failed to load CA file: %s", err)
		}
		tlsConf.CustomCAs = append(tlsConf.CustomCAs, caPEM)
	}
	return rpc.WithTLS(tlsConf), nil
}

func getAgentGateway(handler *agent.Handler) (*rpc.HTTPGateway, error) {
	gwCl := []rpc.ClientOption{rpc.WaitForReady()}
	if viper.GetBool("agent.tls.enabled") {
		tlsConf := rpc.ClientTLSConfig{IncludeSystemCAs: true}
		if viper.GetString("agent.tls.ca") != "" {
			caPEM, err := ioutil.ReadFile(viper.GetString("agent.tls.ca"))
			if err != nil {
				return nil, fmt.Errorf("failed to load CA file: %s", err)
			}
			tlsConf.CustomCAs = append(tlsConf.CustomCAs, caPEM)
		}
		gwCl = append(gwCl, rpc.WithClientTLS(tlsConf))
		gwCl = append(gwCl, rpc.WithInsecureSkipVerify()) // Internally the gateway proxy accept any certificate
	}

	gwOpts := []rpc.HTTPGatewayOption{
		rpc.WithClientOptions(gwCl),
		rpc.WithFilter(handler.QueryResponseFilter()),
	}
	gw, err := rpc.NewHTTPGateway(gwOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP gateway: %s", err)
	}
	return gw, nil
}

func cpuProfile() (func(), error) {
	cpu, err := ioutil.TempFile("", "algoid_cpu_")
	if err != nil {
		return nil, err
	}
	if err := pprof.StartCPUProfile(cpu); err != nil {
		return nil, err
	}
	return func() {
		log.Infof("CPU profile saved at %s", cpu.Name())
		pprof.StopCPUProfile()
		_ = cpu.Close()
	}, nil
}

func memoryProfile() error {
	mem, err := ioutil.TempFile("", "algoid_mem_")
	if err != nil {
		return err
	}
	runtime.GC()
	if err := pprof.WriteHeapProfile(mem); err != nil {
		return err
	}
	log.Infof("memory profile saved at %s", mem.Name())
	return mem.Close()
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
