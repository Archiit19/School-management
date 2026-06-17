package httpx

import (
	"github.com/Archiit19/School-management/pkg/httpx/config"
	"github.com/Archiit19/School-management/pkg/httpx/constants"
	"github.com/Archiit19/School-management/pkg/httpx/internal/client"
	"github.com/Archiit19/School-management/pkg/httpx/providers"
)

// HTTPClientProvider creates HTTP-backed clients.
type HTTPClientProvider struct{}

// Backend returns the HTTP transport identifier.
func (p *HTTPClientProvider) Backend() constants.Backend {
	return constants.BackendHTTP
}

// New builds an HTTP client from configuration.
func (p *HTTPClientProvider) New(cfg config.Config) (providers.Client, error) {
	return client.NewHTTP(cfg), nil
}
