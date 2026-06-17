package httpx

import (
	"github.com/Archiit19/School-management/pkg/httpx/config"
	"github.com/Archiit19/School-management/pkg/httpx/constants"
	"github.com/Archiit19/School-management/pkg/httpx/providers/factories"
)

// ClientFactory creates outbound clients for a specific backend transport.
type ClientFactory struct {
	factory *factories.Factory
}

// Backend returns the default backend for this factory.
func (f *ClientFactory) Backend() constants.Backend {
	return f.factory.Backend()
}

// NewFromConfig builds a client using the factory backend unless cfg.Backend is set.
func (f *ClientFactory) NewFromConfig(cfg config.Config) (Client, error) {
	return f.factory.NewFromConfig(cfg)
}

// New builds a client with the factory backend and default transport settings.
func (f *ClientFactory) New(baseURL, token string) (Client, error) {
	return f.factory.New(baseURL, token)
}

// MustNewFromConfig is like NewFromConfig but panics on error.
func (f *ClientFactory) MustNewFromConfig(cfg config.Config) Client {
	c, err := f.NewFromConfig(cfg)
	if err != nil {
		panic(err)
	}
	return c
}

// MustNew is like New but panics on error.
func (f *ClientFactory) MustNew(baseURL, token string) Client {
	c, err := f.New(baseURL, token)
	if err != nil {
		panic(err)
	}
	return c
}
