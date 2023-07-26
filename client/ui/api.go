package ui

import (
	"time"

	"github.com/algorandfoundation/did-algo/client/internal"
	"github.com/algorandfoundation/did-algo/client/store"
	xlog "go.bryk.io/pkg/log"
	xhttp "go.bryk.io/pkg/net/http"
	mdCors "go.bryk.io/pkg/net/middleware/cors"
	mdGzip "go.bryk.io/pkg/net/middleware/gzip"
	mdLogging "go.bryk.io/pkg/net/middleware/logging"
	mdRecovery "go.bryk.io/pkg/net/middleware/recovery"
)

// LocalAPI makes a local provider instance accessible through
// an HTTP server.
type LocalAPI struct {
	prv *Provider
	log xlog.Logger
	srv *xhttp.Server
}

// LocalAPIServer creates a new instance of the local API server.
func LocalAPIServer(st *store.LocalStore, conf *internal.ClientSettings, log xlog.Logger) (*LocalAPI, error) {
	// provider instances
	p := &Provider{
		st:   st,
		log:  log,
		conf: conf,
	}
	if err := p.connect(); err != nil {
		return nil, err
	}

	// server instance
	opts := []xhttp.Option{
		xhttp.WithPort(9090),
		xhttp.WithHandler(p.ServerHandler()),
		xhttp.WithIdleTimeout(10 * time.Second),
		xhttp.WithMiddleware(
			mdLogging.Handler(log, nil),
			mdCors.Handler(mdCors.Options{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "OPTIONS"},
				AllowedHeaders: []string{"content-type"},
			}),
			mdGzip.Handler(5),
			mdRecovery.Handler(),
		),
	}
	srv, err := xhttp.NewServer(opts...)
	if err != nil {
		_ = p.close()
		return nil, err
	}
	return &LocalAPI{
		prv: p,
		log: log,
		srv: srv,
	}, nil
}

// Start the local API server.
func (el *LocalAPI) Start() error {
	return el.srv.Start()
}

// Stop the local API server.
func (el *LocalAPI) Stop() error {
	if err := el.prv.close(); err != nil {
		el.log.Warning(err.Error())
	}
	return el.srv.Stop(true)
}
