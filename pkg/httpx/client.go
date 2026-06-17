package httpx

import (
	"net/http"

	"github.com/Archiit19/School-management/pkg/httpx/config"
	"github.com/Archiit19/School-management/pkg/httpx/constants"
	"github.com/Archiit19/School-management/pkg/httpx/internal/adapter"
	"github.com/Archiit19/School-management/pkg/httpx/internal/builder"
	"github.com/Archiit19/School-management/pkg/httpx/internal/client"
	"github.com/Archiit19/School-management/pkg/httpx/options"
	"github.com/Archiit19/School-management/pkg/httpx/providers"
	"github.com/Archiit19/School-management/pkg/httpx/providers/factories"
)

// Client performs outbound service-to-service requests.
type Client = providers.Client

// ClientConfig configures an outbound service client.
type ClientConfig = config.Config

// HTTPClient holds transport, retry, and circuit-breaker settings for outbound calls.
type HTTPClient = config.HTTPClient

// DefaultHTTPClient returns production-oriented defaults for internal service calls.
func DefaultHTTPClient() HTTPClient {
	return config.DefaultHTTPClient()
}

// LoadHTTPClientConfigFromEnv loads HTTP client settings from environment variables.
func LoadHTTPClientConfigFromEnv() HTTPClient {
	return config.LoadHTTPClientConfigFromEnv()
}

// Backend identifies the transport used for outbound service calls.
type Backend = constants.Backend

const (
	BackendHTTP         = constants.BackendHTTP
	InternalTokenHeader = constants.InternalTokenHeader
	DefaultTimeout      = constants.DefaultTimeout
)

// ErrCircuitOpen is returned when the circuit breaker is open for a dependency.
var ErrCircuitOpen = builder.ErrCircuitOpen

// New creates an internal client using the HTTP backend and environment-based defaults.
func New(baseURL, token string) Client {
	return MustNewFactory(BackendHTTP).MustNew(baseURL, token)
}

// NewFromConfig creates a client with tuned transport, retry, and circuit breaker.
func NewFromConfig(cfg config.Config) Client {
	return MustNewFactory(BackendHTTP).MustNewFromConfig(cfg)
}

// NewWithOptions creates a client from functional options.
func NewWithOptions(baseURL, token string, opts ...options.Option) Client {
	cfg := options.Apply(config.Config{BaseURL: baseURL, Token: token}, opts...)
	return NewFromConfig(cfg)
}

// NewWithHTTP allows injecting a custom http.Client (e.g. tests or shared transport).
func NewWithHTTP(baseURL, token string, httpClient *http.Client) Client {
	return client.NewHTTPWithClient(baseURL, token, httpClient)
}

// NewHTTPClient builds a production-ready *http.Client with pooling, retry, and optional circuit breaker.
func NewHTTPClient(name string, cfg HTTPClient) *http.Client {
	return builder.NewHTTPClient(name, cfg)
}

// ReadErrorBody reads response body for error messages.
func ReadErrorBody(resp *http.Response) string {
	return adapter.ReadErrorBody(resp)
}

// CheckStatus returns an error when status does not match expectStatus.
func CheckStatus(resp *http.Response, expectStatus int, context string) error {
	return adapter.CheckStatus(resp, expectStatus, context)
}

// CheckStatusAny returns an error when status is not one of the allowed codes.
func CheckStatusAny(resp *http.Response, context string, allowed ...int) error {
	return adapter.CheckStatusAny(resp, context, allowed...)
}

// MustNewFactory returns a factory for the given backend or panics when unsupported.
func MustNewFactory(backend Backend) *ClientFactory {
	f, err := NewClientFactory(backend)
	if err != nil {
		panic(err)
	}
	return f
}

// NewClientFactory returns a factory for the given backend (e.g. "http").
func NewClientFactory(backend Backend) (*ClientFactory, error) {
	factory, err := factories.New(backend, &HTTPClientProvider{})
	if err != nil {
		return nil, err
	}
	return &ClientFactory{factory: factory}, nil
}
