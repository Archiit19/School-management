package factories

import (
	"fmt"

	"github.com/Archiit19/School-management/pkg/httpx/config"
	"github.com/Archiit19/School-management/pkg/httpx/constants"
	"github.com/Archiit19/School-management/pkg/httpx/providers"
)

// Factory creates outbound clients for a registered backend transport.
type Factory struct {
	backend   constants.Backend
	providers map[constants.Backend]providers.Provider
}

// New creates a factory for the given backend.
func New(backend constants.Backend, registry ...providers.Provider) (*Factory, error) {
	if backend == "" {
		backend = constants.BackendHTTP
	}
	providersByBackend := make(map[constants.Backend]providers.Provider, len(registry))
	for _, p := range registry {
		providersByBackend[p.Backend()] = p
	}
	if _, ok := providersByBackend[backend]; !ok {
		return nil, fmt.Errorf("httpx: unsupported backend %q", backend)
	}
	return &Factory{backend: backend, providers: providersByBackend}, nil
}

// Backend returns the default backend for this factory.
func (f *Factory) Backend() constants.Backend {
	return f.backend
}

// NewFromConfig builds a client using the factory backend unless cfg.Backend is set.
func (f *Factory) NewFromConfig(cfg config.Config) (providers.Client, error) {
	backend := cfg.Backend
	if backend == "" {
		backend = f.backend
	}
	provider, ok := f.providers[backend]
	if !ok {
		return nil, fmt.Errorf("httpx: unsupported backend %q", backend)
	}
	return provider.New(cfg)
}

// New builds a client with the factory backend and default transport settings.
func (f *Factory) New(baseURL, token string) (providers.Client, error) {
	return f.NewFromConfig(config.Config{BaseURL: baseURL, Token: token})
}
