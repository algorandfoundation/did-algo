package internal

import (
	"net/http"
	"time"

	"github.com/algorandfoundation/did-algo/info"
	protoV1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/did/resolver"
	pkgHttp "go.bryk.io/pkg/net/http"
	"go.bryk.io/pkg/net/middleware"
	otelHttp "go.bryk.io/pkg/otel/http"
	"google.golang.org/grpc"
)

// DefaultAgentEndpoint defines the standard network agent to use
// when no user-provided value is available.
const DefaultAgentEndpoint = "algo-did.aidtech.network:443"

// ResolverSettings defines the configuration options available when
// deploying a DIF compliant resolver endpoint.
type ResolverSettings struct {
	Port          uint            `json:"port" yaml:"port" mapstructure:"port"`
	ProxyProtocol bool            `json:"proxy_protocol" yaml:"proxy_protocol" mapstructure:"proxy_protocol"`
	TLS           *tlsSettings    `json:"tls" yaml:"tls" mapstructure:"tls"`
	Client        *ClientSettings `json:"client" yaml:"client" mapstructure:"client"`
}

// Load available configuration values and set sensible default values.
func (rs *ResolverSettings) Load(v *viper.Viper) {
	_ = v.UnmarshalKey("resolver", rs)
	if rs.Port == 0 {
		rs.Port = 9091
	}
	if rs.TLS == nil {
		rs.TLS = &tlsSettings{Enabled: false}
	}
	if rs.Client == nil {
		rs.Client = new(ClientSettings)
		_ = rs.Client.Validate()
		rs.Client.Insecure = v.GetBool("resolver.client.insecure")
		if node := v.GetString("resolver.client.node"); node != "" {
			rs.Client.Node = node
		}
	}
}

// Resolver instance.
func (rs *ResolverSettings) Resolver(conn *grpc.ClientConn) (*resolver.Instance, error) {
	// Driver instance
	provider := &provider{client: protoV1.NewAgentAPIClient(conn)}

	// Resolver instance
	return resolver.New(resolver.WithProvider("algo", provider))
}

// ServerOpts returns proper settings when exposing the resolver through
// an HTTP endpoint.
func (rs *ResolverSettings) ServerOpts(handler http.Handler, rc string) []pkgHttp.Option {
	opts := []pkgHttp.Option{
		pkgHttp.WithHandler(handler),
		pkgHttp.WithPort(int(rs.Port)),
		pkgHttp.WithIdleTimeout(10 * time.Second),
		pkgHttp.WithMiddleware(middleware.PanicRecovery()),
		pkgHttp.WithMiddleware(otelHttp.NewMonitor().ServerMiddleware("resolver")),
		pkgHttp.WithMiddleware(middleware.Headers(map[string]string{
			"x-resolver-version":     info.CoreVersion,
			"x-resolver-build-code":  info.BuildCode,
			"x-resolver-release":     rc,
			"x-content-type-options": "nosniff",
		})),
	}
	if rs.ProxyProtocol {
		opts = append(opts, pkgHttp.WithMiddleware(middleware.ProxyHeaders()))
	}
	if rs.TLS.Enabled {
		val, err := rs.TLS.expand()
		if err == nil {
			opts = append(opts, pkgHttp.WithTLS(*val))
		}
	}
	return opts
}

// ClientSettings defines the configuration options available when
// interacting with an AlgoDID network agent.
type ClientSettings struct {
	Node     string `json:"node" yaml:"node" mapstructure:"node"`
	Timeout  uint   `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
	Insecure bool   `json:"insecure" yaml:"insecure" mapstructure:"insecure"`
	Override string `json:"override" yaml:"override" mapstructure:"override"`
	PoW      uint   `json:"pow" yaml:"pow" mapstructure:"pow"`
}

// Validate the settings and load sensible default values.
func (cl *ClientSettings) Validate() error {
	if cl.Node == "" {
		cl.Node = DefaultAgentEndpoint
	}
	if cl.PoW == 0 {
		cl.PoW = 8
	}
	if cl.Timeout == 0 {
		cl.Timeout = 5
	}
	return nil
}

type tlsSettings struct {
	Enabled  bool     `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	SystemCA bool     `json:"system_ca" yaml:"system_ca" mapstructure:"system_ca"`
	Cert     string   `json:"cert" yaml:"cert" mapstructure:"cert"`
	Key      string   `json:"key" yaml:"key" mapstructure:"key"`
	CustomCA []string `json:"custom_ca" yaml:"custom_ca" mapstructure:"custom_ca"`

	// private expanded values
	cert      []byte
	key       []byte
	customCAs [][]byte
}

func (ts *tlsSettings) expand() (*pkgHttp.TLS, error) {
	var err error
	if ts.Cert != "" {
		ts.cert, err = loadPem(ts.Cert)
		if err != nil {
			return nil, err
		}
	}
	if ts.Key != "" {
		ts.key, err = loadPem(ts.Key)
		if err != nil {
			return nil, err
		}
	}
	ts.customCAs = [][]byte{}
	for _, ca := range ts.CustomCA {
		cp, err := loadPem(ca)
		if err != nil {
			return nil, err
		}
		ts.customCAs = append(ts.customCAs, cp)
	}
	return &pkgHttp.TLS{
		Cert:             ts.cert,
		PrivateKey:       ts.key,
		IncludeSystemCAs: ts.SystemCA,
		CustomCAs:        ts.customCAs,
	}, nil
}
