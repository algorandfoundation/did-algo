package resolver

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"go.bryk.io/pkg/did"
)

// Default resolver HTTP client.
var client *http.Client

func init() {
	client = &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   5 * time.Second,
	}
}

// Provider represents an external system able to return DID
// Documents on demand.
type Provider struct {
	// Method value expected on the identifier instance.
	Method string

	// Network location to retrieve DID documents from. The value can
	// be a template with support for the following variables: DID, Method
	// and Subject. For example:
	// https://did.baidu.com/v1/did/resolve/{{.DID}}
	Endpoint string

	// Protocol used to communicate with the endpoint. Currently, HTTP(S)
	// is supported by submitting GET requests.
	Protocol string

	// Compiled endpoint template
	tpl *template.Template
}

// Get the DID document (or the provider's response) for the
// provided identifier instance.
func Get(id string, providers []*Provider) ([]byte, error) {
	// Validate id
	r, err := did.Parse(id)
	if err != nil {
		return nil, err
	}

	// Select provider
	var p *Provider
	for _, p = range providers {
		if p.Method == r.Method() {
			break
		}
	}
	if p == nil {
		return nil, errors.New("unsupported method")
	}

	// Return result
	return p.resolve(r)
}

func (p *Provider) resolve(id *did.Identifier) ([]byte, error) {
	var err error

	// Parse template
	if p.tpl == nil {
		p.tpl, err = template.New(p.Method).Parse(p.Endpoint)
		if err != nil {
			return nil, err
		}
	}

	// Build URL
	buf := bytes.NewBuffer(nil)
	if err = p.tpl.Execute(buf, p.data(id)); err != nil {
		return nil, err
	}

	// Submit request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, buf.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	// Return response
	return io.ReadAll(res.Body)
}

func (p *Provider) data(id *did.Identifier) map[string]string {
	return map[string]string{
		"DID":     id.String(),
		"Method":  id.Method(),
		"Subject": id.Subject(),
	}
}
