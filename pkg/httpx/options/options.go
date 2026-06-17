package options

import (
	"github.com/Archiit19/School-management/pkg/httpx/config"
	"github.com/Archiit19/School-management/pkg/httpx/constants"
)

// Option customizes client configuration.
type Option func(*config.Config)

func WithBackend(backend constants.Backend) Option {
	return func(c *config.Config) { c.Backend = backend }
}

func WithBaseURL(baseURL string) Option {
	return func(c *config.Config) { c.BaseURL = baseURL }
}

func WithToken(token string) Option {
	return func(c *config.Config) { c.Token = token }
}

func WithName(name string) Option {
	return func(c *config.Config) { c.Name = name }
}

func WithHTTP(httpCfg *config.HTTPClient) Option {
	return func(c *config.Config) { c.HTTP = httpCfg }
}

// Apply returns cfg with all options applied.
func Apply(cfg config.Config, opts ...Option) config.Config {
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
